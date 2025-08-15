package core

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
)

// JSONFileProvider implements DataProvider interface using JSON files for persistence.
type JSONFileProvider struct {
	dbPath string
	cache  *inMemoryCache
	mu     sync.RWMutex
}

// inMemoryCache serves as in-memory data cache to improve performance.
type inMemoryCache struct {
	Codebases      map[string]*Codebase             // codebase_id -> Codebase
	Versions       map[string]*Version              // version_id -> Version
	FileIndexes    map[string][]File                // tree_id -> []File
	VersionMapping map[string]*versionMappingRecord // child_version_id -> mapping

	// Indexes for fast lookup
	versionsByCodebase       map[string][]*Version // codebase_id -> sorted []*Version by time
	versionIDByBranchAndName map[string]string     // key: "codebaseID/branch/version" -> versionID
}

// versionMappingRecord is the record structure stored in version_mapping.json.
type versionMappingRecord struct {
	ID              string      `json:"id"`
	CodebaseID      string      `json:"codebase_id"`
	Branch          string      `json:"branch"`
	ChildVersionID  string      `json:"child_version_id"`
	ParentVersionID string      `json:"parent_version_id"`
	LinkageType     LinkageType `json:"linkage_type"`
}

func NewJSONFileProvider(dbPath string) (*JSONFileProvider, error) {
	if err := os.MkdirAll(dbPath, 0755); err != nil {
		return nil, fmt.Errorf("unable to create database directory: %w", err)
	}
	p := &JSONFileProvider{
		dbPath: dbPath,
		cache: &inMemoryCache{
			Codebases:                make(map[string]*Codebase),
			Versions:                 make(map[string]*Version),
			FileIndexes:              make(map[string][]File),
			VersionMapping:           make(map[string]*versionMappingRecord),
			versionsByCodebase:       make(map[string][]*Version),
			versionIDByBranchAndName: make(map[string]string),
		},
	}
	if err := p.load(); err != nil {
		return nil, fmt.Errorf("failed to load data: %w", err)
	}
	p.rebuildIndexes()
	return p, nil
}

// --- Data Loading and Saving ---

func (p *JSONFileProvider) load() error {
	if err := p.loadJSON("codebases.json", &p.cache.Codebases); err != nil {
		return err
	}
	if err := p.loadJSON("versions.json", &p.cache.Versions); err != nil {
		return err
	}
	if err := p.loadJSON("file_indexes.json", &p.cache.FileIndexes); err != nil {
		return err
	}
	if err := p.loadJSON("version_mapping.json", &p.cache.VersionMapping); err != nil {
		return err
	}
	return nil
}

func (p *JSONFileProvider) save(filename string, data interface{}) error {
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filepath.Join(p.dbPath, filename), bytes, 0644)
}

func (p *JSONFileProvider) loadJSON(filename string, target interface{}) error {
	path := filepath.Join(p.dbPath, filename)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil // File not existing is normal situation
	}
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	if len(bytes) == 0 {
		return nil
	}
	return json.Unmarshal(bytes, target)
}

func (p *JSONFileProvider) rebuildIndexes() {
	for _, v := range p.cache.Versions {
		p.cache.versionsByCodebase[v.CodebaseID] = append(p.cache.versionsByCodebase[v.CodebaseID], v)
		key := fmt.Sprintf("%s/%s/%s", v.CodebaseID, v.Branch, v.Version)
		p.cache.versionIDByBranchAndName[key] = v.ID
	}
	for cid := range p.cache.versionsByCodebase {
		sort.Slice(p.cache.versionsByCodebase[cid], func(i, j int) bool {
			return p.cache.versionsByCodebase[cid][i].CreatedAt.After(p.cache.versionsByCodebase[cid][j].CreatedAt)
		})
	}
}

// --- Interface Implementations ---

func (p *JSONFileProvider) CreateCodebase(codebase *Codebase) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, exists := p.cache.Codebases[codebase.ID]; exists {
		return fmt.Errorf("codebase %s already exists", codebase.ID)
	}
	p.cache.Codebases[codebase.ID] = codebase
	return p.save("codebases.json", p.cache.Codebases)
}

func (p *JSONFileProvider) GetCodebaseByID(id string) (*Codebase, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	codebase, ok := p.cache.Codebases[id]
	if !ok {
		return nil, fmt.Errorf("codebase %s not found", id)
	}
	return codebase, nil
}

func (p *JSONFileProvider) DeleteCodebaseByID(id string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Delete all associated content
	delete(p.cache.Codebases, id)
	relatedVersions := p.cache.versionsByCodebase[id]
	for _, v := range relatedVersions {
		delete(p.cache.Versions, v.ID)
		delete(p.cache.FileIndexes, v.TreeID)
		delete(p.cache.VersionMapping, v.ID)
		delete(p.cache.versionIDByBranchAndName, fmt.Sprintf("%s/%s/%s", v.CodebaseID, v.Branch, v.Version))
	}
	delete(p.cache.versionsByCodebase, id)

	// Save all changes
	if err := p.save("codebases.json", p.cache.Codebases); err != nil {
		return err
	}
	if err := p.save("versions.json", p.cache.Versions); err != nil {
		return err
	}
	if err := p.save("file_indexes.json", p.cache.FileIndexes); err != nil {
		return err
	}
	if err := p.save("version_mapping.json", p.cache.VersionMapping); err != nil {
		return err
	}

	// Delete history cache file
	err := os.Remove(filepath.Join(p.dbPath, "history_cache", id+".json"))
	if err != nil && !os.IsNotExist(err) {
		return err // Only return non-"file not found" errors
	}
	return nil
}

func (p *JSONFileProvider) UpdateCodebaseTimestamp(id string, t time.Time) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	codebase, ok := p.cache.Codebases[id]
	if !ok {
		return fmt.Errorf("codebase %s not found", id)
	}
	codebase.UpdatedAt = t
	return p.save("codebases.json", p.cache.Codebases)
}

func (p *JSONFileProvider) CreateVersion(version *Version, files []File) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, exists := p.cache.Versions[version.ID]; exists {
		return fmt.Errorf("version %s already exists", version.ID)
	}

	p.cache.Versions[version.ID] = version
	p.cache.FileIndexes[version.TreeID] = files

	// Update indexes
	p.cache.versionsByCodebase[version.CodebaseID] = append(p.cache.versionsByCodebase[version.CodebaseID], version)
	sort.Slice(p.cache.versionsByCodebase[version.CodebaseID], func(i, j int) bool {
		return p.cache.versionsByCodebase[version.CodebaseID][i].CreatedAt.After(p.cache.versionsByCodebase[version.CodebaseID][j].CreatedAt)
	})
	key := fmt.Sprintf("%s/%s/%s", version.CodebaseID, version.Branch, version.Version)
	p.cache.versionIDByBranchAndName[key] = version.ID

	if err := p.save("versions.json", p.cache.Versions); err != nil {
		return err
	}
	return p.save("file_indexes.json", p.cache.FileIndexes)
}

func (p *JSONFileProvider) GetVersion(codebaseID, branch, version string) (*Version, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	key := fmt.Sprintf("%s/%s/%s", codebaseID, branch, version)
	versionID, ok := p.cache.versionIDByBranchAndName[key]
	if !ok {
		return nil, fmt.Errorf("version %s/%s not found", branch, version)
	}
	v, ok := p.cache.Versions[versionID]
	if !ok {
		return nil, fmt.Errorf("data inconsistency: version ID %s not found", versionID)
	}
	return v, nil
}

func (p *JSONFileProvider) GetFileIndexesByTreeID(treeID string) ([]File, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	files, ok := p.cache.FileIndexes[treeID]
	if !ok {
		return nil, fmt.Errorf("tree %s not found", treeID)
	}
	return files, nil
}

func (p *JSONFileProvider) FindLatestVersionInBranch(codebaseID, branch, excludeVersionID string) (*Version, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	versions, ok := p.cache.versionsByCodebase[codebaseID]
	if !ok {
		return nil, nil // No versions
	}
	for _, v := range versions {
		if v.Branch == branch && v.ID != excludeVersionID {
			return v, nil
		}
	}
	return nil, nil
}

func (p *JSONFileProvider) IsNewBranch(codebaseID, branch, excludeVersionID string) (bool, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	versions, ok := p.cache.versionsByCodebase[codebaseID]
	if !ok {
		return true, nil
	}
	count := 0
	for _, v := range versions {
		if v.Branch == branch && v.ID != excludeVersionID {
			count++
		}
	}
	return count == 0, nil
}

func (p *JSONFileProvider) FindLatestVersionInMain(codebaseID string) (*Version, error) {
	return p.FindLatestVersionInBranch(codebaseID, "main", "")
}

func (p *JSONFileProvider) CreateVersionLink(codebaseID, childID, parentID, branch string, linkType LinkageType) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, exists := p.cache.VersionMapping[childID]; exists {
		return nil // Link already exists
	}
	p.cache.VersionMapping[childID] = &versionMappingRecord{
		ID:              uuid.NewString(),
		CodebaseID:      codebaseID,
		Branch:          branch,
		ChildVersionID:  childID,
		ParentVersionID: parentID,
		LinkageType:     linkType,
	}
	return p.save("version_mapping.json", p.cache.VersionMapping)
}

func (p *JSONFileProvider) GetAllVersionsForMap(codebaseID string) ([]VersionNode, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	var nodes []VersionNode
	versions, ok := p.cache.versionsByCodebase[codebaseID]
	if !ok {
		return nodes, nil
	}
	for _, v := range versions {
		nodes = append(nodes, VersionNode{
			ID:        v.ID,
			Version:   v.Version,
			Branch:    v.Branch,
			Message:   v.Message,
			CreatedAt: v.CreatedAt,
			Stats:     v.Stats,
		})
	}
	return nodes, nil
}

func (p *JSONFileProvider) GetAllVersionEdgesForMap(codebaseID string) ([]VersionEdge, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	var edges []VersionEdge
	for _, m := range p.cache.VersionMapping {
		if m.CodebaseID == codebaseID {
			edges = append(edges, VersionEdge{
				From:        m.ParentVersionID,
				To:          m.ChildVersionID,
				LinkageType: m.LinkageType,
			})
		}
	}
	return edges, nil
}

func (p *JSONFileProvider) GetBranchHeadsForMap(codebaseID string) (map[string]string, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	heads := make(map[string]string)
	versions, ok := p.cache.versionsByCodebase[codebaseID]
	if !ok {
		return heads, nil
	}
	for _, v := range versions {
		if _, ok := heads[v.Branch]; !ok {
			heads[v.Branch] = v.ID
		}
	}
	return heads, nil
}

func (p *JSONFileProvider) GetHistoryCache(codebaseID string) ([]byte, error) {
	path := filepath.Join(p.dbPath, "history_cache", codebaseID+".json")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("cache not found")
	}
	return ioutil.ReadFile(path)
}

func (p *JSONFileProvider) UpdateHistoryCache(codebaseID string, data []byte) error {
	dir := filepath.Join(p.dbPath, "history_cache")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	path := filepath.Join(dir, codebaseID+".json")
	return ioutil.WriteFile(path, data, 0644)
}

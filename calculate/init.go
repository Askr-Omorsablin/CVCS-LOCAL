package calculate

import (
	"fmt"
	"log"
	"main/core"
	"time"

	"github.com/google/uuid"
)

type InitService struct{}

func NewInitService() *InitService {
	return &InitService{}
}

// InitializeCodebase creates new codebase record
func (s *InitService) InitializeCodebase(name, description, branch string) (*core.Codebase, error) {
	log.Printf("Starting codebase initialization: name=%s, branch=%s", name, branch)

	codebase := &core.Codebase{
		ID:          uuid.NewString(),
		Name:        name,
		Description: description,
		Branch:      branch,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := core.GetProvider().CreateCodebase(codebase); err != nil {
		log.Printf("Data provider write failed: %v", err)
		return nil, fmt.Errorf("data storage error: %w", err)
	}

	log.Printf("Codebase created successfully: ID=%s", codebase.ID)
	return codebase, nil
}

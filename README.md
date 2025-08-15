
# CVCS: A Local-First, API-Driven Version Control Snapshot System
## Local Version of CVCS

CVCS is a lightweight version control system designed for AI Coding scenarios. It is not intended to replace Git, but rather to complement it in certain areas, particularly excelling in **holistic snapshot-based archiving** of projects, especially those containing large amounts of binary files (such as design drafts, media assets, model data).

---

### Core Philosophy

- **Snapshots over Commits**: The core of CVCS is "snapshots", not Git's "commits". It captures and stores the complete state of a project at specific points in time, without tracking fine-grained file changes (diffs).
- **Local File System Architecture**: All data (file content and metadata) is stored in the local file system. File content and metadata (JSON files) are saved in a configurable root directory, making it completely self-contained and easy to migrate.
- **API-Driven**: All operations are completed through unified, standardized RESTful APIs, making it easy to integrate into automation workflows, CI/CD, or other upper-level applications.
- **Automatic Lineage Tracking**: The system automatically establishes lineage relationships between versions, forming clear version evolution graphs without manual maintenance of version links.
- **Focus on Archiving and Rollback**: The system's main goal is to safely archive the complete state of a project version and conveniently roll back, download entire versions, or individual files within them.

---

### Automatic Version Lineage Management

CVCS employs an intelligent version lineage management mechanism that automatically maintains parent-child relationships between versions:

#### Same-Branch Linear Lineage
- **Time Series**: Versions within the same branch automatically establish parent-child relationships based on creation time
- **No Version Number Dependency**: Does not rely on semantic meaning of version numbers, completely based on timestamp ordering
- **Automatic Linking**: New versions automatically become child versions of the latest version in that branch

#### Cross-Branch Lineage Relationships
- **Branch Creation**: Specify the origin version of a new branch through the `branch_from` parameter
- **Fork Point Recording**: Clearly record the creation starting point of each branch
- **Independent Evolution**: Each branch independently maintains its own time-series lineage relationships

#### Lineage Relationship Types
- `sequential`: Time-series relationships within the same branch
- `branch_from`: Fork relationships across branches

---

### Comparison with Git

CVCS and Git have fundamental differences in design philosophy, applicable scenarios, and functionality. The following table clearly shows their differences:

| Feature                    | CVCS (This Project)                                    | Git (and Mainstream VCS)                                    |
| :------------------------- | :----------------------------------------------------- | :----------------------------------------------------------- |
| **Basic Unit**             | **Snapshot** - Complete state of project at a point in time | **Commit** - A collection of file changes (diff/patch)      |
| **Core Model**             | **Local File System** - Data uniformly stored in local specified directory | **Distributed** - Each clone contains complete historical records |
| **Design Goal**            | **Project Archiving and Rollback** - Quickly save and restore complete state of project versions | **Collaborative Development and Change Tracking** - Manage fine-grained code evolution history |
| **Branching and Merging**  | **Supports Branch Concept** (through automatic lineage tracking and `branch_from` parameter), **No Merge Support** | Core functionality, providing powerful branch management, merge, rebase and other collaborative tools |
| **Version Lineage**        | **Automatic Management** (automatically establish lineage relationships based on time series and branch creation) | **Manual Management** (requires explicit merge, rebase operations to establish relationships) |
| **Storage Efficiency (Source Code)** | Lower (stores complete files, only general compression) | **Very High** (uses packfile and delta compression, stores only differences) |
| **Storage Efficiency (Binary)** | **Higher** (direct storage or compressed storage, similar to Git LFS) | Lower (native Git not good at handling large binary files, requires Git LFS extension) |
| **Query Capability**       | **Medium** (metadata as JSON files, queried by application in memory) | Weaker (relies on Git commands for log and history search) |
| **Applicable Scenarios**   | Design projects, game assets, ML models, scenarios requiring regular full backups | Software source code, documentation, configuration files and other text-based projects |
| **Client**                 | **Any HTTP Client** (`curl`, `requests`, etc.)        | **Dedicated Git Client**                                    |

## API Overview (Unified Request Body Specification)

- All POST requests use the following structure:
```json
{
  "positions": { ... },  // Position information: locate resource context (such as codebase_id)
  "content": { ... }      // Actual business content (such as path, version, branch, configuration, etc.)
}
```

### Endpoint List
- Initialize codebase
  - POST `/api/v1/codebases/init`
- Create snapshot
  - POST `/api/v1/codebases/snapshots/create`
- Download complete repository archive for specified version
  - POST `/api/v1/codebases/archive/get`
- Download single file
  - POST `/api/v1/codebases/file/get`
- Delete codebase
  - POST `/api/v1/codebases/delete`
- Get codebase version history graph
  - POST `/api/v1/codebases/map/get`
- Manually create parent-child link between two versions (advanced)
  - POST `/api/v1/codebases/map/link`
- **(New)** Configure data storage path
  - POST `/api/v1/config/storage/update`

## Unified Request Body Examples

### 1) Initialize Codebase
Request
```bash
curl -X POST http://localhost:8080/api/v1/codebases/init \
  -H "Content-Type: application/json" \
  -d '{
    "positions": {},
    "content": {
      "name": "my-project",
      "description": "Example project",
      "branch": "main"
    }
  }'
```
Response
```json
{
  "id": "e282be9d-1c19-47d3-8903-f0d152aa6eb6",
  "name": "my-project",
  "description": "Example project",
  "branch": "main",
  "created_at": "2025-08-12T19:48:43Z",
  "updated_at": "2025-08-12T19:48:43Z"
}
```

### 2) Create Snapshot
Request
```bash
# Note: This is a multipart/form-data request, cannot use -d like regular json
# Need to use -F option to specify metadata and files
# Assume main.go and README.md files exist in current directory

METADATA='{
  "positions": {
    "codebase_id": "e282be9d-1c19-47d3-8903-f0d152aa6eb6"
  },
  "content": {
    "version": "v1.0.1",
    "branch": "main",
    "message": "Refactor to support file uploads",
    "auto_linkage": true
  }
}'

curl -X POST http://localhost:8080/api/v1/codebases/snapshots/create \
  -H "Content-Type: multipart/form-data" \
  -F "metadata=${METADATA}" \
  -F "main.go=@./main.go" \
  -F "README.md=@./README.md"
```

#### Example of Creating New Branch
```bash
# Create feature-x branch from main/v1.0.0
METADATA='{
  "positions": {
    "codebase_id": "e282be9d-1c19-47d3-8903-f0d152aa6eb6"
  },
  "content": {
    "version": "v1.1-alpha",
    "branch": "feature-x",
    "message": "Start feature X development",
    "branch_from": {
      "branch": "main",
      "version": "v1.0.0"
    },
    "auto_linkage": true
  }
}'

curl -X POST http://localhost:8080/api/v1/codebases/snapshots/create \
  -H "Content-Type: multipart/form-data" \
  -F "metadata=${METADATA}" \
  -F "main.go=@./main.go" \
  -F "README.md=@./README.md"
```

Description
- **Request Format**: This interface accepts `multipart/form-data`. Clients need to use the `-F` option to pass a JSON string named `metadata` and file streams. The field name of each file stream is its relative path in the codebase.
- **Metadata (`metadata`)**:
  - `positions.codebase_id`: (Required) Codebase ID.
  - `content.branch`: (Optional, defaults to "main") Branch to which the snapshot belongs.
  - `content.version`: (Optional, defaults to "v1") Version number of the snapshot. It's recommended to always specify a meaningful version.
  - `content.message`: (Optional) Version description information.
  - `content.branch_from`: (Optional) Used to create new branch, specify its source. Contains `branch` and `version` fields.
  - `content.auto_linkage`: (Optional, defaults to true) Whether to automatically establish lineage relationships.
- **File Processing**:
  - Image files (`.jpg`, `.jpeg`, `.png`, `.gif`, `.webp`, `.bmp`, `.tiff`) will be directly saved.
  - All other files will be zlib compressed before saving.
- **Automatic Lineage Relationship Establishment**:
  - **Same-branch Linear Lineage**: If `branch_from` is not provided, the system will automatically link the new snapshot to the most recent version in the same branch, forming time-series-based linear lineage relationships.
  - **Cross-branch Lineage**: If `branch_from` is provided, the system will automatically establish lineage relationships between the new version and the specified source version, marking it as a branch creation point.
- **Response**: Returns detailed information about `codebase`, `version`, and `file_tree`. To ensure real-time client state synchronization, the response body will also include the complete updated version graph `version_map`.

### 3) Download Complete Repository Archive
Request
```bash
curl -X POST http://localhost:8080/api/v1/codebases/archive/get \
  -H "Content-Type: application/json" \
  -d '{
    "positions": { "codebase_id": "e282be9d-1c19-47d3-8903-f0d152aa6eb6" },
    "content":   { "branch": "main", "version": "v1.0.1" }
  }' \
  --output my-project-v1.0.1.zip
```

### 4) Download Single File
Request
```bash
curl -X POST http://localhost:8080/api/v1/codebases/file/get \
  -H "Content-Type: application/json" \
  -d '{
    "positions": { "codebase_id": "e282be9d-1c19-47d3-8903-f0d152aa6eb6" },
    "content":   { "branch": "main", "version": "v1.0.1", "path": "res.py" }
  }' \
  --output downloaded_res.py
```

### 5) Delete Codebase
Request
```bash
curl -X POST http://localhost:8080/api/v1/codebases/delete \
  -H "Content-Type: application/json" \
  -d '{
    "positions": { "codebase_id": "e282be9d-1c19-47d3-8903-f0d152aa6eb6" }
  }'
```
Description
- **This is a very dangerous operation that will permanently delete data.**
- The system will delete the storage directory and all metadata files of the specified codebase.

### 6) Get Version History Graph
Request
```bash
curl -X POST http://localhost:8080/api/v1/codebases/map/get \
  -H "Content-Type: application/json" \
  -d '{
    "positions": { "codebase_id": "e282be9d-1c19-47d3-8903-f0d152aa6eb6" }
  }'
```

### 7) Create Version Link
Request
```bash
curl -X POST http://localhost:8080/api/v1/codebases/map/link \
  -H "Content-Type: application/json" \
  -d '{
    "positions": { "codebase_id": "e282be9d-1c19-47d3-8903-f0d152aa6eb6" },
    "content": {
      "child_version": { "branch": "main", "version": "v1.0.1" },
      "parent_version": { "branch": "main", "version": "v1.0.0" }
    }
  }'
```

### 8) Configure Storage Path
Request
```bash
curl -X POST http://localhost:8080/api/v1/config/storage/update \
  -H "Content-Type: application/json" \
  -d '{
    "content": {
      "path": "./my_new_storage_location"
    }
  }'
```
Description
- This endpoint will update the configuration in the user config directory and set the `STORAGE_PATH` variable.
- **Configuration takes effect immediately** without requiring service restart.

Success Response
```json
{
  "message": "Storage path updated successfully. Configuration is now in effect.",
  "path": "./my_new_storage_location"
}
```

## File Processing and Storage

### Data Directory Structure
All data is stored by default in the `cvcs_data` folder under the program's running directory, with the following structure:
```
./cvcs_data/
├── db/                   # Store metadata JSON files
│   ├── codebases.json
│   ├── file_indexes.json
│   ├── version_mapping.json
│   └── versions.json
└── oss/                  # Store actual file content (simulating OSS)
    └── {codebase_name}/
        ├── ...
        └── {file_hash}.zlib
```
You can change this root directory through the configuration API or by directly modifying the config file in the user directory.

### File Processing Rules
- **Image Files**: `.jpg`, `.jpeg`, `.png`, `.gif`, `.webp`, `.bmp`, `.tiff` - Stored directly without compression.
- **Other Files**: All non-image files - Stored after zlib compression.
- File paths and naming maintain their original relative structure.

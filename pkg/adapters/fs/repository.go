package fs

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/aretw0/loam/pkg/core"
	"github.com/aretw0/loam/pkg/git"
	"gopkg.in/yaml.v3"
)

// Repository implements core.Repository using the filesystem and Git.
type Repository struct {
	Path   string
	git    *git.Client
	cache  *cache
	config Config
}

// Config holds the configuration for the filesystem repository.
type Config struct {
	Path      string
	AutoInit  bool
	Gitless   bool
	MustExist bool
	Logger    *slog.Logger
	SystemDir string            // e.g. ".loam"
	IDMap     map[string]string // Map filename -> ID column name (e.g. "users.csv": "email"). User must ensure uniqueness of values in this column.
}

// NewRepository creates a new filesystem-backed repository.
func NewRepository(config Config) *Repository {
	return &Repository{
		Path:   config.Path,
		git:    git.NewClient(config.Path, config.SystemDir+".lock", config.Logger),
		config: config,
		cache:  newCache(config.Path, config.SystemDir),
	}
}

// Begin starts a new transaction.
func (r *Repository) Begin(ctx context.Context) (core.Transaction, error) {
	return NewTransaction(r), nil
}

// Initialize performs the necessary setup for the repository (mkdir, git init).
func (r *Repository) Initialize(ctx context.Context) error {
	// 1. Directory Initialization
	if r.config.MustExist {
		info, err := os.Stat(r.Path)
		if os.IsNotExist(err) {
			return fmt.Errorf("vault path does not exist: %s", r.Path)
		}
		if !info.IsDir() {
			return fmt.Errorf("vault path is not a directory: %s", r.Path)
		}
	} else {
		if err := os.MkdirAll(r.Path, 0755); err != nil {
			return fmt.Errorf("failed to create vault directory: %w", err)
		}
	}

	// 2. Git Initialization
	if !r.config.Gitless {
		if !git.IsInstalled() {
			return fmt.Errorf("git is not installed")
		}

		if !r.git.IsRepo() {
			if r.config.AutoInit {
				if err := r.git.Init(); err != nil {
					return fmt.Errorf("failed to git init: %w", err)
				}
			} else {
				return fmt.Errorf("path is not a git repository: %s", r.Path)
			}
		}
	}

	return nil
}

// Sync synchronizes the repository with its remote.
func (r *Repository) Sync(ctx context.Context) error {
	if r.config.Gitless {
		return fmt.Errorf("cannot sync in gitless mode")
	}

	if !r.git.IsRepo() {
		return fmt.Errorf("path is not a git repository: %s", r.Path)
	}

	unlock, err := r.git.Lock()
	if err != nil {
		return fmt.Errorf("failed to acquire git lock: %w", err)
	}
	defer unlock()

	return r.git.Sync() // This method handles pull/push
}

// Save persists a document to the filesystem and commits it to Git.
// If the document belongs to a collection (e.g. CSV), it updates the specific row.
func (r *Repository) Save(ctx context.Context, doc core.Document) error {
	if doc.ID == "" {
		return fmt.Errorf("document has no ID")
	}

	ext := filepath.Ext(doc.ID)
	// Smart Extension Detection
	if ext == "" {
		if val, ok := doc.Metadata["ext"].(string); ok && val != "" {
			if strings.HasPrefix(val, ".") {
				ext = val
			} else {
				ext = "." + val
			}
		} else {
			ext = ".md" // Default
		}
	}

	// Construct filename.
	filename := doc.ID
	if filepath.Ext(doc.ID) != ext {
		filename = doc.ID + ext
	}

	fullPath := filepath.Join(r.Path, filename)

	// Ensure parent directory exists
	// But first, check if we should intercept for Multi-Doc (Collection)
	if collectionPath, colExt, key, found := r.findCollection(doc.ID); found {
		return r.saveToCollection(ctx, doc, collectionPath, colExt, key)
	}

	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	data, err := serialize(doc, ext)
	if err != nil {
		return fmt.Errorf("failed to serialize document: %w", err)
	}

	if err := writeFileAtomic(fullPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	if !r.config.Gitless {
		unlock, err := r.git.Lock()
		if err != nil {
			return fmt.Errorf("failed to acquire git lock: %w", err)
		}
		defer unlock()

		if err := r.git.Add(filename); err != nil {
			return fmt.Errorf("failed to git add: %w", err)
		}

		msg := "update " + doc.ID
		if val, ok := ctx.Value(core.ChangeReasonKey).(string); ok && val != "" {
			msg = val
		}

		if err := r.git.Commit(msg); err != nil {
			return fmt.Errorf("failed to git commit: %w", err)
		}
	}

	return nil
}

// Get retrieves a document from the filesystem.
func (r *Repository) Get(ctx context.Context, id string) (core.Document, error) {
	// Attempt to find the file.
	// 1. Try exact match (if ID has extension)
	// 2. Try default .md

	filename := id
	ext := filepath.Ext(id)

	if ext == "" {
		ext = ".md"
		filename = id + ext
	}

	fullPath := filepath.Join(r.Path, filename)

	f, err := os.Open(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Fallback: Check if it's a sub-document inside a collection (e.g. CSV)
			if doc, err2 := r.getFromCollection(ctx, id); err2 == nil {
				return doc, nil
			}
			// Return original error if fallback fails
		}
		return core.Document{}, err
	}
	defer f.Close()

	doc, err := parse(f, ext)
	if err != nil {
		return core.Document{}, fmt.Errorf("failed to parse document %s: %w", id, err)
	}
	doc.ID = id

	return *doc, nil
}

func (r *Repository) findCollection(id string) (collectionPath, collectionExt, key string, found bool) {
	dir := filepath.Dir(id)
	key = filepath.Base(id)
	dir = filepath.ToSlash(dir)

	candidates := []string{dir}
	extensions := []string{".csv", ".json"}
	for _, ext := range extensions {
		candidates = append(candidates, dir+ext)
	}

	for _, c := range candidates {
		path := filepath.Join(r.Path, c)
		info, err := os.Stat(path)
		if err == nil && !info.IsDir() {
			return path, filepath.Ext(path), key, true
		}
	}
	return "", "", "", false
}

func (r *Repository) getFromCollection(ctx context.Context, id string) (core.Document, error) {
	collectionPath, collectionExt, key, found := r.findCollection(id)
	if !found {
		return core.Document{}, fmt.Errorf("collection not found")
	}

	// Read Collection
	data, err := os.ReadFile(collectionPath)
	if err != nil {
		return core.Document{}, err
	}

	if collectionExt == ".csv" {
		reader := csv.NewReader(bytes.NewReader(data))
		headers, err := reader.Read()
		if err != nil {
			return core.Document{}, err
		}

		// Determine ID column
		idColName := r.getIDColumn(filepath.Base(collectionPath))
		idCol := -1
		for i, h := range headers {
			if strings.EqualFold(h, idColName) {
				idCol = i
				break
			}
		}
		if idCol == -1 {
			return core.Document{}, fmt.Errorf("csv collection missing '%s' column", idColName)
		}

		// Scan rows
		for {
			row, err := reader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				return core.Document{}, err
			}

			if len(row) > idCol && row[idCol] == key {
				// Match!
				doc := core.Document{
					ID:       id,
					Metadata: make(core.Metadata),
				}

				for i, h := range headers {
					val := row[i]
					if strings.ToLower(h) == "content" {
						doc.Content = val
					} else {
						doc.Metadata[h] = val
					}
				}
				return doc, nil
			}
		}
	}

	return core.Document{}, fmt.Errorf("document not found in collection")
}

func (r *Repository) saveToCollection(ctx context.Context, doc core.Document, collectionPath, collectionExt, key string) error {
	// Read-Modify-Write
	// Lock? Ideally yes. atomic.go helps with write, but race condition on read-mod possible.
	// For now, relies on atomic.go file swap.

	data, err := os.ReadFile(collectionPath)
	if err != nil {
		return err
	}

	if collectionExt == ".csv" {
		reader := csv.NewReader(bytes.NewReader(data))
		allRecords, err := reader.ReadAll()
		if err != nil {
			return err
		}

		if len(allRecords) == 0 {
			return fmt.Errorf("empty csv collection")
		}

		headers := allRecords[0]
		idColName := r.getIDColumn(filepath.Base(collectionPath))
		idCol := -1
		for i, h := range headers {
			if strings.EqualFold(h, idColName) {
				idCol = i
				break
			}
		}
		if idCol == -1 {
			return fmt.Errorf("csv collection missing '%s' column", idColName)
		}

		foundIndex := -1
		for i := 1; i < len(allRecords); i++ {
			row := allRecords[i]
			if len(row) > idCol && row[idCol] == key {
				foundIndex = i
				break
			}
		}

		// Prepare row data
		newRow := make([]string, len(headers))
		// Pre-fill with existing data if found?
		// Or doc overwrites entirely?
		// Repository.Save usually means "replace".
		// But for a sub-document, we probably only have the fields provided in Metadata?
		// If I provide partial metadata, do I lose other columns?
		// Standard Loam Save replaces the document.
		// So we should probably preserve ID and fill others from Doc.

		// Fill ID
		newRow[idCol] = key

		// Fill from Doc
		for i, h := range headers {
			if i == idCol {
				continue
			}
			if strings.ToLower(h) == "content" {
				newRow[i] = doc.Content
				continue
			}
			if val, ok := doc.Metadata[h]; ok {
				newRow[i] = fmt.Sprintf("%v", val)
			} else {
				// If not in metadata...
				// Logic A: Clear it (Replace semantics).
				// Logic B: Keep existing (Patch semantics).
				// Loam Save is Replace. But strictly, if I Get() -> Modify -> Save(), I have all fields.
				// If I construct new Doc -> Save(), I expect only my fields.
				// For CSV, "missing" usually means empty string.
				newRow[i] = ""

				// Optional: Copy existing if found?
				// if foundIndex != -1 && len(allRecords[foundIndex]) > i {
				// 	newRow[i] = allRecords[foundIndex][i]
				// }
				// Let's stick to Replace (Empty if missing) for now to be consistent.
			}
		}

		if foundIndex != -1 {
			// Update
			allRecords[foundIndex] = newRow
		} else {
			// Append
			allRecords = append(allRecords, newRow)
		}

		// Serialize back
		var buf bytes.Buffer
		w := csv.NewWriter(&buf)
		if err := w.WriteAll(allRecords); err != nil {
			return err
		}
		w.Flush()

		// Atomic Write
		return writeFileAtomic(collectionPath, buf.Bytes(), 0644)
	}

	return fmt.Errorf("unsupported collection type for save")
}

// List scans the directory for all documents.
func (r *Repository) List(ctx context.Context) ([]core.Document, error) {
	var docs []core.Document

	// Load Cache Logic
	if err := r.cache.Load(); err != nil {
		// Log? We don't have logger here yet.
		// Ignore loading error for now, cache will start empty.
	}
	seen := make(map[string]bool)

	err := filepath.WalkDir(r.Path, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			// Skip system directories
			if d.Name() == ".git" || d.Name() == r.config.SystemDir {
				return filepath.SkipDir
			}
			return nil
		}

		ext := filepath.Ext(d.Name())
		// Filter supported extensions
		switch ext {
		case ".md", ".json", ".yaml", ".yml", ".csv":
			// OK
		default:
			return nil
		}

		relPath, err := filepath.Rel(r.Path, path)
		if err != nil {
			return err
		}
		relPath = filepath.ToSlash(relPath)

		// Check if it's a collection and flatten it
		if colDocs, err := r.flattenCollection(ctx, path, relPath); err == nil {
			// We don't verify cache for sub-docs yet (TODO)
			// Directly append for prototype
			docs = append(docs, colDocs...)
			// If it was a collection, do we still return the file itself?
			// Maybe yes, maybe no. For now, let's skip the file if it was successfully flattened?
			// Or keep both. Keep both is safer.
		}

		// ID Strategy:
		// If .md, strip extension (legacy behavior).
		// If others, keep extension?
		// For consistency with Get logic:
		// If I List(), I want IDs that I can pass to Get().
		// If I pass "foo.json" to Get(), it works.
		// If I pass "foo" (for foo.md) to Get(), it works.
		id := relPath
		if ext == ".md" {
			id = relPath[0 : len(relPath)-3]
		}

		// Get file info for mtime
		info, err := d.Info()
		if err != nil {
			return nil
		}
		mtime := info.ModTime()

		seen[relPath] = true

		// Check Cache
		if entry, hit := r.cache.Get(relPath, mtime); hit {
			// Cache Hit
			docs = append(docs, core.Document{
				ID: entry.ID,
				Metadata: map[string]interface{}{
					"title": entry.Title,
					"tags":  entry.Tags,
				},
			})
			return nil
		}

		// Cache Miss
		doc, err := r.Get(ctx, id)
		if err != nil {
			return nil // Skip unparseable
		}

		// Update Cache
		var title string
		var tags []string

		if t, ok := doc.Metadata["title"].(string); ok {
			title = t
		}
		if tArr, ok := doc.Metadata["tags"].([]interface{}); ok {
			for _, t := range tArr {
				if s, ok := t.(string); ok {
					tags = append(tags, s)
				}
			}
		}

		r.cache.Set(relPath, &indexEntry{
			ID:           id,
			Title:        title,
			Tags:         tags,
			LastModified: mtime,
		})

		docs = append(docs, doc)
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Save Cache
	r.cache.Prune(seen)
	if err := r.cache.Save(); err != nil {
		// Ignore save error
	}

	return docs, nil

}

func (r *Repository) flattenCollection(ctx context.Context, fullPath, relPath string) ([]core.Document, error) {
	ext := filepath.Ext(fullPath)
	if ext != ".csv" { // Only CSV implemented for now
		return nil, fmt.Errorf("unsupported collection format")
	}

	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, err
	}

	reader := csv.NewReader(bytes.NewReader(data))
	headers, err := reader.Read()
	if err != nil {
		return nil, err
	}

	idColName := r.getIDColumn(filepath.Base(fullPath))
	idCol := -1
	for i, h := range headers {
		if strings.EqualFold(h, idColName) {
			idCol = i
			break
		}
	}
	if idCol == -1 {
		// Valid CSV but missing the configured ID column.
		// Return error? Or empty list? Error is better to signal misconfiguration.
		return nil, fmt.Errorf("missing '%s' column in %s", idColName, filepath.Base(fullPath))
	}

	var docs []core.Document
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if len(row) <= idCol {
			continue
		}

		id := row[idCol]
		// Construct ID: relPath + "/" + id
		// e.g. "users.csv/jane"
		fullID := relPath + "/" + id

		doc := core.Document{
			ID:       fullID,
			Metadata: make(core.Metadata),
		}

		for i, h := range headers {
			val := row[i]
			if strings.ToLower(h) == "content" {
				doc.Content = val
			} else {
				doc.Metadata[h] = val
			}
		}
		docs = append(docs, doc)
	}
	return docs, nil
}

func (r *Repository) saveBatchToCollection(ctx context.Context, collectionPath, collectionExt string, batch map[string]core.Document) error {
	data, err := os.ReadFile(collectionPath)
	if err != nil {
		return err
	}

	if collectionExt == ".csv" {
		reader := csv.NewReader(bytes.NewReader(data))
		allRecords, err := reader.ReadAll()
		if err != nil {
			return err
		}

		if len(allRecords) == 0 {
			return fmt.Errorf("empty csv collection")
		}

		headers := allRecords[0]
		idColName := r.getIDColumn(filepath.Base(collectionPath))
		idCol := -1
		for i, h := range headers {
			if strings.EqualFold(h, idColName) {
				idCol = i
				break
			}
		}
		if idCol == -1 {
			return fmt.Errorf("csv collection missing '%s' column", idColName)
		}

		// Update rows in place
		existingIndices := make(map[string]int)
		for i := 1; i < len(allRecords); i++ {
			row := allRecords[i]
			if len(row) > idCol {
				existingIndices[row[idCol]] = i
			}
		}

		for key, doc := range batch {
			// Prepare row data
			newRow := make([]string, len(headers))
			newRow[idCol] = key

			for i, h := range headers {
				if i == idCol {
					continue
				}
				if strings.EqualFold(h, "content") {
					newRow[i] = doc.Content
					continue
				}
				if val, ok := doc.Metadata[h]; ok {
					newRow[i] = fmt.Sprintf("%v", val)
				} else {
					newRow[i] = "" // Replace with empty if missing
				}
			}

			if idx, ok := existingIndices[key]; ok {
				allRecords[idx] = newRow
			} else {
				allRecords = append(allRecords, newRow)
			}
		}

		// Serialize back
		var buf bytes.Buffer
		w := csv.NewWriter(&buf)
		if err := w.WriteAll(allRecords); err != nil {
			return err
		}
		w.Flush()

		// Atomic Write
		return writeFileAtomic(collectionPath, buf.Bytes(), 0644)
	}

	return fmt.Errorf("unsupported collection type for save")
}

func (r *Repository) getIDColumn(filename string) string {
	if col, ok := r.config.IDMap[filename]; ok {
		return col
	}
	return "id"
}

// Delete removes a note.
func (r *Repository) Delete(ctx context.Context, id string) error {
	filename := id
	ext := filepath.Ext(id)
	if ext == "" {
		ext = ".md"
		filename = id + ext
	}

	fullPath := filepath.Join(r.Path, filename)

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return fmt.Errorf("document not found")
	}

	if r.config.Gitless {
		if err := os.Remove(fullPath); err != nil {
			return fmt.Errorf("failed to remove file: %w", err)
		}
		return nil
	}

	unlock, err := r.git.Lock()
	if err != nil {
		return fmt.Errorf("failed to acquire git lock: %w", err)
	}
	defer unlock()

	if err := r.git.Rm(filename); err != nil {
		return fmt.Errorf("failed to git rm: %w", err)
	}

	if err := r.git.Commit("delete " + id); err != nil {
		return fmt.Errorf("failed to git commit: %w", err)
	}

	return nil
}

// IsGitInstalled checks if git is available in the system path.
func IsGitInstalled() bool {
	return git.IsInstalled()
}

// --- Serialization Helpers (Private) ---

func parse(r io.Reader, ext string) (*core.Document, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	doc := &core.Document{
		Metadata: make(core.Metadata),
	}

	switch ext {
	case ".json":
		var payload map[string]interface{}
		if err := json.Unmarshal(data, &payload); err != nil {
			return nil, fmt.Errorf("invalid json: %w", err)
		}
		if c, ok := payload["content"].(string); ok {
			doc.Content = c
			delete(payload, "content")
		}
		doc.Metadata = payload

	case ".yaml", ".yml":
		var payload map[string]interface{}
		if err := yaml.Unmarshal(data, &payload); err != nil {
			return nil, fmt.Errorf("invalid yaml: %w", err)
		}
		if c, ok := payload["content"].(string); ok {
			doc.Content = c
			delete(payload, "content")
		}
		doc.Metadata = payload

	case ".csv":
		reader := csv.NewReader(bytes.NewReader(data))
		headers, err := reader.Read()
		if err != nil {
			return nil, fmt.Errorf("failed to read csv header: %w", err)
		}
		row, err := reader.Read()
		if err != nil {
			return nil, fmt.Errorf("failed to read csv row: %w", err)
		}

		if len(row) != len(headers) {
			return nil, fmt.Errorf("csv row length mismatch")
		}

		for i, h := range headers {
			if strings.ToLower(h) == "content" {
				doc.Content = row[i]
			} else {
				doc.Metadata[h] = row[i]
			}
		}

	case ".md":
		fallthrough
	default:
		if !bytes.HasPrefix(data, []byte("---\n")) && !bytes.HasPrefix(data, []byte("---\r\n")) {
			doc.Content = string(data)
			return doc, nil
		}

		rest := data[3:]
		parts := bytes.SplitN(rest, []byte("---"), 2)
		if len(parts) == 1 {
			return nil, errors.New("frontmatter started but no closing delimiter found")
		}

		yamlData := parts[0]
		contentData := parts[1]

		if err := yaml.Unmarshal(yamlData, &doc.Metadata); err != nil {
			return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
		}

		doc.Content = strings.TrimPrefix(string(contentData), "\n")
		doc.Content = strings.TrimPrefix(doc.Content, "\r\n")
	}

	return doc, nil
}

func serialize(doc core.Document, ext string) ([]byte, error) {
	switch ext {
	case ".json":
		payload := make(map[string]interface{})
		for k, v := range doc.Metadata {
			payload[k] = v
		}
		payload["content"] = doc.Content
		return json.MarshalIndent(payload, "", "  ")

	case ".yaml", ".yml":
		payload := make(map[string]interface{})
		for k, v := range doc.Metadata {
			payload[k] = v
		}
		payload["content"] = doc.Content
		return yaml.Marshal(payload)

	case ".csv":
		keys := []string{"content"}
		for k := range doc.Metadata {
			keys = append(keys, k)
		}

		var row []string
		row = append(row, doc.Content)
		for _, k := range keys[1:] {
			val := ""
			if v := doc.Metadata[k]; v != nil {
				val = fmt.Sprintf("%v", v)
			}
			row = append(row, val)
		}

		var buf bytes.Buffer
		w := csv.NewWriter(&buf)
		if err := w.Write(keys); err != nil {
			return nil, err
		}
		if err := w.Write(row); err != nil {
			return nil, err
		}
		w.Flush()
		return buf.Bytes(), nil

	case ".md":
		fallthrough
	default:
		var buf bytes.Buffer
		if len(doc.Metadata) > 0 {
			buf.WriteString("---\n")
			encoder := yaml.NewEncoder(&buf)
			encoder.SetIndent(2)
			if err := encoder.Encode(doc.Metadata); err != nil {
				return nil, err
			}
			encoder.Close()
			buf.WriteString("---\n")
		}
		buf.WriteString(doc.Content)
		return buf.Bytes(), nil
	}
}

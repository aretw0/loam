package git

import (
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
	"sync"
)

// Client wraps git command execution with a global lock for minimal thread safety
// within a single process.
type Client struct {
	WorkDir string
	Logger  *slog.Logger
	mu      sync.Mutex
}

// NewClient creates a new git client for the given working directory.
func NewClient(workDir string, logger *slog.Logger) *Client {
	return &Client{
		WorkDir: workDir,
		Logger:  logger,
	}
}

// Run executes a raw git command in the working directory.
// It is protected by the client's mutex to ensure sequential access.
func (c *Client) Run(args ...string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.Logger != nil {
		c.Logger.Debug("executing git", "args", args, "dir", c.WorkDir)
	}

	cmd := exec.Command("git", args...)
	cmd.Dir = c.WorkDir

	out, err := cmd.CombinedOutput()
	output := string(out)

	if err != nil {
		return output, fmt.Errorf("git %s failed: %w\nOutput: %s", args[0], err, output)
	}

	return strings.TrimSpace(output), nil
}

// Init initializes a new git repository if one doesn't exist.
func (c *Client) Init() error {
	// Check if already exists to avoid error? git init is safe to re-run.
	_, err := c.Run("init")
	return err
}

// Add adds files to the stage.
func (c *Client) Add(files ...string) error {
	if len(files) == 0 {
		return nil
	}
	args := append([]string{"add"}, files...)
	_, err := c.Run(args...)
	return err
}

// Rm removes files from the working tree and from the index.
func (c *Client) Rm(files ...string) error {
	if len(files) == 0 {
		return nil
	}
	args := append([]string{"rm", "-f"}, files...)
	_, err := c.Run(args...)
	return err
}

// Commit records changes to the repository.
func (c *Client) Commit(msg string) error {
	_, err := c.Run("commit", "-m", msg)
	return err
}

// Status returns the porcelain status of the repo.
func (c *Client) Status() (string, error) {
	return c.Run("status", "--porcelain")
}

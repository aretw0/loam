package git

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Client wraps git command execution with a global file-based lock for process safety.
type Client struct {
	WorkDir  string
	Logger   *slog.Logger
	lockPath string
}

// NewClient creates a new git client for the given working directory.
func NewClient(workDir string, logger *slog.Logger) *Client {
	return &Client{
		WorkDir:  workDir,
		Logger:   logger,
		lockPath: ".loam.lock", // Lock file name
	}
}

// Lock acquires a file-based lock. It blocks until the lock is acquired.
func (c *Client) Lock() (func(), error) {
	fullLockPath := filepath.Join(c.WorkDir, c.lockPath)

	for {
		// Try to create lock file atomically
		f, err := os.OpenFile(fullLockPath, os.O_CREATE|os.O_EXCL, 0666)
		if err == nil {
			f.Close()
			// Return unlock function
			return func() {
				os.Remove(fullLockPath)
			}, nil
		}

		if os.IsExist(err) {
			// Lock exists, wait and retry
			// Simple spinlock with backoff.
			// TODO: Add timeout to prevent infinite deadlocks?
			time.Sleep(10 * time.Millisecond)
			continue
		}

		return nil, fmt.Errorf("failed to acquire lock: %w", err)
	}
}

// Run executes a raw git command in the working directory.
// NOTE: It does NOT acquire the lock automatically. The caller must manage transaction safety via Client.Lock().
func (c *Client) Run(args ...string) (string, error) {
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

// HasRemote checks if there is a 'origin' remote configured.
// For now, we hardcode 'origin' as the default remote to check.
func (c *Client) HasRemote() bool {
	// simple check: git remote get-url origin
	_, err := c.Run("remote", "get-url", "origin")
	return err == nil
}

// Sync performs a pull --rebase and then a push.
// It assumes the caller handles locking if necessary, though git operations themselves are somewhat atomic,
// coordinating multiple git commands usually requires a lock to prevent state changes in between.
func (c *Client) Sync() error {
	// 1. Pull --rebase
	// We want to rebase local changes on top of upstream
	if _, err := c.Run("pull", "--rebase", "origin", "main"); err != nil {
		// Try master if main fails? Or just rely on default branch config?
		// For now let's just do 'git pull --rebase' which uses the tracked branch.
		// If no upstream is set, this might fail.
		// Let's retry without arguments to rely on default tracking configuration.
		if _, err2 := c.Run("pull", "--rebase"); err2 != nil {
			return fmt.Errorf("pull --rebase failed: %w (ensure you have set up a tracking branch)", err2)
		}
	}

	// 2. Push
	if _, err := c.Run("push"); err != nil {
		return fmt.Errorf("push failed: %w", err)
	}

	return nil
}

// Restore restores working tree files (discards changes in working directory).
// Use with caution.
func (c *Client) Restore(files ...string) error {
	if len(files) == 0 {
		return nil
	}
	// git restore <files>
	args := append([]string{"restore"}, files...)
	_, err := c.Run(args...)
	return err
}

// Clean removes untracked files from the working tree.
func (c *Client) Clean(files ...string) error {
	if len(files) == 0 {
		return nil
	}
	// git clean -f <files>
	// careful: git clean usually takes a path. If we pass specific files, it might need -f.
	// Actually `git clean` works on paths.
	args := append([]string{"clean", "-f"}, files...)
	_, err := c.Run(args...)
	return err
}

// IsRepo checks if the current working directory is a valid git repository.
func (c *Client) IsRepo() bool {
	// Check for .git directory
	gitDir := filepath.Join(c.WorkDir, ".git")
	info, err := os.Stat(gitDir)
	return err == nil && info.IsDir()
}

// IsInstalled checks if git is available in the system PATH.
func IsInstalled() bool {
	_, err := exec.LookPath("git")
	return err == nil
}

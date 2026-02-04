// Package git provides Git repository operations.
package git

import (
	"fmt"
	"strings"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// Repository interface defines Git operations
type Repository interface {
	CurrentBranch() (string, error)
	CreateBranch(name string) error
	Checkout(branch string) error
	Status() (map[string]string, error)
	GetCommits(branch, baseBranch string) ([]*Commit, error)
	Push(branch string) error
	BranchExists(name string) (bool, error)
	GetRemoteBranch(branch string) (string, error)
	GetRootPath() (string, error)
}

// GitRepository implements Repository using go-git
//
//nolint:revive // Type name matches domain terminology
type GitRepository struct {
	repo *git.Repository
	path string
}

// Commit represents a git commit
type Commit struct {
	Hash    string
	Message string
	Author  string
	Date    string
}

// OpenRepository opens a git repository at the given path
func OpenRepository(path string) (*GitRepository, error) {
	repo, err := git.PlainOpen(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}

	return &GitRepository{
		repo: repo,
		path: path,
	}, nil
}

// CurrentBranch returns the current branch name
func (r *GitRepository) CurrentBranch() (string, error) {
	head, err := r.repo.Head()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD: %w", err)
	}

	if !head.Name().IsBranch() {
		return "", fmt.Errorf("HEAD is not a branch")
	}

	return head.Name().Short(), nil
}

// CreateBranch creates a new branch
func (r *GitRepository) CreateBranch(name string) error {
	head, err := r.repo.Head()
	if err != nil {
		return fmt.Errorf("failed to get HEAD: %w", err)
	}

	refName := plumbing.NewBranchReferenceName(name)
	ref := plumbing.NewHashReference(refName, head.Hash())

	err = r.repo.Storer.SetReference(ref)
	if err != nil {
		return fmt.Errorf("failed to create branch: %w", err)
	}

	return nil
}

// Checkout switches to a different branch
func (r *GitRepository) Checkout(branch string) error {
	w, err := r.repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	err = w.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branch),
	})
	if err != nil {
		return fmt.Errorf("failed to checkout branch: %w", err)
	}

	return nil
}

// Status returns the status of the working tree
func (r *GitRepository) Status() (map[string]string, error) {
	w, err := r.repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("failed to get worktree: %w", err)
	}

	status, err := w.Status()
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %w", err)
	}

	result := make(map[string]string)
	for file, fileStatus := range status {
		var statusStr string
		switch {
		case fileStatus.Staging == git.Added:
			statusStr = "added"
		case fileStatus.Staging == git.Modified:
			statusStr = "modified"
		case fileStatus.Staging == git.Deleted:
			statusStr = "deleted"
		case fileStatus.Worktree == git.Modified:
			statusStr = "modified"
		case fileStatus.Worktree == git.Deleted:
			statusStr = "deleted"
		case fileStatus.Staging == git.Untracked:
			statusStr = "untracked"
		default:
			statusStr = "unknown"
		}
		result[file] = statusStr
	}

	return result, nil
}

// GetCommits returns commits between base branch and target branch
func (r *GitRepository) GetCommits(branch, baseBranch string) ([]*Commit, error) {
	// Get branch reference
	branchRef, err := r.repo.Reference(plumbing.NewBranchReferenceName(branch), true)
	if err != nil {
		return nil, fmt.Errorf("failed to get branch reference: %w", err)
	}

	// Get base branch reference
	baseRef, err := r.repo.Reference(plumbing.NewBranchReferenceName(baseBranch), true)
	if err != nil {
		return nil, fmt.Errorf("failed to get base branch reference: %w", err)
	}

	// Get commit iterator
	commitIter, err := r.repo.Log(&git.LogOptions{
		From: branchRef.Hash(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get commit log: %w", err)
	}
	defer commitIter.Close()

	var commits []*Commit
	baseHash := baseRef.Hash()

	err = commitIter.ForEach(func(c *object.Commit) error {
		// Stop when we reach the base branch
		if c.Hash == baseHash {
			return fmt.Errorf("reached base branch")
		}

		commits = append(commits, &Commit{
			Hash:    c.Hash.String(),
			Message: strings.TrimSpace(c.Message),
			Author:  c.Author.Name,
			Date:    c.Author.When.Format("2006-01-02 15:04:05"),
		})

		return nil
	})

	// Ignore the error about reaching base branch
	if err != nil && err.Error() != "reached base branch" {
		return nil, fmt.Errorf("failed to iterate commits: %w", err)
	}

	return commits, nil
}

// Push pushes a branch to remote
func (r *GitRepository) Push(branch string) error {
	err := r.repo.Push(&git.PushOptions{
		RemoteName: "origin",
		RefSpecs: []config.RefSpec{
			config.RefSpec(fmt.Sprintf("refs/heads/%s:refs/heads/%s", branch, branch)),
		},
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("failed to push: %w", err)
	}

	return nil
}

// BranchExists checks if a branch exists locally
func (r *GitRepository) BranchExists(name string) (bool, error) {
	_, err := r.repo.Reference(plumbing.NewBranchReferenceName(name), true)
	if err == plumbing.ErrReferenceNotFound {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check branch: %w", err)
	}
	return true, nil
}

// GetRemoteBranch returns the remote tracking branch for a local branch
func (r *GitRepository) GetRemoteBranch(branch string) (string, error) {
	_, err := r.repo.Reference(plumbing.NewBranchReferenceName(branch), true)
	if err != nil {
		return "", fmt.Errorf("failed to get branch: %w", err)
	}

	cfg, err := r.repo.Config()
	if err != nil {
		return "", fmt.Errorf("failed to get config: %w", err)
	}

	for _, b := range cfg.Branches {
		if b.Name == branch {
			return fmt.Sprintf("%s/%s", b.Remote, branch), nil
		}
	}

	return "", fmt.Errorf("no remote tracking branch found")
}

// GetRootPath returns the root path of the repository
func (r *GitRepository) GetRootPath() (string, error) {
	w, err := r.repo.Worktree()
	if err != nil {
		return "", fmt.Errorf("failed to get worktree: %w", err)
	}
	return w.Filesystem.Root(), nil
}

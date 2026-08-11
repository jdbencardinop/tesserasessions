package statusprovider

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveGitUsesCommonRepositoryRootForLinkedWorktree(t *testing.T) {
	repo := filepath.Join(t.TempDir(), "repo")
	worktree := filepath.Join(t.TempDir(), "worktree")
	runGitTestCommand(t, "", "init", repo)
	runGitTestCommand(t, repo, "config", "user.name", "Test User")
	runGitTestCommand(t, repo, "config", "user.email", "test@example.com")
	if err := os.WriteFile(filepath.Join(repo, "README.md"), []byte("test\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGitTestCommand(t, repo, "add", "README.md")
	runGitTestCommand(t, repo, "commit", "-m", "initial")
	runGitTestCommand(t, repo, "branch", "-M", "main")
	runGitTestCommand(t, repo, "worktree", "add", "-b", "feature/status", worktree)

	root, branch, err := resolveGit(context.Background(), worktree)
	if err != nil {
		t.Fatal(err)
	}
	wantRoot, err := filepath.EvalSymlinks(repo)
	if err != nil {
		t.Fatal(err)
	}
	gotRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}
	if gotRoot != wantRoot {
		t.Fatalf("repo root = %q, want common root %q", gotRoot, wantRoot)
	}
	if branch != "feature/status" {
		t.Fatalf("branch = %q, want feature/status", branch)
	}
}

func runGitTestCommand(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %s failed: %v\n%s", strings.Join(args, " "), err, out)
	}
}

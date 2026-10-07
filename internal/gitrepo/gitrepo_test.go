package gitrepo

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestActivity(t *testing.T) {
	dir := t.TempDir()
	run(t, dir, "git", "init")
	run(t, dir, "git", "config", "user.name", "Test User")
	run(t, dir, "git", "config", "user.email", "test@example.com")

	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("one\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	run(t, dir, "git", "add", "a.txt")
	runEnv(t, dir, []string{
		"GIT_AUTHOR_DATE=2026-10-05T10:00:00-07:00",
		"GIT_COMMITTER_DATE=2026-10-05T10:01:00-07:00",
	}, "git", "commit", "-m", "first")

	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("one\ntwo\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	run(t, dir, "git", "add", "a.txt")
	runEnv(t, dir, []string{
		"GIT_AUTHOR_DATE=2026-10-06T11:00:00-07:00",
		"GIT_COMMITTER_DATE=2026-10-06T11:01:00-07:00",
	}, "git", "commit", "-m", "second")

	repo, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}

	first, err := repo.FirstCommitTime("HEAD")
	if err != nil {
		t.Fatal(err)
	}
	if want := "2026-10-05T10:01:00-07:00"; first.Format(time.RFC3339) != want {
		t.Fatalf("first commit time = %s, want %s", first.Format(time.RFC3339), want)
	}

	start := mustParse(t, "2026-10-05T00:00:00-07:00")
	end := mustParse(t, "2026-10-12T00:00:00-07:00")

	got, err := repo.Activity("HEAD", start, end)
	if err != nil {
		t.Fatal(err)
	}

	if len(got.Commits) != 2 {
		t.Fatalf("commits = %d, want 2", len(got.Commits))
	}
	if got.FilesChanged != 1 {
		t.Fatalf("files changed = %d, want 1", got.FilesChanged)
	}
	if got.LinesAdded != 2 {
		t.Fatalf("lines added = %d, want 2", got.LinesAdded)
	}
}

func run(t *testing.T, dir, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%s %v: %v\n%s", name, args, err, out)
	}
}

func runEnv(t *testing.T, dir string, env []string, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), env...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%s %v: %v\n%s", name, args, err, out)
	}
}

func mustParse(t *testing.T, value string) time.Time {
	t.Helper()
	got, err := time.Parse(time.RFC3339, value)
	if err != nil {
		t.Fatal(err)
	}
	return got
}

package gitrepo

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Repository struct { path string }

type Commit struct {
	SHA string
	CommitTime time.Time
	AuthorTime time.Time
	Author string
	Subject string
}

type Activity struct {
	Commits []Commit
	FilesChanged int
	LinesAdded int
	LinesRemoved int
}

func Open(path string) (*Repository, error) {
	abs, err := filepath.Abs(path)
	if err != nil { return nil, err }
	out, err := exec.Command("git", "-C", abs, "rev-parse", "--show-toplevel").Output()
	if err != nil { return nil, fmt.Errorf("%q is not a Git repository: %w", path, err) }
	return &Repository{path: strings.TrimSpace(string(out))}, nil
}

func (r *Repository) Path() string { return r.path }

func (r *Repository) FirstCommitTime(ref string) (time.Time, error) {
	out, err := exec.Command("git", "-C", r.path, "rev-list", "--reverse", ref).Output()
	if err != nil { return time.Time{}, fmt.Errorf("find first commit: %w", err) }
	lines := strings.Fields(string(out))
	if len(lines) == 0 { return time.Time{}, fmt.Errorf("ref %q has no commits", ref) }

	firstSHA := lines[0]
	out, err = exec.Command("git", "-C", r.path, "show", "-s", "--format=%cI", firstSHA).Output()
	if err != nil { return time.Time{}, fmt.Errorf("read first commit time: %w", err) }
	value := strings.TrimSpace(string(out))
	t, err := time.Parse(time.RFC3339, value)
	if err != nil { return time.Time{}, fmt.Errorf("parse first commit time: %w", err) }
	return t, nil
}

func (r *Repository) Activity(ref string, start, endExclusive time.Time) (Activity, error) {
	commits, err := r.commits(ref, start, endExclusive)
	if err != nil { return Activity{}, err }
	filesChanged, added, removed, err := r.stats(ref, start, endExclusive)
	if err != nil { return Activity{}, err }
	return Activity{Commits: commits, FilesChanged: filesChanged, LinesAdded: added, LinesRemoved: removed}, nil
}

func (r *Repository) commits(ref string, start, endExclusive time.Time) ([]Commit, error) {
	const sep = "\x1f"
	const rec = "\x1e"
	format := "%H%x1f%cI%x1f%aI%x1f%an%x1f%s%x1e"
	out, err := exec.Command("git", "-C", r.path, "log", "--reverse",
		"--since="+start.Format(time.RFC3339),
		"--until="+endExclusive.Format(time.RFC3339),
		"--format="+format, ref).Output()
	if err != nil { return nil, fmt.Errorf("git log: %w", err) }

	raw := strings.Trim(string(out), rec+"\n")
	if raw == "" { return nil, nil }
	records := strings.Split(raw, rec)
	result := make([]Commit, 0, len(records))
	for _, record := range records {
		record = strings.TrimSpace(record)
		if record == "" { continue }
		parts := strings.Split(record, sep)
		if len(parts) != 5 { return nil, fmt.Errorf("unexpected git log record with %d fields", len(parts)) }
		ct, err := time.Parse(time.RFC3339, parts[1])
		if err != nil { return nil, fmt.Errorf("parse commit time: %w", err) }
		at, err := time.Parse(time.RFC3339, parts[2])
		if err != nil { return nil, fmt.Errorf("parse author time: %w", err) }
		result = append(result, Commit{SHA: parts[0], CommitTime: ct, AuthorTime: at, Author: parts[3], Subject: parts[4]})
	}
	return result, nil
}

func (r *Repository) stats(ref string, start, endExclusive time.Time) (int, int, int, error) {
	out, err := exec.Command("git", "-C", r.path, "log",
		"--since="+start.Format(time.RFC3339),
		"--until="+endExclusive.Format(time.RFC3339),
		"--format=", "--numstat", ref).Output()
	if err != nil { return 0, 0, 0, fmt.Errorf("git log --numstat: %w", err) }

	files := map[string]struct{}{}
	added, removed := 0, 0
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" { continue }
		parts := strings.SplitN(line, "\t", 3)
		if len(parts) != 3 { continue }
		files[parts[2]] = struct{}{}
		if n, err := strconv.Atoi(parts[0]); err == nil { added += n }
		if n, err := strconv.Atoi(parts[1]); err == nil { removed += n }
	}
	if err := scanner.Err(); err != nil { return 0, 0, 0, err }
	return len(files), added, removed, nil
}

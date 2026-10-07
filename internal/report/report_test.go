package report

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ciroque/week-in-review/internal/gitrepo"
)

func TestResolveExplicitSunday(t *testing.T) {
	loc := time.FixedZone("test", -7*60*60)

	got, err := ResolveWeekEnding(time.Time{}, loc, "2026-10-11")
	if err != nil {
		t.Fatal(err)
	}
	if got.Format("2006-01-02") != "2026-10-11" {
		t.Fatalf("got %s", got)
	}
}

func TestResolveExplicitNonSundayFails(t *testing.T) {
	if _, err := ResolveWeekEnding(time.Time{}, time.UTC, "2026-10-10"); err == nil {
		t.Fatal("expected error")
	}
}

func TestDefaultOnMondayUsesPreviousDay(t *testing.T) {
	now := time.Date(2026, 10, 12, 8, 0, 0, 0, time.UTC)

	got, err := ResolveWeekEnding(now, time.UTC, "")
	if err != nil {
		t.Fatal(err)
	}
	if want := "2026-10-11"; got.Format("2006-01-02") != want {
		t.Fatalf("got %s, want %s", got.Format("2006-01-02"), want)
	}
}

func TestDefaultOnSundayUsesPriorSunday(t *testing.T) {
	now := time.Date(2026, 10, 11, 23, 59, 0, 0, time.UTC)

	got, err := ResolveWeekEnding(now, time.UTC, "")
	if err != nil {
		t.Fatal(err)
	}
	if want := "2026-10-04"; got.Format("2006-01-02") != want {
		t.Fatalf("got %s, want %s", got.Format("2006-01-02"), want)
	}
}

func TestInitCreatesEmptyJournal(t *testing.T) {
	path := filepath.Join(t.TempDir(), "WEEK-IN-REVIEW.md")

	if err := Init(path, "America/Los_Angeles"); err != nil {
		t.Fatal(err)
	}

	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)

	for _, want := range []string{
		"# Week in Review",
		"append-only chronological project journal",
		"Reporting timezone: `America/Los_Angeles`.",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %q in:\n%s", want, text)
		}
	}

	if strings.Contains(text, "## Week of") {
		t.Fatalf("initialized journal unexpectedly contains a weekly entry:\n%s", text)
	}
}

func TestInitRefusesToOverwriteExistingJournal(t *testing.T) {
	path := filepath.Join(t.TempDir(), "WEEK-IN-REVIEW.md")
	if err := os.WriteFile(path, []byte("existing\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := Init(path, "UTC"); err == nil {
		t.Fatal("expected existing-file error")
	}

	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "existing\n" {
		t.Fatalf("existing journal was modified: %q", body)
	}
}

func TestAppendCreatesJournalAndRejectsDuplicate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "WEEK-IN-REVIEW.md")
	loc := time.UTC

	period := PeriodForWeekEnding(time.Date(2026, 10, 11, 0, 0, 0, 0, loc), loc)
	activity := gitrepo.Activity{
		Commits: []gitrepo.Commit{
			{
				SHA:        "1234567890abcdef",
				CommitTime: time.Date(2026, 10, 5, 10, 0, 0, 0, loc),
				AuthorTime: time.Date(2026, 10, 5, 9, 59, 0, 0, loc),
				Author:     "A | B",
				Subject:    "Do | thing",
			},
		},
		FilesChanged: 2,
		LinesAdded:   10,
		LinesRemoved: 3,
	}

	if err := Append(path, "UTC", period, activity, time.Date(2026, 10, 12, 1, 0, 0, 0, loc)); err != nil {
		t.Fatal(err)
	}

	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)

	for _, want := range []string{
		"append-only chronological project journal",
		"## Week of 2026-10-05 through 2026-10-11",
		"`1234567`",
		"A \\| B",
		"Do \\| thing",
		"**Commits:** 1",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %q in:\n%s", want, text)
		}
	}

	if err := Append(path, "UTC", period, activity, time.Now()); err == nil {
		t.Fatal("expected duplicate-week error")
	}
}

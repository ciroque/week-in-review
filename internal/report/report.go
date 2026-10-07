package report

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ciroque/week-in-review/internal/gitrepo"
)

type Period struct {
	Start        time.Time
	End          time.Time
	EndExclusive time.Time
}

func ResolveWeekEnding(now time.Time, loc *time.Location, explicit string) (time.Time, error) {
	if explicit != "" {
		t, err := time.ParseInLocation("2006-01-02", explicit, loc)
		if err != nil {
			return time.Time{}, fmt.Errorf("parse --week-ending: %w", err)
		}
		if t.Weekday() != time.Sunday {
			return time.Time{}, fmt.Errorf("--week-ending must be a Sunday: %s", explicit)
		}
		return midnight(t, loc), nil
	}

	localNow := now.In(loc)
	today := midnight(localNow, loc)

	daysSinceSunday := int(today.Weekday())
	if daysSinceSunday == 0 {
		daysSinceSunday = 7
	}

	return today.AddDate(0, 0, -daysSinceSunday), nil
}

func PeriodForWeekEnding(end time.Time, loc *time.Location) Period {
	end = midnight(end, loc)
	start := end.AddDate(0, 0, -6)
	return Period{
		Start:        start,
		End:          end,
		EndExclusive: end.AddDate(0, 0, 1),
	}
}

func Init(path, timezone string) error {
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("%s already exists; refusing to overwrite it", path)
	} else if !os.IsNotExist(err) {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	return os.WriteFile(path, []byte(journalHeader(timezone)), 0o644)
}

func Append(path, timezone string, period Period, activity gitrepo.Activity, generated time.Time) error {
	existing, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	header := weekHeader(period)
	if strings.Contains(string(existing), header) {
		return fmt.Errorf("week %s through %s already exists; journal is append-only",
			period.Start.Format("2006-01-02"),
			period.End.Format("2006-01-02"),
		)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	if len(existing) == 0 {
		if _, err := f.WriteString(journalHeader(timezone)); err != nil {
			return err
		}
	}

	var b strings.Builder
	fmt.Fprintf(&b, "\n%s\n\n", header)
	fmt.Fprintf(&b, "Generated: %s\n\n", code(generated.Format(time.RFC3339)))
	b.WriteString("| Commit Time | Author Time | Commit | Author | Description |\n")
	b.WriteString("|---|---|---|---|---|\n")

	for _, c := range activity.Commits {
		short := c.SHA
		if len(short) > 7 {
			short = short[:7]
		}
		fmt.Fprintf(
			&b,
			"| %s | %s | %s | %s | %s |\n",
			code(c.CommitTime.Format(time.RFC3339)),
			code(c.AuthorTime.Format(time.RFC3339)),
			code(short),
			escape(c.Author),
			escape(c.Subject),
		)
		fmt.Fprintf(&b, "<!-- commit: %s -->\n", c.SHA)
	}

	fmt.Fprintf(&b, "\n### Activity\n\n- **Commits:** %d\n- **Files changed:** %d\n- **Lines added:** %d\n- **Lines removed:** %d\n\n---\n",
		len(activity.Commits), activity.FilesChanged, activity.LinesAdded, activity.LinesRemoved)

	_, err = f.WriteString(b.String())
	return err
}

func journalHeader(timezone string) string {
	return fmt.Sprintf("# Week in Review\n\nThis document is an append-only chronological project journal generated entirely from objective repository metadata.\n\nGit history is the source of truth. Entries are generated automatically and previously recorded weeks are not regenerated or modified.\n\nReporting timezone: %s.\n\n---\n", code(timezone))
}

func weekHeader(period Period) string {
	return fmt.Sprintf(
		"## Week of %s through %s",
		period.Start.Format("2006-01-02"),
		period.End.Format("2006-01-02"),
	)
}

func midnight(t time.Time, loc *time.Location) time.Time {
	t = t.In(loc)
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc)
}

func escape(s string) string {
	s = strings.ReplaceAll(s, "|", "\\|")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	return s
}

func code(s string) string {
	return "`" + strings.ReplaceAll(s, "`", "\\`") + "`"
}

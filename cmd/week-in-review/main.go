package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ciroque/week-in-review/internal/gitrepo"
	"github.com/ciroque/week-in-review/internal/report"
)

func main() {
	var repoPath, output, timezone, weekEnding, ref string
	var initJournal bool

	flag.StringVar(&repoPath, "repo", ".", "Git repository to inspect")
	flag.StringVar(&output, "output", "WEEK-IN-REVIEW.md", "Journal path, relative to --repo")
	flag.StringVar(&timezone, "timezone", "America/Los_Angeles", "IANA timezone used for reporting boundaries")
	flag.StringVar(&weekEnding, "week-ending", "", "Sunday ending the reporting week, YYYY-MM-DD; defaults to the most recently completed Sunday")
	flag.StringVar(&ref, "ref", "HEAD", "Git ref to inspect")
	flag.BoolVar(&initJournal, "init", false, "Build a new journal from all repository history, including the current week")
	flag.Parse()

	loc, err := time.LoadLocation(timezone)
	if err != nil { fatalf("load timezone %q: %v", timezone, err) }

	repo, err := gitrepo.Open(repoPath)
	if err != nil { fatalf("open repository: %v", err) }

	outputPath := output
	if !filepath.IsAbs(outputPath) { outputPath = filepath.Join(repo.Path(), outputPath) }

	if initJournal {
		if weekEnding != "" { fatalf("--week-ending cannot be used with --init") }
		first, err := repo.FirstCommitTime(ref)
		if err != nil { fatalf("find first commit: %v", err) }

		periods := report.PeriodsThrough(first, time.Now(), loc)
		entries := make([]report.Entry, 0, len(periods))
		for _, period := range periods {
			activity, err := repo.Activity(ref, period.Start, period.EndExclusive)
			if err != nil { fatalf("read git history for week ending %s: %v", period.End.Format("2006-01-02"), err) }
			if len(activity.Commits) == 0 { continue }
			entries = append(entries, report.Entry{Period: period, Activity: activity})
		}

		if err := report.Init(outputPath, timezone, entries, time.Now().In(loc)); err != nil {
			fatalf("initialize journal: %v", err)
		}
		fmt.Printf("initialized %s with %d active weeks\n", outputPath, len(entries))
		return
	}

	end, err := report.ResolveWeekEnding(time.Now(), loc, weekEnding)
	if err != nil { fatalf("resolve reporting period: %v", err) }
	period := report.PeriodForWeekEnding(end, loc)

	activity, err := repo.Activity(ref, period.Start, period.EndExclusive)
	if err != nil { fatalf("read git history: %v", err) }

	if err := report.Append(outputPath, timezone, period, activity, time.Now().In(loc)); err != nil {
		fatalf("write report: %v", err)
	}

	fmt.Printf("recorded week %s through %s in %s\n", period.Start.Format("2006-01-02"), period.End.Format("2006-01-02"), outputPath)
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "week-in-review: "+format+"\n", args...)
	os.Exit(1)
}

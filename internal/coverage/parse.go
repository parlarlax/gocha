package coverage

import (
	"fmt"

	"golang.org/x/tools/cover"
)

type FileCoverage struct {
	FileName  string
	Covered   int64
	Total     int64
	Percent   float64
	Profiles  []*cover.Profile
}

type Report struct {
	Files   []FileCoverage
	Total   float64
}

func Parse(coverprofile string) (*Report, error) {
	profiles, err := cover.ParseProfiles(coverprofile)
	if err != nil {
		return nil, fmt.Errorf("parse coverage profile: %w", err)
	}

	report := &Report{}
	var totalStmts, coveredStmts int64

	for _, p := range profiles {
		var covered, total int64
		for _, block := range p.Blocks {
			stmts := int64(block.NumStmt)
			total += stmts
			if block.Count > 0 {
				covered += stmts
			}
		}

		var pct float64
		if total > 0 {
			pct = float64(covered) / float64(total) * 100
		}

		report.Files = append(report.Files, FileCoverage{
			FileName: p.FileName,
			Covered:  covered,
			Total:    total,
			Percent:  pct,
			Profiles: []*cover.Profile{p},
		})

		totalStmts += total
		coveredStmts += covered
	}

	if totalStmts > 0 {
		report.Total = float64(coveredStmts) / float64(totalStmts) * 100
	}

	return report, nil
}

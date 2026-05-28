# Changelog

All notable changes to gocha will be documented in this file.

## [v0.6.3] - 2026-05-28

### Fixed
- **Git dirty detection** — changed from `git status --porcelain` to `git diff HEAD --quiet` so untracked files (e.g. generated `report.html`, `coverage.out`) no longer incorrectly trigger `+dirty` in the report header

## [v0.6.0] - 2026-05-28

### Added
- **Failed test output** — click any failed test row to expand inline error log (captures up to 16KB tail of test output)
- **Slowest Tests panel** — top 10 slowest non-skipped tests shown above the package list
- **Total run duration** — wall-clock time shown in header pass-rate bar and as a stat card; computed from first/last event timestamp
- **Markdown summary export** — `-md <path>` flag writes a PR-friendly summary (use `-md -` for stdout); includes verdict line, stats table, failed test list in `<details>`, and coverage total

## [v0.5.0] - 2026-05-28

### Added
- Project name shown in report header with purple accent bar
- Auto-detected from `go.mod` (last module path segment) by default
- `-title` flag to override with a custom project name

## [v0.4.0] - 2026-05-28

### Added
- Dark/light theme toggle in report header with 🌙/☀️ icon — preference persisted via localStorage
- 🍵 favicon embedded as inline SVG data URI (no external file needed)

### Changed
- Light theme redesigned with lavender-tinted palette (no pure white) for reduced eye strain

## [v0.3.0] - 2026-05-28

### Added
- Branded report header: 🍵 icon, tagline "Go test report generator", GitHub link, version badge
- Version auto-detected from build info (`runtime/debug`) — shows tagged version when installed via `go install`

### Changed
- Test name column now truncates long names with ellipsis; hover shows full name via native tooltip
- Status and Duration columns fixed-width so columns stay aligned across all package tables

## [v0.2.0] - 2026-05-21

### Added
- Pass rate progress bar in report header
- Tests grouped by package with collapsible sections
- Search box and filter buttons (All / Pass / Fail / Skip) in Tests tab
- Slow test badge (⚡) for tests taking ≥ 1s
- Tab counts showing total tests, coverage %, and source file count

### Changed
- Failed tests now appear first within each package group

## [v0.1.0] - 2026-05-21

### Added
- Initial release
- Parse `go test -json` output into structured test results
- Parse coverage profiles using `golang.org/x/tools/cover`
- Generate self-contained HTML report with 3 tabs: Tests, Coverage, Source
- Source code highlighted green (covered) / red (uncovered)
- Runner mode: `gocha -- ./...` runs `go test` internally
- Consumer mode: `go test -json ./... | gocha -cover coverage.out`
- Flag passthrough via `--` separator
- Temp coverprofile in runner mode (no cwd pollution)
- Exit code preserved from `go test` (CI-friendly)
- MIT License
- GitHub Actions CI workflow example

[v0.2.0]: https://github.com/parlarlax/gocha/compare/v0.1.0...v0.2.0
[v0.1.0]: https://github.com/parlarlax/gocha/releases/tag/v0.1.0

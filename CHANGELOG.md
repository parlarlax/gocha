# Changelog

All notable changes to gocha will be documented in this file.

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

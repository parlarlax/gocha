// Package gocha generates a single self-contained HTML report combining
// Go test results and coverage data.
//
// # Runner mode
//
//	gocha [flags] -- [go test flags] <packages>
//
// # Consumer mode (pipe)
//
//	go test -json -coverprofile=c.out ./... | gocha -cover c.out
//
// See https://github.com/parlarlax/gocha for documentation.
package gocha

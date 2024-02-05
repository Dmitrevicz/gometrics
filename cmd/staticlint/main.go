// Package main - entry point for custom static check linter.
package main

import "github.com/Dmitrevicz/gometrics/pkg/staticlint"

// TODO: implement staticlint
// 1. [done] forbid os.Exit in main
// 2. custom multichecker including:
//   - golang.org/x/tools/go/analysis/passes
//   - all SA analyzators from staticcheck.io
//   - plus at least 1 more analyzer from staticcheck.io
//   - two or more public 3rd-party analyzers
// 3. describe godoc
// 4. go test coverage >= 55%

func main() {
	staticlint.Mount()
}

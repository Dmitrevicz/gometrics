// Package staticchecks defines functions to retrieve staticcheck.io analyzers.
//
// https://staticcheck.dev/docs/checks/
// https://pkg.go.dev/honnef.co/go/tools
package staticchecks

import (
	"golang.org/x/tools/go/analysis"
	"honnef.co/go/tools/quickfix"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
)

// Analyzers returns list of all staticcheck analyzers.
// func Analyzers() []*analysis.Analyzer {
// 	return nil
// }

// Staticcheck - the SA category of checks, codenamed staticcheck, contains all
// checks that are concerned with the correctness of code.
//
// https://staticcheck.dev/docs/checks/#SA
func Staticcheck(exclude ...string) []*analysis.Analyzer {
	skipper := newSkipper(exclude...)

	s := len(staticcheck.Analyzers) - skipper.Size()
	if s < 0 {
		s = 0
	}

	r := make([]*analysis.Analyzer, 0, s)

	for _, alz := range staticcheck.Analyzers {
		if skipper.Check(alz.Analyzer.Name) {
			continue
		}

		r = append(r, alz.Analyzer)
	}

	return r
}

// Simple - the S category of checks, codenamed simple, contains all checks that
// are concerned with simplifying code.
//
// https://staticcheck.dev/docs/checks/#SA
func Simple(exclude ...string) []*analysis.Analyzer {
	skipper := newSkipper(exclude...)

	s := len(simple.Analyzers) - skipper.Size()
	if s < 0 {
		s = 0
	}

	r := make([]*analysis.Analyzer, 0, s)

	for _, alz := range simple.Analyzers {
		if skipper.Check(alz.Analyzer.Name) {
			continue
		}

		r = append(r, alz.Analyzer)
	}

	return r
}

// Stylecheck - the ST category of checks, codenamed stylecheck, contains all
// checks that are concerned with stylistic issues.
//
// https://staticcheck.dev/docs/checks/#ST
func Stylecheck(exclude ...string) []*analysis.Analyzer {
	skipper := newSkipper(exclude...)

	s := len(stylecheck.Analyzers) - skipper.Size()
	if s < 0 {
		s = 0
	}

	r := make([]*analysis.Analyzer, 0, s)

	for _, alz := range stylecheck.Analyzers {
		if skipper.Check(alz.Analyzer.Name) {
			continue
		}

		r = append(r, alz.Analyzer)
	}

	return r
}

// Quickfix - the QF category of checks, codenamed quickfix, contains checks
// that are used as part of gopls for automatic refactorings. In the context of
// gopls, diagnostics of these checks will usually show up as hints, sometimes
// as information-level diagnostics.
//
// https://staticcheck.dev/docs/checks/#QF
func Quickfix(exclude ...string) []*analysis.Analyzer {
	skipper := newSkipper(exclude...)

	s := len(quickfix.Analyzers) - skipper.Size()
	if s < 0 {
		s = 0
	}

	r := make([]*analysis.Analyzer, 0, s)

	for _, alz := range quickfix.Analyzers {
		if skipper.Check(alz.Analyzer.Name) {
			continue
		}

		r = append(r, alz.Analyzer)
	}

	return r
}

type skipper struct {
	exclude map[string]struct{}
}

func newSkipper(exclude ...string) (s skipper) {
	if len(exclude) == 0 {
		return
	}

	s.exclude = make(map[string]struct{}, len(exclude))
	for _, e := range exclude {
		s.exclude[e] = struct{}{}
	}

	return
}

func (r skipper) Check(s string) bool {
	_, ok := r.exclude[s]

	return ok
}

func (r skipper) Size() int {
	return len(r.exclude)
}

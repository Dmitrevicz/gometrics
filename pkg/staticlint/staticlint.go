// Package staticlint represents custom static check linter.
//
// Increment-19.
//
// Создайте свой multichecker, состоящий из:
//   - стандартных статических анализаторов пакета golang.org/x/tools/go/analysis/passes;
//   - всех анализаторов класса SA пакета staticcheck.io;
//   - не менее одного анализатора остальных классов пакета staticcheck.io;
//   - двух или более любых публичных анализаторов на ваш выбор;
//   - собственный анализатор, запрещающий использовать прямой вызов os.Exit
//     в функции main пакета main.
package staticlint

import (
	"github.com/Dmitrevicz/gometrics/pkg/staticlint/noosexit"
	"github.com/Dmitrevicz/gometrics/pkg/staticlint/staticchecks"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/defers"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/timeformat"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
)

// Mount builds custom multichecker.
func Mount() {
	analyzers := []*analysis.Analyzer{
		// my custom analyzer
		noosexit.Analyzer,

		// passes analyzers
		defers.Analyzer,
		printf.Analyzer,
		shadow.Analyzer,
		structtag.Analyzer,
		tests.Analyzer,
		timeformat.Analyzer,
		unmarshal.Analyzer,

		// 3rd-party analyzers
		// ...
	}

	// staticcheck.io analyzers
	analyzers = append(analyzers, staticchecks.Staticcheck()...)
	analyzers = append(analyzers, staticchecks.Simple()...)
	analyzers = append(analyzers, staticchecks.Stylecheck()...)
	analyzers = append(analyzers, staticchecks.Quickfix("QF1001")...)

	// 3rd-party analyzers
	// TODO:
	// ...

	multichecker.Main(analyzers...)
}

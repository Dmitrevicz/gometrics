// Package staticlint represents custom static check linter.
package staticlint

import (
	"github.com/Dmitrevicz/gometrics/pkg/staticlint/noosexit"
	"golang.org/x/tools/go/analysis/multichecker"
)

// Mount builds custom multichecker.
func Mount() {
	multichecker.Main(
		noosexit.Analyzer,
	)
}

// Package noosexit defines an analyzer that forbids os.Exit calls in main.
//
// Increment-19 task:
//
// > Напишите и добавьте в multichecker собственный анализатор, запрещающий
// использовать прямой вызов os.Exit в функции main пакета main. При
// необходимости перепишите код своего проекта так, чтобы он удовлетворял
// данному анализатору.
package noosexit

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// Analyzer definition. Checks for os.Exit calls in main.
var Analyzer = &analysis.Analyzer{
	Name: "noosexitmain",
	Doc:  "check for os.Exit() calls in main",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		if file.Name.Name != "main" {
			// don't inspect packages other than main
			continue
		}

		// XXX: might also make use of file.Decls instead of walking over the
		// whole file tree
		ast.Inspect(file, func(node ast.Node) bool {
			switch x := node.(type) {
			case *ast.FuncDecl:
				if x.Name.Name != "main" {
					// don't inspect functions other than main()
					return false
				}

				for _, stmt := range x.Body.List {
					if exprStmt, ok := stmt.(*ast.ExprStmt); ok {
						if callExpr, ok := exprStmt.X.(*ast.CallExpr); ok {
							if sel, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
								if sel.Sel.Name == "Exit" && len(callExpr.Args) == 1 {
									if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == "os" {
										pass.Reportf(sel.Pos(), "os.Exit must not be called from main")
									}
								}
							}
						}
					}
				}
			}

			return true
		})
	}

	return nil, nil
}

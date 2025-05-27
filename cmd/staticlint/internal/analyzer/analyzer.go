// Package analyzer provides an analysis-compatible analyzer.
package analyzer

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// ExitAnalyzer checks for os.Exit() calls in main().
var ExitAnalyzer = &analysis.Analyzer{
	Name: "exit",
	Doc:  "check for os.Exit() calls in main()",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		if file.Name == nil {
			continue
		}

		if file.Name.Name != "main" {
			continue
		}

		ast.Inspect(file, func(node ast.Node) bool {
			switch x := node.(type) {
			case *ast.FuncDecl:
				return x.Name.Name == "main"
			case *ast.CallExpr:
				if s, ok := x.Fun.(*ast.SelectorExpr); ok {
					ident, ok := s.X.(*ast.Ident)
					if ok && ident.Name == "os" && s.Sel.Name == "Exit" {
						pass.Reportf(x.Pos(), "os.Exit called directly in main func of the main package")
					}
				}

			}
			return true
		})
	}
	return nil, nil
}

package analyzer

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
)

var analyzer = &analysis.Analyzer{
	Name: "function-file-matcher",
	Doc:  "Check that every package has a file for every public function.",
	Run:  run,
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
	},
}

func run(pass *analysis.Pass) (interface{}, error) {
	filenames := make(map[string]bool)
	functionVisitor := &functionVisitor{pass: pass, filenames: filenames}
	ast.Walk(functionVisitor, pass.Files[0])

	for _, file := range pass.Files {
		filename := pass.Fset.File(file.Pos()).Name()
		if _, ok := filenames[filename]; !ok {
			pass.Reportf(file.Pos(), "Package is missing file for function in %s", filename)
		}
	}

	return nil, nil
}

type functionVisitor struct {
	pass      *analysis.Pass
	filenames map[string]bool
}

func (v *functionVisitor) Visit(node ast.Node) ast.Visitor {
	switch node := node.(type) {
	case *ast.FuncDecl:
		if !node.Name.IsExported() {
			return nil
		}

		filename := v.pass.Fset.File(node.Pos()).Name()
		if _, ok := v.filenames[filename]; !ok {
			v.filenames[filename] = true
		}
	}

	return v
}

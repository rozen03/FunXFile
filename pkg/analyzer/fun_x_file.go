package analyzer

import (
	"fmt"
	"go/ast"
	"path/filepath"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
)

var Analyzer = &analysis.Analyzer{
	Name: "Fun_x_File",
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

	var suggestions []analysis.SuggestedFix
	for _, file := range pass.Files {
		filename := pass.Fset.File(file.Pos()).Name()
		if _, ok := filenames[filename]; ok {
			suggestion := analysis.SuggestedFix{
				Message: fmt.Sprintf("Create file %s to add function", getFileName(filename, pass.Pkg.Name())),
				TextEdits: []analysis.TextEdit{
					{
						Pos:     file.Pos(),
						End:     file.Pos(),
						NewText: []byte(fmt.Sprintf("package %s\\n\\nfunc MyFunction() {\\n}\\n\\n", pass.Pkg.Name())),
					},
				},
			}
			suggestions = append(suggestions, suggestion)
		}
	}

	if len(suggestions) > 0 {
		pass.Report(analysis.Diagnostic{
			Pos:            pass.Files[0].Pos(),
			Message:        "Package is missing file for function",
			SuggestedFixes: suggestions,
		})
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
		} else {
			v.filenames[filename] = false
		}
	}

	return v
}

func getFileName(filename, packageName string) string {
	return filepath.Join(filepath.Dir(filename), packageName+".go")
}

func (v *functionVisitor) Vosot(node ast.Node) ast.Visitor {
	return v
}

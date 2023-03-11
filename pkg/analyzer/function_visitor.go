package analyzer

import (
	"fmt"
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
)

type functionVisitor struct {
	pass      *analysis.Pass
	fileDecls map[string][]*ast.FuncDecl // map of filename to []*ast.FuncDecl
}

func (v *functionVisitor) Visit(node ast.Node) ast.Visitor {
	switch node := node.(type) {
	case *ast.FuncDecl:
		if !node.Name.IsExported() {
			return nil
		}

		filename := v.pass.Fset.File(node.Pos()).Name()
		funcNames := v.getFuncNames(filename)
		if len(funcNames) > 1 {
			v.reportError(filename, nil, funcNames)
		}
	}

	return v
}

func (v *functionVisitor) getFuncNames(filename string) []string {
	var funcNames []string

	for _, decl := range v.fileDecls[filename] {
		if decl.Name.IsExported() {
			funcNames = append(funcNames, decl.Name.Name)
		}
	}

	return funcNames
}

func (v *functionVisitor) reportError(filename string, suggestion *analysis.SuggestedFix, funcNames []string) {
	suggestions := []analysis.SuggestedFix{}
	if suggestion != nil {
		suggestions = append(suggestions, *suggestion)
	}

	message := fmt.Sprintf("Package %s has more than one public function in file %s. Public functions in this file: %s", v.pass.Pkg.Name(), filename, strings.Join(funcNames, ", "))
	v.pass.Report(analysis.Diagnostic{
		Pos:            v.pass.Files[0].Pos(),
		Message:        message,
		SuggestedFixes: suggestions,
	})
}

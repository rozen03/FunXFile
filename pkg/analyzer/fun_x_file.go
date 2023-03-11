package analyzer

import (
	"fmt"
	"go/ast"
	"reflect"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
)

var Analyzer = &analysis.Analyzer{
	Name:       "fin_x_file",
	Doc:        "Check that every package has a file for every public function.",
	Run:        run,
	ResultType: reflect.TypeOf(""),
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
	},
}

func run(pass *analysis.Pass) (interface{}, error) {
	fileDecls := make(map[string][]*ast.FuncDecl) // map of filename to []*ast.FuncDecl
	functionVisitor := &functionVisitor{pass: pass, fileDecls: fileDecls}

	// Retrieve all public function declarations from the package's files and store them in a map
	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			if f, ok := decl.(*ast.FuncDecl); ok {
				if f.Name.IsExported() {
					filename := pass.Fset.File(f.Pos()).Name()
					if fileDecls[filename] == nil {
						fileDecls[filename] = []*ast.FuncDecl{}
					}
					fileDecls[filename] = append(fileDecls[filename], f)
				}
			}
		}
	}

	ast.Walk(functionVisitor, pass.Files[0])

	var possibleFileNames []string
	for _, funcs := range fileDecls {
		for _, f := range funcs {
			if f == nil {
				continue
			}

			possibleFileNames = append(possibleFileNames, fmt.Sprintf("%s.go", f.Name.Name))
		}
	}

	// Write a list of possible file names to the document
	output := "Possible file names:\\\\\\\\n"
	for _, name := range possibleFileNames {
		output += fmt.Sprintf("- %s\\\\\\\\n", name)
	}

	return output, nil
}

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

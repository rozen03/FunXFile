package analyzer

import (
	"fmt"
	"go/ast"
	"reflect"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
)

func GetAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name:       "fin_x_file",
		Doc:        "Check that every package has a file for every public function.",
		Run:        run,
		ResultType: reflect.TypeOf(""),
		Requires: []*analysis.Analyzer{
			inspect.Analyzer,
		},
	}
}

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

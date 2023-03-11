package analyzer

import (
	"reflect"

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

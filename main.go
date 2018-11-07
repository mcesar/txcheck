package main

import (
	"fmt"
	"os"

	"golang.org/x/tools/go/callgraph"
	"golang.org/x/tools/go/callgraph/cha"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/ssa/ssautil"
)

const (
	usage = `txcheck: check if DML statements are inside a transaction.

Usage:

	txcheck package...
`
	errMsg = "function '%s' calls DML method but does not call Begin"
)

func main() {
	var warnings []string
	for _, filename := range os.Args[1:] {
		w, err := checkTx(os.Args[1:]...)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error checking '%v': %v\n", filename, err)
		}
		warnings = append(warnings, w...)
	}
	for _, w := range warnings {
		fmt.Println(w)
	}
}

func checkTx(args ...string) ([]string, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf(usage)
	}
	cfg := &packages.Config{Mode: packages.LoadAllSyntax}
	initial, err := packages.Load(cfg, args...)
	if err != nil {
		return nil, fmt.Errorf("could not load packages: %v", err)
	}
	if packages.PrintErrors(initial) > 0 {
		return nil, fmt.Errorf("packages contain errors")
	}
	// Create and build SSA-form program representation.
	prog, _ /*pkgs*/ := ssautil.AllPackages(initial, 0)
	prog.Build()
	cg := cha.CallGraph(prog)
	cg.DeleteSyntheticNodes()
	callersOfDML := make(map[string]bool)
	callersOfBegin := make(map[string]bool)
	dml := []string{"InsertInto", "Update", "DeleteFrom"}
	if err := callgraph.GraphVisitEdges(cg, func(edge *callgraph.Edge) error {
		pp := edge.Caller.Func.Package().Pkg.Path()
		if pp == "command-line-arguments" || contains(args, pp) {
			qualifiedName := fmt.Sprintf("%v.%v", pp, edge.Caller.Func.Name())
			if edge.Callee.Func.Name() == "Begin" {
				callersOfBegin[qualifiedName] = true
			} else if contains(dml, edge.Callee.Func.Name()) {
				callersOfDML[qualifiedName] = true
			}
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("could not visit edges: %v", err)
	}
	var warnings []string
	for function := range callersOfDML {
		if !callersOfBegin[function] {
			warnings = append(
				warnings,
				fmt.Sprintf(errMsg, function),
			)
		}
	}
	return warnings, nil
}

func contains(ss []string, s string) bool {
	for _, each := range ss {
		if each == s {
			return true
		}
	}
	return false
}

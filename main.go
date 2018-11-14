package main

import (
	"fmt"
	"io"
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
	warningMsg = "function '%s' calls DML method but does not call Begin\n"
	errMsg     = "error running checker: %v\n"
)

var (
	out    io.Writer = os.Stdout
	errout io.Writer = os.Stderr
	begin            = []string{"Begin", "BeginTx"}
	dml              = []string{
		"InsertInto",
		"Update",
		"DeleteFrom",
		"Exec",
		"ExecContext",
	}
)

func main() {
	warnings, err := (&checker{}).run(os.Args[1:]...)
	if err != nil {
		fmt.Fprintf(errout, errMsg, err)
	}
	for _, w := range warnings {
		fmt.Fprint(out, w)
	}
}

type checker struct {
	callersOfDML   map[string]bool
	callersOfBegin map[string]bool
	callers        map[string][]string
}

func (c *checker) run(args ...string) ([]string, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf(usage)
	}
	pkgs, err := c.loadPackages(args)
	if err != nil {
		return nil, fmt.Errorf("could not load packages: %v", err)
	}
	cg, err := c.computeCallGraph(pkgs)
	if err != nil {
		return nil, fmt.Errorf("could not compute call graph: %v", err)
	}
	err = c.analyzeGraph(cg, pkgs)
	if err != nil {
		return nil, fmt.Errorf("could not analyze call graph: %v", err)
	}
	return c.warnings(), nil
}

func (c *checker) loadPackages(args []string) ([]*packages.Package, error) {
	cfg := &packages.Config{Mode: packages.LoadAllSyntax}
	pkgs, err := packages.Load(cfg, args...)
	if err != nil {
		return nil, err
	}
	if packages.PrintErrors(pkgs) > 0 {
		return nil, fmt.Errorf("packages contain errors")
	}
	return pkgs, nil
}

func (c *checker) computeCallGraph(initial []*packages.Package) (cg *callgraph.Graph, err error) {
	// Create and build SSA-form program representation.
	prog, _ := ssautil.AllPackages(initial, 0)
	prog.Build()
	cg = cha.CallGraph(prog)
	cg.DeleteSyntheticNodes()
	return cg, nil
}

func (c *checker) analyzeGraph(cg *callgraph.Graph, pkgs []*packages.Package) error {
	var initialPackages []string
	for _, pkg := range pkgs {
		initialPackages = append(initialPackages, pkg.PkgPath)
	}
	qualifiedName := func(p, f string) string {
		return fmt.Sprintf("%v.%v", p, f)
	}
	isInitialPackage := func(p string) bool {
		return p == "command-line-arguments" ||
			p == "github.com/mcesar/dbrx" ||
			contains(initialPackages, p)
	}
	c.callersOfDML = make(map[string]bool)
	c.callersOfBegin = make(map[string]bool)
	c.callers = make(map[string][]string)
	err := callgraph.GraphVisitEdges(cg, func(edge *callgraph.Edge) error {
		callerpp := edge.Caller.Func.Package().Pkg.Path()
		callerqn := qualifiedName(callerpp, edge.Caller.Func.Name())
		if isInitialPackage(callerpp) {
			if contains(begin, edge.Callee.Func.Name()) {
				c.callersOfBegin[callerqn] = true
			} else if contains(dml, edge.Callee.Func.Name()) {
				c.callersOfDML[callerqn] = true
			}
		}
		calleepp := edge.Callee.Func.Package().Pkg.Path()
		if isInitialPackage(calleepp) {
			calleeqn := qualifiedName(calleepp, edge.Callee.Func.Name())
			c.callers[calleeqn] = append(c.callers[calleeqn], callerqn)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("could not visit edges: %v", err)
	}
	return nil
}

func (c *checker) warnings() []string {
	var warnings []string
	for function := range c.callersOfDML {
		if !c.isBeginCalledBy(function, map[string]bool{}) {
			warnings = append(warnings, fmt.Sprintf(warningMsg, function))
		}
	}
	return warnings
}

func (c *checker) isBeginCalledBy(f string, visited map[string]bool) bool {
	visited[f] = true
	if c.callersOfBegin[f] {
		return true
	}
	if len(c.callers[f]) == 0 {
		return false
	}
	for _, caller := range c.callers[f] {
		if !visited[caller] && !c.isBeginCalledBy(caller, visited) {
			return false
		}
	}
	return true
}

func contains(ss []string, s string) bool {
	for _, each := range ss {
		if each == s {
			return true
		}
	}
	return false
}

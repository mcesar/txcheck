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
	errMsg     = "error checking '%v': %v\n"
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
	var warnings []string
	for _, filename := range os.Args[1:] {
		w, err := (&checker{}).run(os.Args[1:]...)
		if err != nil {
			fmt.Fprintf(errout, errMsg, filename, err)
		}
		warnings = append(warnings, w...)
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
	cg, err := c.computeCallGraph(args)
	if err != nil {
		return nil, fmt.Errorf("could not compute call graph: %v", err)
	}
	err = c.analyzeGraph(cg, args)
	if err != nil {
		return nil, fmt.Errorf("could not analyze call graph: %v", err)
	}
	return c.warnings(), nil
}

func (c *checker) computeCallGraph(args []string) (*callgraph.Graph, error) {
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
	return cg, nil
}

func (c *checker) analyzeGraph(cg *callgraph.Graph, args []string) error {
	qualifiedName := func(p, f string) string {
		return fmt.Sprintf("%v.%v", p, f)
	}
	ownpackage := func(p string) bool {
		return p == "command-line-arguments" || contains(args, p)
	}
	c.callersOfDML = make(map[string]bool)
	c.callersOfBegin = make(map[string]bool)
	c.callers = make(map[string][]string)
	err := callgraph.GraphVisitEdges(cg, func(edge *callgraph.Edge) error {
		callerpp := edge.Caller.Func.Package().Pkg.Path()
		callerqn := qualifiedName(callerpp, edge.Caller.Func.Name())
		if ownpackage(callerpp) {
			if contains(begin, edge.Callee.Func.Name()) {
				c.callersOfBegin[callerqn] = true
			} else if contains(dml, edge.Callee.Func.Name()) {
				c.callersOfDML[callerqn] = true
			}
		}
		calleepp := edge.Callee.Func.Package().Pkg.Path()
		if ownpackage(calleepp) {
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
		if !c.isBeginCalledBy(function) {
			warnings = append(warnings, fmt.Sprintf(warningMsg, function))
		}
	}
	return warnings
}

func (c *checker) isBeginCalledBy(f string) bool {
	if c.callersOfBegin[f] {
		return true
	}
	if len(c.callers[f]) == 0 {
		return false
	}
	for _, caller := range c.callers[f] {
		if !c.isBeginCalledBy(caller) {
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

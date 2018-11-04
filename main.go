package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
)

const errMsg = "function '%s' calls method %s but does not call Begin"

func main() {
	var warnings []string
	for _, filename := range os.Args[1:] {
		w, err := checkTx(filename, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error checking '%v': %v\n", filename, err)
		}
		warnings = append(warnings, w...)
	}
	for _, w := range warnings {
		fmt.Println(w)
	}
}

func checkTx(filename string, src interface{}) ([]string, error) {
	fs := token.NewFileSet()
	f, err := parser.ParseFile(fs, filename, src, parser.AllErrors)
	if err != nil {
		return nil, fmt.Errorf("could not parse: %v", err)
	}
	v := &fileVisitor{}
	ast.Walk(v, f)
	var warnings []string
	for _, fdv := range v.funcDeclVisitors {
		if (fdv.hasInsertInto || fdv.hasUpdate || fdv.hasDeleteFrom) && !fdv.hasBegin {
			var method string
			if fdv.hasInsertInto {
				method = "InsertInto"
			} else if fdv.hasUpdate {
				method = "Update"
			} else if fdv.hasDeleteFrom {
				method = "DeleteFrom"
			}
			warnings = append(
				warnings,
				fmt.Sprintf(errMsg, fdv.funcName, method),
			)
		}
	}
	return warnings, nil
}

type fileVisitor struct {
	funcDeclVisitors []*funcDeclVisitor
}

func (v *fileVisitor) Visit(n ast.Node) ast.Visitor {
	if node, ok := n.(*ast.FuncDecl); ok {
		fdv := &funcDeclVisitor{funcName: node.Name.Name}
		v.funcDeclVisitors = append(v.funcDeclVisitors, fdv)
		return fdv
	}
	return v
}

type funcDeclVisitor struct {
	funcName                                          string
	hasInsertInto, hasUpdate, hasDeleteFrom, hasBegin bool
}

func (v *funcDeclVisitor) Visit(n ast.Node) ast.Visitor {
	if node, ok := n.(*ast.CallExpr); ok {
		if sel, ok := node.Fun.(*ast.SelectorExpr); ok {
			switch sel.Sel.Name {
			case "InsertInto":
				v.hasInsertInto = true
			case "Update":
				v.hasUpdate = true
			case "DeleteFrom":
				v.hasDeleteFrom = true
			case "Begin":
				v.hasBegin = true
			}
		}
	}
	return v
}

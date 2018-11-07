package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestCheck(t *testing.T) {
	/*
		gopath, err := filepath.Abs("testdata")
		if err != nil {
			t.Fatal(err)
		}
	*/
	cases := []struct {
		name   string
		input  string
		output []string
		err    error
	}{
		{
			name: `Calls InsertInto method but does not call Begin,
				must return warning`,
			input: `
				package main
				import "github.com/gocraft/dbr"
				func main() {
					conn, _ := dbr.Open("sqlite", ":memory:", nil)
					sess := conn.NewSession(nil)
					sess.InsertInto("t").Columns("c").Values("v").Exec()
				}
			`,
			output: []string{fmt.Sprintf(errMsg, "command-line-arguments.main")},
			err:    nil,
		},
		{
			name: `Calls InsertInto and Begin methods, must not return warning`,
			input: `
				package main
				import "github.com/gocraft/dbr"
				func main() {
					conn, _ := dbr.Open("sqlite", ":memory:", nil)
					sess := conn.NewSession(nil)
					tx, _ := sess.Begin()
					tx.InsertInto("t").Columns("c").Values("v").Exec()
				}
			`,
			output: nil,
			err:    nil,
		},
		{
			name: `Calls Update method but does not call Begin,
				must return warning`,
			input: `
				package main
				import "github.com/gocraft/dbr"
				func main() {
					conn, _ := dbr.Open("sqlite", ":memory:", nil)
					sess := conn.NewSession(nil)
					sess.Update("t").Set("name", "n").Exec()
				}
			`,
			output: []string{fmt.Sprintf(errMsg, "command-line-arguments.main")},
			err:    nil,
		},
		{
			name: `Calls DeleteFrom method but does not call Begin,
				must return warning`,
			input: `
				package main
				import "github.com/gocraft/dbr"
				func main() {
					conn, _ := dbr.Open("sqlite", ":memory:", nil)
					sess := conn.NewSession(nil)
					sess.DeleteFrom("t").Exec()
				}
			`,
			output: []string{fmt.Sprintf(errMsg, "command-line-arguments.main")},
			err:    nil,
		},
		{
			name: `Parse error, must return error`,
			input: `
				package main
				import "github.com/gocraft/dbr"
				func main() {
					dbr.Open("sqlite", ":memory:", nil
				}
			`,
			output: nil,
			err:    fmt.Errorf("packages contain errors"),
		},
	}
	for _, c := range cases {
		dir := filepath.Join(os.TempDir(), "txcheck_test")
		err := os.Mkdir(dir, os.ModePerm)
		if err != nil && !os.IsExist(err) {
			t.Fatal(err)
		}
		file := filepath.Join(dir, "main.go")
		err = ioutil.WriteFile(file, []byte(c.input), os.ModePerm)
		if err != nil {
			t.Fatal(err)
		}
		output, err := checkTx(file)
		if !reflect.DeepEqual(c.output, output) {
			t.Errorf("Want %v, got %v", c.output, output)
		}
		if !reflect.DeepEqual(c.err, err) {
			t.Errorf("Want %v, got %v", c.err, err)
		}
	}
}

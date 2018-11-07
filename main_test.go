package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestCheckerRun(t *testing.T) {
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
			err:    fmt.Errorf("could not compute call graph: packages contain errors"),
		},
		{
			name: `Calls InsertInto at one function and Begin at the caller,
				must not return warning`,
			input: `
				package main
				import "github.com/gocraft/dbr"
				func main() {
					conn, _ := dbr.Open("sqlite", ":memory:", nil)
					sess := conn.NewSession(nil)
					tx, _ := sess.Begin()
					f1(tx)
				}
				func f1(tx *dbr.Tx) {
					tx.InsertInto("t").Columns("c").Values("v").Exec()
				}
			`,
			output: nil,
			err:    nil,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
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
			output, err := (&checker{}).run(file)
			if !reflect.DeepEqual(c.output, output) {
				t.Errorf("Want %v, got %v", c.output, output)
			}
			if !reflect.DeepEqual(c.err, err) {
				t.Errorf("Want %v, got %v", c.err, err)
			}
		})
	}
}

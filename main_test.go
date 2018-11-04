package main

import (
	"fmt"
	"reflect"
	"testing"
)

func TestCheck(t *testing.T) {
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
			output: []string{fmt.Sprintf(errMsg, "main", "InsertInto")},
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
			output: []string{fmt.Sprintf(errMsg, "main", "Update")},
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
			output: []string{fmt.Sprintf(errMsg, "main", "DeleteFrom")},
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
			err: fmt.Errorf("could not parse: 5:40: missing ',' " +
				"before newline in argument list (and 6 more errors)"),
		},
	}
	for _, c := range cases {
		output, err := checkTx("", c.input)
		if !reflect.DeepEqual(c.output, output) {
			t.Errorf("Want %v, got %v", c.output, output)
		}
		if !reflect.DeepEqual(c.err, err) {
			t.Errorf("Want %v, got %v", c.err, err)
		}
	}
}

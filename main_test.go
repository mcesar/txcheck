package main

import (
	"bytes"
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
		output string
		err    string
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
			output: fmt.Sprintf(warningMsg, "command-line-arguments.main"),
			err:    "",
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
			output: "",
			err:    "",
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
			output: fmt.Sprintf(warningMsg, "command-line-arguments.main"),
			err:    "",
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
			output: fmt.Sprintf(warningMsg, "command-line-arguments.main"),
			err:    "",
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
			output: "",
			err: fmt.Sprintf(
				errMsg,
				filepath.Join(os.TempDir(), "txcheck_test", "main.go"),
				"could not compute call graph: packages contain errors",
			),
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
			output: "",
			err:    "",
		},
		{
			name: `Calls Exec method but does not call Begin,
				must return warning`,
			input: `
				package main
				import "database/sql"
				var db *sql.DB
				func main() {
					db.Exec("INSERT INTO t(c) values('v');")
				}
			`,
			output: fmt.Sprintf(warningMsg, "command-line-arguments.main"),
			err:    "",
		},
		{
			name: `Calls ExecContext method but does not call Begin,
				must return warning`,
			input: `
				package main
				import (
					"context"
					"database/sql"
				)
				var db *sql.DB
				func main() {
					db.ExecContext(
						context.Background(),
						"INSERT INTO t(c) values('v');",
					)
				}
			`,
			output: fmt.Sprintf(warningMsg, "command-line-arguments.main"),
			err:    "",
		},
		{
			name: `Calls Exec method and BeginTx, must not return warning`,
			input: `
				package main
				import (
					"context"
					"database/sql"
				)
				var db *sql.DB
				func main() {
					tx, _ := db.BeginTx(context.Background(), nil)
					tx.Exec("INSERT INTO t(c) values('v');")
				}
			`,
			output: "",
			err:    "",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			oldArgs := os.Args
			defer func() { os.Args = oldArgs }()
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
			os.Args = []string{"cmd", file}
			buf := new(bytes.Buffer)
			errbuf := new(bytes.Buffer)
			out = buf
			errout = errbuf
			main()
			if !reflect.DeepEqual(c.output, buf.String()) {
				t.Errorf("Want %v, got %v", c.output, buf.String())
			}
			if !reflect.DeepEqual(c.err, errbuf.String()) {
				t.Errorf("Want %v, got %v", c.err, errbuf.String())
			}
		})
	}
}

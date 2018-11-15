# txcheck
txcheck is a program for checking that you call Begin whenever you call DML functions in go programs. Supports `database/sql` and `github.com/gocraft/dbr`.

## Install
```
go get -u github.com/mcesar/txcheck
```
## Use

For basic usage, just give the package path of interest as the first argument:
```
txcheck github.com/mcesar/txcheck
```
## Check
Given the following program, `txcheck ` warns that function main calls `Exec` but does no call `Begin` (or `BeginTx`).
```go
package main
import (
  "context"
  "database/sql"
)
var db *sql.DB
func main() {
  // tx, _ := db.Begin()
  tx.Exec("INSERT INTO t(c) values('v');")
  // ...
}
```

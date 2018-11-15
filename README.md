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

module test

go 1.25.1

require (
	github.com/mikeschinkel/go-fsfix v0.1.0
	github.com/xmlui-org/xmlui-test-server/xmluisvr v0.0.0
)

require (
	github.com/lib/pq v1.10.9 // indirect
	github.com/mattn/go-sqlite3 v1.14.32 // indirect
)

replace github.com/xmlui-org/xmlui-test-server/xmluisvr => ../xmluisvr

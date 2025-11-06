module github.com/xmlui-org/localdev/xmluisvr

go 1.25.3

require (
	github.com/lib/pq v1.10.9
	github.com/mattn/go-sqlite3 v1.14.32
	github.com/mikeschinkel/go-cfgstore v0.0.0-00010101000000-000000000000
	github.com/mikeschinkel/go-cliutil v0.0.0-00010101000000-000000000000
	github.com/mikeschinkel/go-doterr v0.0.0-00010101000000-000000000000
	github.com/mikeschinkel/go-dt v0.0.0-00010101000000-000000000000
	github.com/mikeschinkel/go-dt/appinfo v0.0.0-00010101000000-000000000000
	github.com/mikeschinkel/go-dt/de v0.0.0-00010101000000-000000000000
	github.com/mikeschinkel/go-dt/dtx v0.0.0-00010101000000-000000000000
	github.com/mikeschinkel/go-fsfix v0.1.0
	github.com/mikeschinkel/go-jsontest v0.0.0-00010101000000-000000000000
	github.com/mikeschinkel/go-jsonxtractr v0.0.0-00010101000000-000000000000
	github.com/mikeschinkel/go-rfc9457 v0.0.0-00010101000000-000000000000
	github.com/stretchr/testify v1.11.1
)

require github.com/mikeschinkel/go-testutil v0.0.0-00010101000000-000000000000

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/tidwall/gjson v1.18.0 // indirect
	github.com/tidwall/match v1.2.0 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	github.com/mikeschinkel/go-cfgstore => ../../../go-pkgs/go-cfgstore
	github.com/mikeschinkel/go-cliutil => ../../../go-pkgs/go-cliutil
	github.com/mikeschinkel/go-doterr => ../../../go-pkgs/go-doterr
	github.com/mikeschinkel/go-dt => ../../../go-pkgs/go-dt
	github.com/mikeschinkel/go-dt/appinfo => ../../../go-pkgs/go-dt/appinfo
	github.com/mikeschinkel/go-dt/de => ../../../go-pkgs/go-dt/de
	github.com/mikeschinkel/go-dt/dtx => ../../../go-pkgs/go-dt/dtx
	github.com/mikeschinkel/go-fsfix => ../../../go-pkgs/go-fsfix
	github.com/mikeschinkel/go-jsontest => ../../../go-pkgs/go-jsontest
	github.com/mikeschinkel/go-jsonxtractr => ../../../go-pkgs/go-jsonxtractr
	github.com/mikeschinkel/go-rfc9457 => ../../../go-pkgs/go-rfc9457
	github.com/mikeschinkel/go-testutil => ../../../go-pkgs/go-testutil
)

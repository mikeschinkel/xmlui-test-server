module xmluisvr

go 1.25.3

require github.com/xmlui-org/localsvr/xmluisvr v0.0.0-00010101000000-000000000000

require (
	github.com/lib/pq v1.10.9 // indirect
	github.com/mattn/go-sqlite3 v1.14.32 // indirect
	github.com/mikeschinkel/go-cfgstore v0.0.0-00010101000000-000000000000 // indirect
	github.com/mikeschinkel/go-cliutil v0.0.0-20251105231813-8ce963ade5dd // indirect
	github.com/mikeschinkel/go-doterr v0.0.0-00010101000000-000000000000 // indirect
	github.com/mikeschinkel/go-dt v0.0.0-20251105233453-a7985f775567 // indirect
	github.com/mikeschinkel/go-dt/appinfo v0.0.0-20251106125543-42540c8e051a // indirect
	github.com/mikeschinkel/go-dt/de v0.0.0-20251105233453-a7985f775567 // indirect
	github.com/mikeschinkel/go-dt/dtx v0.0.0-20251107040413-53a1559d69c5 // indirect
	github.com/mikeschinkel/go-jsonxtractr v0.0.0-00010101000000-000000000000 // indirect
	github.com/mikeschinkel/go-rfc9457 v0.0.0-00010101000000-000000000000 // indirect
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
	github.com/xmlui-org/localsvr/xmluisvr => ../xmluisvr
)

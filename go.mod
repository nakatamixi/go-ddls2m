module github.com/nakatamixi/go-ddls2m

go 1.13

require (
	cloud.google.com/go/spanner v1.1.0
	github.com/juju/errors v0.0.0-20190930114154-d42613fe1ab9 // indirect
	github.com/knocknote/vitess-sqlparser v0.0.0-20190712090058-385243f72d33
	golang.org/x/xerrors v0.0.0-20190717185122-a985d3407aa7
)

replace github.com/knocknote/vitess-sqlparser => github.com/nakatamixi/vitess-sqlparser v0.0.0-20191030035102-acd30bb46a50

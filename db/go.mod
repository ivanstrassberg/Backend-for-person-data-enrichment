module db

go 1.23.0

toolchain go1.23.9

replace api => ../api

require (
	api v0.0.0-00010101000000-000000000000
	github.com/golang-migrate/migrate/v4 v4.18.3
	github.com/lib/pq v1.10.9
	github.com/mattes/migrate v3.0.1+incompatible
	github.com/pressly/goose/v3 v3.24.3
)

require (
	github.com/mfridman/interpolate v0.0.2 // indirect
	github.com/sethvargo/go-retry v0.3.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/sync v0.14.0 // indirect
)

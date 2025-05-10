module effective-mobile

go 1.23.0

toolchain go1.23.9

replace api => ./api

replace db => ./db

require (
	api v0.0.0-00010101000000-000000000000
	db v0.0.0-00010101000000-000000000000
	github.com/golang-migrate/migrate/v4 v4.18.3
	github.com/joho/godotenv v1.5.1
	github.com/lib/pq v1.10.9
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/go-sql-driver/mysql v1.9.2 // indirect
	github.com/mattes/migrate v3.0.1+incompatible // indirect
	github.com/mfridman/interpolate v0.0.2 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pressly/goose v2.7.0+incompatible // indirect
	github.com/pressly/goose/v3 v3.24.3 // indirect
	github.com/sethvargo/go-retry v0.3.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/sync v0.14.0 // indirect
)

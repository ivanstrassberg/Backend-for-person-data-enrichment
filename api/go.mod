module api

replace db => ../db

go 1.23.0

toolchain go1.23.9

require db v0.0.0-00010101000000-000000000000

require (
	github.com/golang-migrate/migrate v3.5.4+incompatible // indirect
	github.com/golang-migrate/migrate/v4 v4.18.3 // indirect
	github.com/lib/pq v1.10.9 // indirect
)

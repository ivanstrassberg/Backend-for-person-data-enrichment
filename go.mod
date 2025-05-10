module effective-mobile

go 1.23.9

replace api => ./api

replace db => ./db

replace docs => ./docs

require (
	api v0.0.0-00010101000000-000000000000
	db v0.0.0-00010101000000-000000000000c
	docs v0.0.0-00010101000000-000000000000
	github.com/golang-migrate/migrate/v4 v4.18.3
	github.com/joho/godotenv v1.5.1
	github.com/lib/pq v1.10.9
)

require (
	github.com/KyleBanks/depth v1.2.1 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/jsonreference v0.20.0 // indirect
	github.com/go-openapi/spec v0.20.6 // indirect
	github.com/go-openapi/swag v0.19.15 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/kr/pretty v0.3.0 // indirect
	github.com/mailru/easyjson v0.7.6 // indirect
	github.com/mattes/migrate v3.0.1+incompatible // indirect
	github.com/mfridman/interpolate v0.0.2 // indirect
	github.com/pressly/goose/v3 v3.24.3 // indirect
	github.com/sethvargo/go-retry v0.3.0 // indirect
	github.com/swaggo/files v0.0.0-20220610200504-28940afbdbfe // indirect
	github.com/swaggo/http-swagger v1.3.4 // indirect
	github.com/swaggo/swag v1.16.4 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/net v0.40.0 // indirect
	golang.org/x/sync v0.14.0 // indirect
	golang.org/x/tools v0.24.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

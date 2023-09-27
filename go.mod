module github.com/go-rel/sqlite3

go 1.21

toolchain go1.21.0

require (
	github.com/go-rel/rel v0.40.0
	github.com/go-rel/sql v0.15.1-0.20230927020931-5b67559d2fe1
	github.com/mattn/go-sqlite3 v1.14.17
	github.com/stretchr/testify v1.8.4
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/serenize/snaker v0.0.0-20201027110005-a7ad2135616e // indirect
	golang.org/x/text v0.4.0 // indirect
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/go-rel/sql v0.15.1-0.20230926233117-4af81443c5e1 => ../sql/

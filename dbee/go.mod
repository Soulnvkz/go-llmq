module sol/dbee

go 1.23.5

replace sol/proto => ../proto

require (
	github.com/lib/pq v1.10.9
	sol/proto v0.0.0-00010101000000-000000000000
)

require (
	github.com/golang-migrate/migrate/v4 v4.18.2
	github.com/joho/godotenv v1.5.1
)

require (
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	go.uber.org/atomic v1.7.0 // indirect
)

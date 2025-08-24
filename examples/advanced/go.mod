module github.com/azrod/kivigo/examples/advanced

go 1.24.5

replace (
	github.com/azrod/kivigo => ../../
	github.com/azrod/kivigo/backend/redis => ../../backend/redis
)

require (
	github.com/azrod/kivigo v1.2.0
	github.com/azrod/kivigo/backend/redis v0.0.0-00010101000000-000000000000
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/redis/go-redis/v9 v9.11.0 // indirect
	golang.org/x/sync v0.16.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

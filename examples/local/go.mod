module github.com/kivigo/kivigo/examples/local

go 1.24.5

replace (
	github.com/kivigo/kivigo => ../../
	github.com/kivigo/kivigo/backend/local => ../../backend/local
)

require (
	github.com/kivigo/kivigo v1.3.0
	github.com/kivigo/kivigo/backend/local v0.0.0-00010101000000-000000000000
)

require (
	github.com/pkg/errors v0.9.1 // indirect
	go.etcd.io/bbolt v1.4.3 // indirect
	golang.org/x/sync v0.16.0 // indirect
	golang.org/x/sys v0.29.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

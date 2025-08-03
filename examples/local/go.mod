module github.com/azrod/kivigo/examples/local

go 1.24.5

replace (
	github.com/azrod/kivigo => ../../
	github.com/azrod/kivigo/backend/local => ../../backend/local
)

require (
	github.com/azrod/kivigo v1.1.0
	github.com/azrod/kivigo/backend/local v0.0.0-00010101000000-000000000000
)

require (
	github.com/kr/text v0.2.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	go.etcd.io/bbolt v1.4.2 // indirect
	golang.org/x/sync v0.16.0 // indirect
	golang.org/x/sys v0.29.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

package testinfra

import "strings"

// Registry for backends
var backends = map[string]BackendInfra{}

func RegisterBackend(b BackendInfra) {
	backends[strings.ToLower(b.Name())] = b
}

func GetBackend(name string) BackendInfra {
	return backends[name]
}

func ListBackends() []string {
	names := make([]string, 0, len(backends))
	for k := range backends {
		names = append(names, k)
	}

	return names
}

package kivigo

import (
	"fmt"

	"github.com/azrod/kivigo/pkg/backend"
	"github.com/azrod/kivigo/pkg/client"
	"github.com/azrod/kivigo/pkg/encoder"
)

// New creates and returns a new KiviGo client instance using the provided backend and options.
//
// The [backend] parameter must implement the [backend.Backend] type and represents the storage backend (e.g. [backend.Local], [backend.Redis]).
// The [opts] parameter allows you to specify one or more [client.Options], such as the encoder to use for value serialization.
// The encoder must implement the [models.Encoder] interface.
//
// If no encoder is provided, [encoder.JSON] will be used by default.
//
// Example (basic):
//
//	backend := backend.Local(local.Option{Path: "./"})
//	client, err := kivigo.New(backend)
//	if err != nil {
//	    panic(err)
//	}
//	defer client.Close()
//
// Example (with YAML encoder):
//
//	backend := backend.Local(local.Option{Path: "./"})
//	client, err := kivigo.New(backend, func(opt client.Option) client.Option {
//	    opt.Encoder = encoder.YAML
//	    return opt
//	})
//	if err != nil {
//	    panic(err)
//	}
//	defer client.Close()
//
// Example (with Redis backend):
//
//	backend := backend.Redis(redis.Option{Addr: "localhost:6379"})
//	client, err := kivigo.New(backend)
//	if err != nil {
//	    panic(err)
//	}
//	defer client.Close()
//
// See [client.Client] for available methods and [models.KV] for backend interface details.
//
// Returns a [client.Client] and an error if initialization fails.
func New(backend backend.Backend, opts ...client.Options) (client.Client, error) {
	if backend == nil {
		return client.Client{}, fmt.Errorf("backend cannot be nil")
	}

	opt := client.Option{}
	for _, o := range opts {
		opt = o(opt)
	}

	if opt.Encoder == nil {
		opt.Encoder = encoder.JSON
	}

	return backend(opt)
}

package kivigo

import (
	"fmt"

	"github.com/kivigo/encoders/json"

	"github.com/kivigo/kivigo/pkg/client"
	"github.com/kivigo/kivigo/pkg/models"
)

// New creates and returns a new KiviGo client instance using the provided backend and options.
//
// The [backend] parameter must implement the [models.KV] interface and represents the storage backend (e.g. [backend/local], [backend/redis], or a custom backend).
// The [opts] parameter allows you to specify one or more [client.Options], such as the encoder to use for value serialization.
// The encoder must implement the [models.Encoder] interface.
//
// If no encoder is provided, [encoder.JSON] will be used by default.
//
// # Examples
//
// Basic usage with the local backend (BoltDB):
//
//		import (
//		    "github.com/kivigo/kivigo"
//		    "github.com/kivigo/backends/local"
//		)
//		func main() {
//		    kvStore, err := local.New(local.Option{Path: "./"})
//		    if err != nil {
//	         panic(err)
//	     }
//		    client, err := kivigo.New(kvStore)
//		    if err != nil {
//		        panic(err)
//		    }
//		    defer client.Close()
//
//		}
//
// Usage with Redis backend and YAML encoder:
//
//	import (
//	    "github.com/kivigo/kivigo"
//	    "github.com/kivigo/backends/redis"
//	    "github.com/kivigo/kivigo/pkg/client"
//	    "github.com/kivigo/encoders/yaml"
//	)
//	func main() {
//	    kvStore, err := redis.New(redis.Option{Addr: "localhost:6379"})
//	    if err != nil {
//	        panic(err)
//	    }
//	    client, err := kivigo.New(kvStore, func(opt client.Option) client.Option {
//	        opt.Encoder = yaml.New()
//	        return opt
//	    })
//	    if err != nil {
//	        panic(err)
//	    }
//	    defer client.Close()
//	    // ... use client ...
//	}
//
// See [client.Client] for available methods and [models.KV] for backend interface details.
//
// Returns a [client.Client] and an error if initialization fails.
func New(backend models.KV, opts ...client.Options) (client.Client, error) {
	if backend == nil {
		return client.Client{}, fmt.Errorf("backend cannot be nil")
	}

	opt := client.Option{}
	for _, o := range opts {
		opt = o(opt)
	}

	if opt.Encoder == nil {
		opt.Encoder = json.New()
	}

	return client.New(backend, opt)
}

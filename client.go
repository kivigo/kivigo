package kivigo

import (
	"fmt"

	"github.com/azrod/kivigo/pkg/backend"
	"github.com/azrod/kivigo/pkg/client"
	"github.com/azrod/kivigo/pkg/encoder"
)

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

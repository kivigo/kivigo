package backend

import (
	"fmt"

	"github.com/azrod/kivigo/pkg/backend/local"
	"github.com/azrod/kivigo/pkg/backend/redis"
	"github.com/azrod/kivigo/pkg/client"
	"github.com/azrod/kivigo/pkg/models"
)

type Backend func(Opts client.Option) (client.Client, error)

func Redis(opt redis.Option) Backend {
	return func(opts client.Option) (client.Client, error) {
		c, err := redis.New(opt, opts)
		if err != nil {
			return client.Client{}, err
		}

		return client.New(c, opts)
	}
}

func Local(opt local.Option) Backend {
	return func(opts client.Option) (client.Client, error) {
		c, err := local.New(opt, opts)
		if err != nil {
			return client.Client{}, err
		}

		return client.New(c, opts)
	}
}

func CustomBackend(kv models.KV) Backend {
	return func(opts client.Option) (client.Client, error) {
		if kv == nil {
			return client.Client{}, fmt.Errorf("kv cannot be nil")
		}

		return client.New(kv, opts)
	}
}

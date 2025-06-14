package client_test

import (
	"testing"

	"github.com/azrod/kivigo/pkg/client"
	"github.com/azrod/kivigo/pkg/encoder"
	"github.com/azrod/kivigo/pkg/mock"
)

func TestSetAndGet(t *testing.T) {
	mockKV := &mock.MockKV{Data: map[string][]byte{}}

	c, err := client.New(mockKV, client.Option{Encoder: encoder.JSON})
	if err != nil {
		t.Fatal(err)
	}

	key := "foo"
	value := "bar"

	if err := c.Set(t.Context(), key, value); err != nil {
		t.Fatal(err)
	}

	var got string
	if err := c.Get(t.Context(), key, &got); err != nil {
		t.Fatal(err)
	}

	if got != value {
		t.Errorf("expected %s, got %s", value, got)
	}
}

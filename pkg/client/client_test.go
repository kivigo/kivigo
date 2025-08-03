package client_test

import (
	"context"
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

	if err := c.Set(context.Background(), key, value); err != nil {
		t.Fatal(err)
	}

	var got string
	if err := c.Get(context.Background(), key, &got); err != nil {
		t.Fatal(err)
	}

	if got != value {
		t.Errorf("expected %s, got %s", value, got)
	}
}

func TestClientClose(t *testing.T) {
	mockKV := &mock.MockKV{Data: map[string][]byte{}}

	c, err := client.New(mockKV, client.Option{Encoder: encoder.JSON})
	if err != nil {
		t.Fatal(err)
	}

	if err := c.Close(); err != nil {
		t.Errorf("expected nil error on Close, got %v", err)
	}
}

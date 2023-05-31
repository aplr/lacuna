package pubsub

import (
	"context"
	"testing"
)

func TestNewPubSubReturnsClient(t *testing.T) {
	// arrange
	ctx := context.Background()

	// act
	_, err := NewPubSub(ctx, &Config{ProjectID: "test"})

	// assert
	if err != nil {
		t.Error(err)
	}
}

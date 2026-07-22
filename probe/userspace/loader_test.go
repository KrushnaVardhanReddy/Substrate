package userspace

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type MockLoader struct {
	events [][]byte
	idx    int
}

func (m *MockLoader) ReadEvent() ([]byte, error) {
	if m.idx >= len(m.events) {
		return nil, io.EOF
	}
	event := m.events[m.idx]
	m.idx++
	return event, nil
}

func TestStreamer(t *testing.T) {
	mockLoader := &MockLoader{
		events: [][]byte{
			[]byte("GET /api/v1/users HTTP/1.1\r\nHost: example.com\r\n\r\n"),
			[]byte("POST /api/v1/users HTTP/1.1\r\nHost: example.com\r\n\r\n"),
		},
	}

	streamer := &Streamer{Loader: mockLoader}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := streamer.Start(ctx)
	assert.ErrorIs(t, err, io.EOF)
}

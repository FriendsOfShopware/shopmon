package telemetry

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnsurePath(t *testing.T) {
	tests := []struct {
		name        string
		endpoint    string
		defaultPath string
		want        string
	}{
		{
			name:        "no path appends default",
			endpoint:    "http://localhost:4318",
			defaultPath: "/v1/logs",
			want:        "http://localhost:4318/v1/logs",
		},
		{
			name:        "root path appends default",
			endpoint:    "http://localhost:4318/",
			defaultPath: "/v1/logs",
			want:        "http://localhost:4318/v1/logs",
		},
		{
			name:        "existing path unchanged",
			endpoint:    "http://localhost:4318/custom",
			defaultPath: "/v1/logs",
			want:        "http://localhost:4318/custom",
		},
		{
			name:        "invalid url unchanged",
			endpoint:    "://not a url",
			defaultPath: "/v1/logs",
			want:        "://not a url",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ensurePath(tt.endpoint, tt.defaultPath))
		})
	}
}

func TestParseSamplerRatio(t *testing.T) {
	tests := []struct {
		name string
		set  bool
		env  string
		want float64
	}{
		{name: "unset defaults to 1.0", set: false, want: 1.0},
		{name: "empty defaults to 1.0", set: true, env: "", want: 1.0},
		{name: "half", set: true, env: "0.5", want: 0.5},
		{name: "one", set: true, env: "1.0", want: 1.0},
		{name: "zero", set: true, env: "0.0", want: 0.0},
		{name: "above one clamps", set: true, env: "2.5", want: 1.0},
		{name: "below zero clamps", set: true, env: "-1", want: 0.0},
		{name: "invalid defaults to 1.0", set: true, env: "not-a-number", want: 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.set {
				t.Setenv("OTEL_TRACES_SAMPLER_RATIO", tt.env)
			} else {
				require.NoError(t, os.Unsetenv("OTEL_TRACES_SAMPLER_RATIO"))
			}
			assert.Equal(t, tt.want, parseSamplerRatio())
		})
	}
}

type fakeHandler struct {
	calls   int
	enabled bool
	err     error
}

func (f *fakeHandler) Enabled(context.Context, slog.Level) bool { return f.enabled }

func (f *fakeHandler) Handle(context.Context, slog.Record) error {
	f.calls++
	return f.err
}

func (f *fakeHandler) WithAttrs([]slog.Attr) slog.Handler { return f }

func (f *fakeHandler) WithGroup(string) slog.Handler { return f }

func TestMultiHandlerHandle(t *testing.T) {
	t.Run("all handlers invoked when one errors", func(t *testing.T) {
		errFirst := errors.New("first failed")
		h1 := &fakeHandler{enabled: true, err: errFirst}
		h2 := &fakeHandler{enabled: true}

		mh := newMultiHandler(h1, h2)
		err := mh.Handle(context.Background(), slog.Record{})

		assert.Equal(t, 1, h1.calls)
		assert.Equal(t, 1, h2.calls)
		assert.ErrorIs(t, err, errFirst)
	})

	t.Run("joins multiple errors", func(t *testing.T) {
		err1 := errors.New("e1")
		err2 := errors.New("e2")
		h1 := &fakeHandler{enabled: true, err: err1}
		h2 := &fakeHandler{enabled: true, err: err2}

		mh := newMultiHandler(h1, h2)
		err := mh.Handle(context.Background(), slog.Record{})

		assert.ErrorIs(t, err, err1)
		assert.ErrorIs(t, err, err2)
	})

	t.Run("nil when all succeed", func(t *testing.T) {
		h1 := &fakeHandler{enabled: true}
		h2 := &fakeHandler{enabled: true}

		mh := newMultiHandler(h1, h2)
		assert.NoError(t, mh.Handle(context.Background(), slog.Record{}))
		assert.Equal(t, 1, h1.calls)
		assert.Equal(t, 1, h2.calls)
	})

	t.Run("disabled handler skipped", func(t *testing.T) {
		h1 := &fakeHandler{enabled: false, err: errors.New("should not run")}
		h2 := &fakeHandler{enabled: true}

		mh := newMultiHandler(h1, h2)
		assert.NoError(t, mh.Handle(context.Background(), slog.Record{}))
		assert.Equal(t, 0, h1.calls)
		assert.Equal(t, 1, h2.calls)
	})
}

// +build !integration unit

package opa

import (
	"context"
	"os"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/open-policy-agent/opa/storage/inmem"
	"github.com/stretchr/testify/require"
)

func Test(t *testing.T) {
	var r error
	ctx := context.Background()

	f, err := os.Open("example/store-pretty.json")
	require.NoError(t, err, "read example store JSON file")
	defer f.Close()

	l := log.NewLogfmtLogger(os.Stderr)
	store := inmem.NewFromReader(f)

	t.Run("default policy", func(t *testing.T) {
		s, err := New(ctx, l)
		require.NoError(t, err, "read example store JSON file")
		s.store = store

		r = s.initPartialResult(ctx)
		if r != nil {
			t.Error(r)
		}
	})
}

package jobs

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/friendsofshopware/shopmon/api/internal/shopwareaccount"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type syncResultStub struct {
	err error
}

func (s syncResultStub) Sync(context.Context, []string, string) error {
	return s.err
}

func TestHandleStoreExtensionSyncAcksRateLimit(t *testing.T) {
	err := handleStoreExtensionSync(context.Background(), syncResultStub{
		err: &shopwareaccount.APIError{StatusCode: http.StatusTooManyRequests},
	}, StoreExtensionSync{Names: []string{"FroshTools"}, ShopwareVersion: "6.6.0.0"})
	require.NoError(t, err, "rate-limit abort must be acked")
}

func TestHandleStoreExtensionSyncAcksWrappedRateLimit(t *testing.T) {
	err := handleStoreExtensionSync(context.Background(), syncResultStub{
		err: fmt.Errorf("store rate limited after probing 1/2 version(s): %w", &shopwareaccount.APIError{StatusCode: http.StatusTooManyRequests}),
	}, StoreExtensionSync{})
	require.NoError(t, err, "wrapped rate-limit must be acked")
}

func TestHandleStoreExtensionSyncPropagatesOtherErrors(t *testing.T) {
	want := errors.New("all store probes failed for 2 version(s)")
	err := handleStoreExtensionSync(context.Background(), syncResultStub{err: want}, StoreExtensionSync{})
	assert.Equal(t, want, err)
}

func TestHandleStoreExtensionSyncPropagatesNon429APIError(t *testing.T) {
	want := &shopwareaccount.APIError{StatusCode: http.StatusInternalServerError}
	err := handleStoreExtensionSync(context.Background(), syncResultStub{err: want}, StoreExtensionSync{})
	assert.Equal(t, want, err)
}

package shopware

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/crypto"
	"github.com/friendsofshopware/shopmon/api/internal/environment"
	"github.com/friendsofshopware/shopmon/api/internal/monitoring"
	shopwareclient "github.com/friendsofshopware/shopmon/api/internal/shopware"
)

type Gateway struct {
	appSecret string
}

var _ monitoring.EnvironmentGateway = (*Gateway)(nil)

func NewGateway(appSecret string) *Gateway {
	return &Gateway{appSecret: appSecret}
}

func (g *Gateway) ValidateConnection(ctx context.Context, credentials monitoring.ConnectionCredentials) (monitoring.ShopInfo, error) {
	client := shopwareclient.NewClient(credentials.URL, credentials.ClientID, credentials.ClientSecret, credentials.EnvironmentToken)
	body, err := client.Get(ctx, "/_info/config")
	if err != nil {
		return monitoring.ShopInfo{}, fmt.Errorf("fetch shop config: %w", err)
	}
	var info monitoring.ShopInfo
	if err := json.Unmarshal(body, &info); err != nil {
		return monitoring.ShopInfo{}, fmt.Errorf("decode shop config: %w", err)
	}
	return info, nil
}

func (g *Gateway) EncryptSecret(secret string) (string, error) {
	encrypted, err := crypto.Encrypt(secret, g.appSecret)
	if err != nil {
		return "", fmt.Errorf("encrypt client secret: %w", err)
	}
	return encrypted, nil
}

func (g *Gateway) ClearCache(ctx context.Context, credentials monitoring.StoredCredentials) error {
	client, err := g.client(credentials)
	if err != nil {
		return err
	}
	if _, err := client.Delete(ctx, "/_action/cache", nil); err != nil {
		return fmt.Errorf("%w: clear environment cache: %w", monitoring.ErrRemoteOperation, err)
	}
	return nil
}

func (g *Gateway) RescheduleTask(ctx context.Context, credentials monitoring.StoredCredentials, taskID string, nextExecution time.Time) error {
	client, err := g.client(credentials)
	if err != nil {
		return err
	}
	body := map[string]any{
		"status":            "scheduled",
		"nextExecutionTime": nextExecution.Format("2006-01-02T15:04:05+00:00"),
	}
	if _, err := client.Patch(ctx, "/scheduled-task/"+taskID, body); err != nil {
		return fmt.Errorf("%w: reschedule environment task: %w", monitoring.ErrRemoteOperation, err)
	}
	if _, err := client.Post(ctx, "/_action/scheduled-task/run", nil); err != nil {
		slog.WarnContext(ctx, "failed to trigger scheduled task runner", "error", err)
	}
	return nil
}

func (g *Gateway) client(credentials monitoring.StoredCredentials) (*shopwareclient.Client, error) {
	client, err := environment.NewClientFromEncrypted(
		credentials.URL,
		credentials.ClientID,
		credentials.EncryptedClientSecret,
		credentials.EnvironmentToken,
		g.appSecret,
	)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", monitoring.ErrCredentialDecryption, err)
	}
	return client, nil
}

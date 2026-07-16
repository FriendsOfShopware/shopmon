package shopware

import (
	"context"
	"fmt"
	"net/http"

	"github.com/friendsofshopware/shopmon/api/internal/catalog"
	"github.com/friendsofshopware/shopmon/api/internal/shopwareaccount"
)

type Gateway struct {
	client *shopwareaccount.Client
}

var _ catalog.CompatibilityGateway = (*Gateway)(nil)

func NewGateway(client *shopwareaccount.Client) *Gateway {
	return &Gateway{client: client}
}

func (g *Gateway) CheckCompatibility(ctx context.Context, command catalog.CompatibilityCommand) ([]byte, error) {
	extensions := make([]shopwareaccount.CompatibilityExtension, 0, len(command.Extensions))
	for _, extension := range command.Extensions {
		extensions = append(extensions, shopwareaccount.CompatibilityExtension{
			Name:    extension.Name,
			Version: extension.Version,
		})
	}
	response, err := g.client.CheckExtensionCompatibility(ctx, command.CurrentVersion, command.FutureVersion, extensions)
	if err != nil {
		return nil, fmt.Errorf("call Shopware account API: %w", err)
	}
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("shopware account API returned status %d: %s", response.StatusCode, response.Body)
	}
	return response.Body, nil
}

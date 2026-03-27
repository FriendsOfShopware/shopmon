package handler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/friendsofshopware/shopmon/api/internal/crypto"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/shopware"
)

type shopConnectionInfo struct {
	Version string `json:"version"`
}

func (h *Handler) encryptShopSecret(secret string) (string, error) {
	encryptedSecret, err := crypto.Encrypt(secret, h.cfg.AppSecret)
	if err != nil {
		return "", fmt.Errorf("encrypt client secret: %w", err)
	}

	return encryptedSecret, nil
}

func (h *Handler) validateShopConnection(ctx context.Context, shopURL, clientID, clientSecret, shopToken string) (*shopConnectionInfo, error) {
	client := shopware.NewClient(shopURL, clientID, clientSecret, shopToken)
	body, err := client.Get(ctx, "/_info/config")
	if err != nil {
		return nil, fmt.Errorf("fetch shop config: %w", err)
	}

	var info shopConnectionInfo
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, fmt.Errorf("decode shop config: %w", err)
	}

	return &info, nil
}

func (h *Handler) newShopwareClientFromCredentials(creds *queries.GetShopCredentialsRow) (*shopware.Client, error) {
	decryptedSecret, err := crypto.Decrypt(creds.ClientSecret, h.cfg.AppSecret)
	if err != nil {
		return nil, fmt.Errorf("decrypt client secret: %w", err)
	}

	return shopware.NewClient(creds.Url, creds.ClientID, decryptedSecret, creds.ShopToken), nil
}

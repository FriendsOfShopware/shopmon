// Package environment holds pure helpers shared between the HTTP handlers and
// background jobs that operate on environments. It must not depend on net/http
// or the database query layer: helpers take primitive inputs so both callers
// can pass their own row shapes.
package environment

import (
	"fmt"

	"github.com/friendsofshopware/shopmon/api/internal/crypto"
	"github.com/friendsofshopware/shopmon/api/internal/shopware"
)

// NewClientFromEncrypted decrypts an environment's stored client secret with the
// application secret and constructs an authenticated Shopware client. Both the
// HTTP handlers and the scrape job persist the client secret encrypted, so this
// is the single place that turns stored credentials into a usable client.
func NewClientFromEncrypted(url, clientID, encryptedSecret, environmentToken, appSecret string) (*shopware.Client, error) {
	decryptedSecret, err := crypto.Decrypt(encryptedSecret, appSecret)
	if err != nil {
		return nil, fmt.Errorf("decrypt client secret: %w", err)
	}

	return shopware.NewClient(url, clientID, decryptedSecret, environmentToken), nil
}

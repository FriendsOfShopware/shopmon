package checker

import (
	"fmt"
	"testing"

	"github.com/friendsofshopware/shopmon/api/internal/catalog/securityplugin"
	"github.com/stretchr/testify/assert"
)

// The checker duplicates the catalog's Shopware-line-to-plugin-branch map so it
// stays free of catalog dependencies. This sweep pins the two copies together:
// when a new plugin branch is added to the catalog, this test fails until the
// checker's switch learns it too.
func TestPluginBranchMapMatchesCatalog(t *testing.T) {
	t.Parallel()

	for major := 5; major <= 9; major++ {
		for minor := 0; minor <= 30; minor++ {
			v := fmt.Sprintf("%d.%d.0.0", major, minor)
			assert.Equal(t, securityplugin.PluginBranchForShopware(v), pluginBranchForShopware(v),
				"branch maps disagree for Shopware %s", v)
		}
	}
}

func TestPluginBranchForShopwareToleratesVPrefix(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "4", pluginBranchForShopware("v6.7.8.2"))
	assert.Equal(t, "4", securityplugin.PluginBranchForShopware("v6.7.8.2"))
}

func TestPluginMajor(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "4", pluginMajor("4.0.10"))
	assert.Equal(t, "3", pluginMajor("v3.0.15"))
	assert.Equal(t, "4", pluginMajor("4"))
	assert.Equal(t, "", pluginMajor(""))
}

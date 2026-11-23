package securityplugin

import (
	"context"
	"testing"

	"github.com/friendsofshopware/shopmon/api/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The backport map is keyed per (ghsa, branch), so pruning must be too. A flat
// GHSA keep-set would let an advisory still named by the 4.x branch shield the
// stale 3.x row, leaving 6.6 shops credited with a backport their plugin no
// longer carries — a false "covered" verdict, the dangerous direction.
func TestPersistPrunesPerBranchNotGlobally(t *testing.T) {
	env := testutil.Setup(t)
	ctx := context.Background()

	svc := &Service{pool: env.Pool, queries: env.Queries}

	// Both branches initially backport GHSA-SHARED; only 3 backports GHSA-OLD.
	require.NoError(t, svc.persist(ctx,
		[]Fix{
			{GhsaID: "GHSA-SHARED", Branch: "3", PluginVersion: "3.0.15"},
			{GhsaID: "GHSA-SHARED", Branch: "4", PluginVersion: "4.0.10"},
			{GhsaID: "GHSA-OLD", Branch: "3", PluginVersion: "3.0.12"},
		},
		[]Coverage{{Branch: "3", GhsaCount: 2}, {Branch: "4", GhsaCount: 1}},
		[]string{"3", "4"},
	))

	rows, err := env.Queries.ListSecurityPluginFixes(ctx)
	require.NoError(t, err)
	require.Len(t, rows, 3)

	// Re-derive: branch 3 no longer names either advisory; branch 4 still names
	// GHSA-SHARED. The 3.x rows must go even though GHSA-SHARED survives on 4.
	require.NoError(t, svc.persist(ctx,
		[]Fix{
			{GhsaID: "GHSA-SHARED", Branch: "4", PluginVersion: "4.0.10"},
		},
		[]Coverage{{Branch: "3", GhsaCount: 0}, {Branch: "4", GhsaCount: 1}},
		[]string{"3", "4"},
	))

	rows, err = env.Queries.ListSecurityPluginFixes(ctx)
	require.NoError(t, err)
	require.Len(t, rows, 1, "only the branch-4 row should survive: %+v", rows)
	assert.Equal(t, "GHSA-SHARED", rows[0].GhsaID)
	assert.Equal(t, "4", rows[0].PluginBranch)
}

// A branch absent from the accepted list keeps its map untouched, so a parse
// this sync rejected never clears fixes that are still valid.
func TestPersistLeavesUnacceptedBranchesAlone(t *testing.T) {
	env := testutil.Setup(t)
	ctx := context.Background()

	svc := &Service{pool: env.Pool, queries: env.Queries}

	require.NoError(t, svc.persist(ctx,
		[]Fix{
			{GhsaID: "GHSA-KEEP", Branch: "2", PluginVersion: "2.0.9"},
			{GhsaID: "GHSA-OTHER", Branch: "4", PluginVersion: "4.0.10"},
		},
		[]Coverage{{Branch: "2", GhsaCount: 1}, {Branch: "4", GhsaCount: 1}},
		[]string{"2", "4"},
	))

	// Branch 2's parse was rejected this round, so it is not in the accepted
	// list and must be left exactly as it was.
	require.NoError(t, svc.persist(ctx,
		[]Fix{{GhsaID: "GHSA-OTHER", Branch: "4", PluginVersion: "4.0.11"}},
		[]Coverage{{Branch: "4", GhsaCount: 1}},
		[]string{"4"},
	))

	rows, err := env.Queries.ListSecurityPluginFixes(ctx)
	require.NoError(t, err)
	require.Len(t, rows, 2)

	byBranch := map[string]string{}
	for _, row := range rows {
		byBranch[row.PluginBranch] = row.PluginVersion
	}
	assert.Equal(t, "2.0.9", byBranch["2"], "an unaccepted branch must not be pruned")
	assert.Equal(t, "4.0.11", byBranch["4"])
}

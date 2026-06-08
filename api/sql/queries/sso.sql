-- name: ListSSOProviders :many
SELECT id, issuer, oidc_config, saml_config, user_id, provider_id, organization_id, domain
FROM sso_provider WHERE organization_id = $1;

-- name: UpdateSSOProvider :exec
UPDATE sso_provider SET issuer = $1, oidc_config = $2, domain = $3 WHERE provider_id = $4 AND organization_id = $5;

-- name: DeleteSSOProvider :exec
DELETE FROM sso_provider WHERE provider_id = $1 AND organization_id = $2;

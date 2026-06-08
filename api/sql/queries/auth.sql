-- name: CreateUser :one
INSERT INTO "user" (id, name, email, email_verified, created_at, updated_at, role, notifications)
VALUES ($1, $2, $3, false, NOW(), NOW(), 'user', '[]'::jsonb)
RETURNING id;

-- name: GetUserByEmail :one
SELECT id, name, email, email_verified, image, created_at, updated_at, role, banned, ban_reason, ban_expires, notifications
FROM "user" WHERE email = $1;

-- name: UpdateUserEmailVerified :exec
UPDATE "user" SET email_verified = true, updated_at = NOW() WHERE id = $1;

-- name: UpdateUserPassword :exec
UPDATE account SET password = $1, updated_at = NOW() WHERE user_id = $2 AND provider_id = 'credential';

-- name: UpdateUserProfile :exec
UPDATE "user" SET name = $1, image = $2, updated_at = NOW() WHERE id = $3;

-- name: CreateAccount :exec
INSERT INTO account (id, account_id, provider_id, user_id, password, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, NOW(), NOW());

-- name: GetAccountByProviderAndUser :one
SELECT id, account_id, provider_id, user_id, password, access_token, refresh_token
FROM account WHERE provider_id = $1 AND user_id = $2;

-- name: GetAccountByProviderAndAccountID :one
SELECT a.id, a.account_id, a.provider_id, a.user_id, a.password,
       u.id AS "user_id_2", u.name AS user_name, u.email AS user_email, u.role AS user_role,
       u.notifications AS user_notifications, u.image AS user_image, u.banned AS user_banned,
       u.email_verified AS user_email_verified
FROM account a
JOIN "user" u ON u.id = a.user_id
WHERE a.provider_id = $1 AND a.account_id = $2;

-- name: CreateSession :one
INSERT INTO session (id, expires_at, token, created_at, updated_at, ip_address, user_agent, user_id)
VALUES ($1, $2, $3, NOW(), NOW(), $4, $5, $6)
RETURNING id, token;

-- name: DeleteSession :exec
DELETE FROM session WHERE token = $1;

-- name: DeleteUserSessions :exec
DELETE FROM session WHERE user_id = $1;

-- name: CreateVerification :exec
INSERT INTO verification (id, identifier, value, expires_at, created_at, updated_at)
VALUES ($1, $2, $3, $4, NOW(), NOW());

-- name: GetVerificationByValue :one
SELECT id, identifier, value, expires_at FROM verification WHERE value = $1 AND expires_at > NOW();

-- name: DeleteVerification :exec
DELETE FROM verification WHERE id = $1;

-- name: SetActiveOrganization :exec
UPDATE session SET active_organization_id = $1, updated_at = NOW() WHERE token = $2;

-- name: CreatePasskey :exec
INSERT INTO passkey (id, name, public_key, user_id, credential_id, counter, device_type, backed_up, transports, created_at, aaguid)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), $10);

-- name: GetPasskeysByUserID :many
SELECT id, name, public_key, user_id, credential_id, counter, device_type, backed_up, transports, created_at, aaguid
FROM passkey WHERE user_id = $1;

-- name: GetPasskeyByCredentialID :one
SELECT p.id, p.name, p.public_key, p.user_id, p.credential_id, p.counter, p.device_type, p.backed_up, p.transports, p.created_at, p.aaguid,
       u.id AS uid, u.name AS user_name, u.email AS user_email, u.role AS user_role, u.notifications AS user_notifications, u.banned AS user_banned
FROM passkey p
JOIN "user" u ON u.id = p.user_id
WHERE p.credential_id = $1;

-- name: UpdatePasskeyCounter :exec
UPDATE passkey SET counter = $1 WHERE id = $2;

-- name: DeletePasskey :exec
DELETE FROM passkey WHERE id = $1 AND user_id = $2;

-- name: GetSSOProviderByDomain :one
SELECT id, issuer, oidc_config, provider_id, organization_id, domain
FROM sso_provider WHERE domain = $1;

-- name: GetSSOProviderByProviderID :one
SELECT id, issuer, oidc_config, provider_id, organization_id, domain
FROM sso_provider WHERE provider_id = $1;

-- name: IsOrgMember :one
SELECT COUNT(*) > 0 AS is_member FROM member WHERE organization_id = $1 AND user_id = $2;

-- name: AdminListUsers :many
SELECT id, name, email, email_verified, image, role, banned, ban_reason, created_at
FROM "user" ORDER BY created_at DESC LIMIT $1 OFFSET $2;

-- name: AdminSetUserRole :exec
UPDATE "user" SET role = $1, updated_at = NOW() WHERE id = $2;

-- name: AdminBanUser :exec
UPDATE "user" SET banned = true, ban_reason = $1, updated_at = NOW() WHERE id = $2;

-- name: AdminUnbanUser :exec
UPDATE "user" SET banned = false, ban_reason = NULL, ban_expires = NULL, updated_at = NOW() WHERE id = $1;

-- name: AdminCreateImpersonationSession :one
INSERT INTO session (id, expires_at, token, created_at, updated_at, ip_address, user_agent, user_id, impersonated_by)
VALUES ($1, $2, $3, NOW(), NOW(), $4, $5, $6, $7)
RETURNING id, token;

-- name: ListUserSessions :many
SELECT id, expires_at, created_at, ip_address, user_agent, impersonated_by
FROM session WHERE user_id = $1 AND expires_at > NOW() ORDER BY created_at DESC;

-- name: DeleteSessionByID :exec
DELETE FROM session WHERE id = $1 AND user_id = $2;

-- name: ListUserAccounts :many
SELECT id, provider_id, account_id, created_at FROM account WHERE user_id = $1;

-- name: DeleteAccountByProviderAndUser :exec
DELETE FROM account WHERE provider_id = $1 AND user_id = $2;

-- name: UpdateUserEmail :exec
UPDATE "user" SET email = $1, updated_at = NOW() WHERE id = $2;

-- name: UpdateUserName :exec
UPDATE "user" SET name = $1, updated_at = NOW() WHERE id = $2;

-- name: DeleteUser :exec
DELETE FROM "user" WHERE id = $1;

-- name: ListUserPasskeys :many
SELECT id, name, credential_id, device_type, backed_up, created_at FROM passkey WHERE user_id = $1;

-- name: ListUserOrganizations :many
SELECT o.id, o.name, o.slug, o.logo, o.created_at, m.role
FROM organization o
JOIN member m ON m.organization_id = o.id
WHERE m.user_id = $1
ORDER BY o.name;

-- name: DeleteInvitationByID :exec
DELETE FROM invitation WHERE id = $1;

-- name: CreateSSOProvider :exec
INSERT INTO sso_provider (id, issuer, oidc_config, provider_id, organization_id, domain)
VALUES ($1, $2, $3, $4, $5, $6);

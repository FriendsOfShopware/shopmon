-- name: CreateOrganization :one
INSERT INTO organization (id, name, slug, created_at)
VALUES ($1, $2, $3, NOW())
RETURNING id;

-- name: GetOrganizationByID :one
SELECT id, name, slug, logo, created_at FROM organization WHERE id = $1;

-- name: UpdateOrganization :exec
UPDATE organization SET name = $1, slug = $2, logo = $3 WHERE id = $4;

-- name: DeleteOrganization :exec
DELETE FROM organization WHERE id = $1;

-- name: CreateMember :exec
INSERT INTO member (id, organization_id, user_id, role, created_at)
VALUES ($1, $2, $3, $4, NOW());

-- name: GetMemberRole :one
SELECT role FROM member WHERE organization_id = $1 AND user_id = $2;

-- name: ListMembers :many
SELECT m.id, m.organization_id, m.user_id, m.role, m.created_at,
       u.name AS user_name, u.email AS user_email, u.image AS user_image
FROM member m
JOIN "user" u ON u.id = m.user_id
WHERE m.organization_id = $1
ORDER BY m.created_at;

-- name: UpdateMemberRole :exec
UPDATE member SET role = $1 WHERE organization_id = $2 AND user_id = $3;

-- name: DeleteMember :exec
DELETE FROM member WHERE organization_id = $1 AND user_id = $2;

-- name: CountOrgMembers :one
SELECT COUNT(*)::int FROM member WHERE organization_id = $1;

-- name: CreateInvitation :one
INSERT INTO invitation (id, organization_id, email, role, status, expires_at, created_at, inviter_id)
VALUES ($1, $2, $3, $4, 'pending', $5, NOW(), $6)
RETURNING id;

-- name: GetInvitationByID :one
SELECT i.id, i.organization_id, i.email, i.role, i.status, i.expires_at, i.inviter_id,
       o.name AS organization_name,
       u.name AS inviter_name
FROM invitation i
JOIN organization o ON o.id = i.organization_id
JOIN "user" u ON u.id = i.inviter_id
WHERE i.id = $1 AND i.status = 'pending' AND i.expires_at > NOW();

-- name: ListInvitations :many
SELECT i.id, i.organization_id, i.email, i.role, i.status, i.expires_at, i.created_at, i.inviter_id,
       u.name AS inviter_name
FROM invitation i
JOIN "user" u ON u.id = i.inviter_id
WHERE i.organization_id = $1 AND i.status = 'pending'
ORDER BY i.created_at DESC;

-- name: UpdateInvitationStatus :exec
UPDATE invitation SET status = $1 WHERE id = $2;

-- name: CountOrgOwners :one
SELECT COUNT(*)::int FROM member WHERE organization_id = $1 AND role = 'owner';

-- name: DeleteInvitation :exec
DELETE FROM invitation WHERE id = $1;

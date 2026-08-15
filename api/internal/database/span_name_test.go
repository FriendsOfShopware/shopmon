package database

import "testing"

func TestSQLSpanName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		stmt string
		want string
	}{
		{
			name: "sqlc named insert",
			stmt: `-- name: UpsertEnvironmentExtension :exec
INSERT INTO environment_extension (id) VALUES ($1)
`,
			want: "UpsertEnvironmentExtension",
		},
		{
			name: "sqlc named select",
			stmt: `-- name: GetEnvironmentByID :one
SELECT * FROM environment WHERE id = $1
`,
			want: "GetEnvironmentByID",
		},
		{
			name: "sqlc named update",
			stmt: `-- name: UpdateEnvironmentName :exec
UPDATE environment SET name = $1 WHERE id = $2
`,
			want: "UpdateEnvironmentName",
		},
		{
			name: "sqlc named delete",
			stmt: `-- name: DeleteEnvironment :exec
DELETE FROM environment WHERE id = $1
`,
			want: "DeleteEnvironment",
		},
		{
			name: "plain select without sqlc comment",
			stmt: `SELECT id FROM environment WHERE id = $1`,
			want: "SELECT",
		},
		{
			name: "plain insert without sqlc comment",
			stmt: `INSERT INTO environment (name) VALUES ($1)`,
			want: "INSERT",
		},
		{
			name: "plain with query",
			stmt: `WITH recent AS (SELECT 1) SELECT * FROM recent`,
			want: "WITH",
		},
		{
			name: "leading whitespace before sqlc name",
			stmt: `
  -- name: ListEnvironmentsByOrganization :many
SELECT id FROM environment
`,
			want: "ListEnvironmentsByOrganization",
		},
		{
			name: "leading whitespace before plain sql",
			stmt: `

		UPDATE environment SET name = $1 WHERE id = $2
`,
			want: "UPDATE",
		},
		{
			name: "multiple non-sqlc comments before statement",
			stmt: `-- skip vacuum during maintenance
-- caller holds advisory lock
DELETE FROM environment_check WHERE environment_id = $1
`,
			want: "DELETE",
		},
		{
			name: "non-sqlc comments then sqlc name",
			stmt: `-- prepared for scrape
-- name: CreateEnvironment :one
INSERT INTO environment (name) VALUES ($1) RETURNING id
`,
			want: "CreateEnvironment",
		},
		{
			name: "empty statement",
			stmt: "   \n\t  ",
			want: "UNKNOWN",
		},
		{
			name: "only comments",
			stmt: "-- just a comment\n-- another",
			want: "UNKNOWN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := SQLSpanName(tt.stmt); got != tt.want {
				t.Fatalf("SQLSpanName() = %q, want %q", got, tt.want)
			}
		})
	}
}

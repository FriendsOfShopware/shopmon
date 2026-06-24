// Seeded fixture data from `api/fixtures.go` (run `cd api && go run . fixtures`).
// All seeded users share the password below.
export const PASSWORD = "password";

export const USERS = {
  owner: { email: "owner@fos.gg", name: "Owner", role: "admin" },
  admin: { email: "admin@fos.gg", name: "Admin" },
  member: { email: "member@fos.gg", name: "Member" },
  // `regular` has no organization membership — useful for empty-state coverage.
  regular: { email: "regular@fos.gg", name: "Regular" },
} as const;

export const ORG = { name: "Acme Corp" } as const;
export const SHOP = { name: "Acme Shop", id: 1 } as const;
// Seeded environments belonging to the shop above.
export const ENVIRONMENTS = {
  production: { id: 1, name: "Production" },
  staging: { id: 2, name: "Staging" },
} as const;

export const STORAGE_STATE = "e2e/.auth/user.json";

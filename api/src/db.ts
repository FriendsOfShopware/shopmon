import { relations } from "drizzle-orm";
import {
  boolean,
  index,
  integer,
  jsonb,
  pgTable,
  real,
  serial,
  text,
  timestamp,
  unique,
} from "drizzle-orm/pg-core";
import { drizzle, type PostgresJsDatabase } from "drizzle-orm/postgres-js";
import postgres from "postgres";
import type { ExtensionDiff, NotificationLink } from "./types/index.ts";

type LastChangelog = {
  date: Date;
  from: string;
  to: string;
};

export const organization = pgTable("organization", {
  id: text("id").primaryKey(),
  name: text("name").notNull(),
  slug: text("slug").notNull().unique(),
  logo: text("logo"),
  createdAt: timestamp("created_at").notNull(),
  metadata: text("metadata"),
});

export const project = pgTable("project", {
  id: serial("id").primaryKey(),
  organizationId: text("organization_id")
    .notNull()
    .references(() => organization.id, { onDelete: "cascade" }),
  name: text("name").notNull(),
  description: text("description"),
  createdAt: timestamp("created_at").notNull(),
  updatedAt: timestamp("updated_at").notNull(),
});

export const shop = pgTable("shop", {
  id: serial("id").primaryKey(),
  organizationId: text("organization_id")
    .notNull()
    .references(() => organization.id),
  projectId: integer("project_id")
    .notNull()
    .references(() => project.id),
  name: text("name").notNull(),
  status: text("status").notNull().default("green"),
  url: text("url").notNull(),
  favicon: text("favicon"),
  clientId: text("client_id").notNull(),
  clientSecret: text("client_secret").notNull(),
  shopwareVersion: text("shopware_version").notNull(),
  lastScrapedAt: timestamp("last_scraped_at"),
  lastScrapedError: text("last_scraped_error"),
  ignores: jsonb("ignores").default([]).$type<string[]>().notNull(),
  shopImage: text("shop_image"),
  lastChangelog: jsonb("last_changelog").default({}).$type<LastChangelog>(),
  connectionIssueCount: integer("connection_issue_count").default(0).notNull(),
  sitespeedEnabled: boolean("sitespeed_enabled").default(false).notNull(),
  sitespeedUrls: jsonb("sitespeed_urls").default([]).$type<string[]>().notNull(),
  createdAt: timestamp("created_at").notNull(),
});

export const shopSitespeed = pgTable("shop_sitespeed", {
  id: serial("id").primaryKey(),
  shopId: integer("shop_id").references(() => shop.id),
  createdAt: timestamp("created_at").notNull(),
  ttfb: integer("ttfb"),
  fullyLoaded: integer("fully_loaded"),
  largestContentfulPaint: integer("largest_contentful_paint"),
  firstContentfulPaint: integer("first_contentful_paint"),
  cumulativeLayoutShift: real("cumulative_layout_shift"),
  transferSize: integer("transfer_size"),
});

export const shopChangelog = pgTable("shop_changelog", {
  id: serial("id").primaryKey(),
  shopId: integer("shop_id").references(() => shop.id),
  extensions: jsonb("extensions").notNull().$type<ExtensionDiff[]>(),
  oldShopwareVersion: text("old_shopware_version"),
  newShopwareVersion: text("new_shopware_version"),
  date: timestamp("date").notNull(),
});

export const userNotification = pgTable(
  "user_notification",
  {
    id: serial("id").primaryKey(),
    userId: text("user_id")
      .notNull()
      .references(() => user.id),
    key: text("key").notNull(),
    level: text("level").notNull(),
    title: text("title").notNull(),
    message: text("message").notNull(),
    link: jsonb("link").notNull().$type<NotificationLink>(),
    read: boolean("read").notNull().default(false),
    createdAt: timestamp("created_at").notNull(),
  },
  (table) => ({
    keyUnique: unique().on(table.userId, table.key),
  }),
);

export const user = pgTable("user", {
  id: text("id").primaryKey(),
  name: text("name").notNull(),
  email: text("email").notNull().unique(),
  emailVerified: boolean("email_verified").default(false).notNull(),
  image: text("image"),
  createdAt: timestamp("created_at").notNull(),
  updatedAt: timestamp("updated_at")
    .$onUpdate(() => new Date())
    .notNull(),
  role: text("role").default("user").notNull(),
  banned: boolean("banned").default(false),
  banReason: text("ban_reason"),
  banExpires: timestamp("ban_expires"),
  notifications: jsonb("notifications").default([]).$type<string[]>(),
});

export const session = pgTable(
  "session",
  {
    id: text("id").primaryKey(),
    expiresAt: timestamp("expires_at").notNull(),
    token: text("token").notNull().unique(),
    createdAt: timestamp("created_at").notNull(),
    updatedAt: timestamp("updated_at")
      .$onUpdate(() => new Date())
      .notNull(),
    ipAddress: text("ip_address"),
    userAgent: text("user_agent"),
    userId: text("user_id")
      .notNull()
      .references(() => user.id, { onDelete: "cascade" }),
    impersonatedBy: text("impersonated_by"),
    activeOrganizationId: text("active_organization_id"),
  },
  (table) => [index("session_userId_idx").on(table.userId)],
);

export const account = pgTable(
  "account",
  {
    id: text("id").primaryKey(),
    accountId: text("account_id").notNull(),
    providerId: text("provider_id").notNull(),
    userId: text("user_id")
      .notNull()
      .references(() => user.id, { onDelete: "cascade" }),
    accessToken: text("access_token"),
    refreshToken: text("refresh_token"),
    idToken: text("id_token"),
    accessTokenExpiresAt: timestamp("access_token_expires_at"),
    refreshTokenExpiresAt: timestamp("refresh_token_expires_at"),
    scope: text("scope"),
    password: text("password"),
    createdAt: timestamp("created_at").notNull(),
    updatedAt: timestamp("updated_at")
      .$onUpdate(() => new Date())
      .notNull(),
  },
  (table) => [index("account_userId_idx").on(table.userId)],
);

export const verification = pgTable(
  "verification",
  {
    id: text("id").primaryKey(),
    identifier: text("identifier").notNull(),
    value: text("value").notNull(),
    expiresAt: timestamp("expires_at").notNull(),
    createdAt: timestamp("created_at").notNull(),
    updatedAt: timestamp("updated_at")
      .$onUpdate(() => new Date())
      .notNull(),
  },
  (table) => [index("verification_identifier_idx").on(table.identifier)],
);

export const passkey = pgTable(
  "passkey",
  {
    id: text("id").primaryKey(),
    name: text("name"),
    publicKey: text("public_key").notNull(),
    userId: text("user_id")
      .notNull()
      .references(() => user.id, { onDelete: "cascade" }),
    credentialID: text("credential_id").notNull(),
    counter: integer("counter").notNull(),
    deviceType: text("device_type").notNull(),
    backedUp: boolean("backed_up").notNull(),
    transports: text("transports"),
    createdAt: timestamp("created_at"),
    aaguid: text("aaguid"),
  },
  (table) => [
    index("passkey_userId_idx").on(table.userId),
    index("passkey_credentialID_idx").on(table.credentialID),
  ],
);

export const lock = pgTable("lock", {
  key: text("key").primaryKey(),
  expires: timestamp("expires").notNull(),
  createdAt: timestamp("created_at").notNull(),
});

// Shop scrape info tables (flat structure)
export const shopExtension = pgTable(
  "shop_extension",
  {
    id: serial("id").primaryKey(),
    shopId: integer("shop_id")
      .notNull()
      .references(() => shop.id, { onDelete: "cascade" }),
    name: text("name").notNull(),
    label: text("label").notNull(),
    active: boolean("active").notNull(),
    version: text("version").notNull(),
    latestVersion: text("latest_version"),
    installed: boolean("installed").notNull(),
    ratingAverage: integer("rating_average"),
    storeLink: text("store_link"),
    changelog: jsonb("changelog").$type<
      | {
          version: string;
          text: string;
          creationDate: string;
          isCompatible: boolean;
        }[]
      | null
    >(),
    installedAt: text("installed_at"),
  },
  (table) => ({
    shopExtUnique: unique().on(table.shopId, table.name),
  }),
);

export const shopScheduledTask = pgTable(
  "shop_scheduled_task",
  {
    id: serial("id").primaryKey(),
    shopId: integer("shop_id")
      .notNull()
      .references(() => shop.id, { onDelete: "cascade" }),
    taskId: text("task_id").notNull(),
    name: text("name").notNull(),
    status: text("status").notNull(),
    interval: integer("interval").notNull(),
    overdue: boolean("overdue").notNull(),
    lastExecutionTime: text("last_execution_time"),
    nextExecutionTime: text("next_execution_time"),
  },
  (table) => ({
    shopTaskUnique: unique().on(table.shopId, table.taskId),
  }),
);

export const shopQueue = pgTable(
  "shop_queue",
  {
    id: serial("id").primaryKey(),
    shopId: integer("shop_id")
      .notNull()
      .references(() => shop.id, { onDelete: "cascade" }),
    name: text("name").notNull(),
    size: integer("size").notNull(),
  },
  (table) => ({
    shopQueueUnique: unique().on(table.shopId, table.name),
  }),
);

export const shopCache = pgTable("shop_cache", {
  id: serial("id").primaryKey(),
  shopId: integer("shop_id")
    .notNull()
    .references(() => shop.id, { onDelete: "cascade" })
    .unique(),
  environment: text("environment").notNull(),
  httpCache: boolean("http_cache").notNull(),
  cacheAdapter: text("cache_adapter").notNull(),
});

export const shopCheck = pgTable(
  "shop_check",
  {
    id: serial("id").primaryKey(),
    shopId: integer("shop_id")
      .notNull()
      .references(() => shop.id, { onDelete: "cascade" }),
    checkId: text("check_id").notNull(),
    level: text("level").notNull(),
    message: text("message").notNull(),
    source: text("source").notNull(),
    link: text("link"),
  },
  (table) => ({
    shopCheckUnique: unique().on(table.shopId, table.checkId),
  }),
);

export const member = pgTable(
  "member",
  {
    id: text("id").primaryKey(),
    organizationId: text("organization_id")
      .notNull()
      .references(() => organization.id, { onDelete: "cascade" }),
    userId: text("user_id")
      .notNull()
      .references(() => user.id, { onDelete: "cascade" }),
    role: text("role").default("member").notNull(),
    createdAt: timestamp("created_at").notNull(),
  },
  (table) => [
    index("member_organizationId_idx").on(table.organizationId),
    index("member_userId_idx").on(table.userId),
  ],
);

export const invitation = pgTable(
  "invitation",
  {
    id: text("id").primaryKey(),
    organizationId: text("organization_id")
      .notNull()
      .references(() => organization.id, { onDelete: "cascade" }),
    email: text("email").notNull(),
    role: text("role"),
    status: text("status").default("pending").notNull(),
    expiresAt: timestamp("expires_at").notNull(),
    createdAt: timestamp("created_at").notNull(),
    inviterId: text("inviter_id")
      .notNull()
      .references(() => user.id, { onDelete: "cascade" }),
  },
  (table) => [
    index("invitation_organizationId_idx").on(table.organizationId),
    index("invitation_email_idx").on(table.email),
  ],
);

export const ssoProvider = pgTable("sso_provider", {
  id: text("id").primaryKey(),
  issuer: text("issuer").notNull(),
  oidcConfig: text("oidc_config"),
  samlConfig: text("saml_config"),
  userId: text("user_id").references(() => user.id, { onDelete: "cascade" }),
  providerId: text("provider_id").notNull().unique(),
  organizationId: text("organization_id"),
  domain: text("domain").notNull(),
});

export type ApiKeyScope = "deployments";

export const projectApiKey = pgTable("project_api_key", {
  id: text("id").primaryKey(),
  projectId: integer("project_id")
    .notNull()
    .references(() => project.id, { onDelete: "cascade" }),
  name: text("name").notNull(),
  token: text("token").notNull().unique(),
  scopes: jsonb("scopes").notNull().$type<ApiKeyScope[]>(),
  createdAt: timestamp("created_at").notNull(),
  lastUsedAt: timestamp("last_used_at"),
});

// Auth Relations
export const userRelations = relations(user, ({ many }) => ({
  sessions: many(session),
  accounts: many(account),
  passkeys: many(passkey),
  members: many(member),
  invitations: many(invitation),
  ssoProviders: many(ssoProvider),
}));

export const sessionRelations = relations(session, ({ one }) => ({
  user: one(user, {
    fields: [session.userId],
    references: [user.id],
  }),
}));

export const accountRelations = relations(account, ({ one }) => ({
  user: one(user, {
    fields: [account.userId],
    references: [user.id],
  }),
}));

export const passkeyRelations = relations(passkey, ({ one }) => ({
  user: one(user, {
    fields: [passkey.userId],
    references: [user.id],
  }),
}));

export const deploymentToken = pgTable("deployment_token", {
  id: text("id").primaryKey(),
  shopId: integer("shop_id")
    .notNull()
    .references(() => shop.id, { onDelete: "cascade" }),
  token: text("token").notNull().unique(),
  name: text("name").notNull(),
  createdAt: timestamp("created_at").notNull(),
  lastUsedAt: timestamp("last_used_at"),
});

export const deployment = pgTable("deployment", {
  id: serial("id").primaryKey(),
  shopId: integer("shop_id")
    .notNull()
    .references(() => shop.id, { onDelete: "cascade" }),
  name: text("name").notNull(),
  command: text("command").notNull(),
  returnCode: integer("return_code").notNull(),
  startDate: timestamp("start_date").notNull(),
  endDate: timestamp("end_date").notNull(),
  executionTime: text("execution_time").notNull(),
  composer: jsonb("composer").default({}).$type<Record<string, string>>(),
  reference: text("reference"),
  createdAt: timestamp("created_at").notNull(),
});

export const memberRelations = relations(member, ({ one }) => ({
  organization: one(organization, {
    fields: [member.organizationId],
    references: [organization.id],
  }),
  user: one(user, {
    fields: [member.userId],
    references: [user.id],
  }),
}));

export const invitationRelations = relations(invitation, ({ one }) => ({
  organization: one(organization, {
    fields: [invitation.organizationId],
    references: [organization.id],
  }),
  user: one(user, {
    fields: [invitation.inviterId],
    references: [user.id],
  }),
}));

export const ssoProviderRelations = relations(ssoProvider, ({ one }) => ({
  user: one(user, {
    fields: [ssoProvider.userId],
    references: [user.id],
  }),
}));

// App Relations
export const projectRelations = relations(project, ({ one, many }) => ({
  organization: one(organization, {
    fields: [project.organizationId],
    references: [organization.id],
  }),
  shops: many(shop),
  apiKeys: many(projectApiKey),
}));

export const projectApiKeyRelations = relations(projectApiKey, ({ one }) => ({
  project: one(project, {
    fields: [projectApiKey.projectId],
    references: [project.id],
  }),
}));

export const shopRelations = relations(shop, ({ one }) => ({
  organization: one(organization, {
    fields: [shop.organizationId],
    references: [organization.id],
  }),
  project: one(project, {
    fields: [shop.projectId],
    references: [project.id],
  }),
}));

export const organizationRelations = relations(organization, ({ many }) => ({
  projects: many(project),
  shops: many(shop),
}));

export const deploymentRelations = relations(deployment, ({ one }) => ({
  shop: one(shop, {
    fields: [deployment.shopId],
    references: [shop.id],
  }),
}));

export const deploymentTokenRelations = relations(deploymentToken, ({ one }) => ({
  shop: one(shop, {
    fields: [deploymentToken.shopId],
    references: [shop.id],
  }),
}));

export const schema = {
  shop,
  shopSitespeed,
  shopChangelog,
  userNotification,

  // Shop scrape info tables
  shopExtension,
  shopScheduledTask,
  shopQueue,
  shopCache,
  shopCheck,

  // Better Auth
  user,
  session,
  account,
  verification,
  passkey,
  organization,
  project,
  member,
  invitation,
  ssoProvider,

  // Project API Keys
  projectApiKey,

  // Lock
  lock,

  // Deployments
  deployment,
  deploymentToken,

  // Auth Relations
  userRelations,
  sessionRelations,
  accountRelations,
  passkeyRelations,
  memberRelations,
  invitationRelations,
  ssoProviderRelations,

  // App Relations
  projectRelations,
  projectApiKeyRelations,
  shopRelations,
  organizationRelations,
  deploymentRelations,
  deploymentTokenRelations,
};

export type Drizzle = PostgresJsDatabase<typeof schema>;
let db: Drizzle | undefined;
let sqlClient: ReturnType<typeof postgres> | undefined;

export function getConnection() {
  if (db !== undefined) {
    return db;
  }

  const connectionString =
    process.env.DATABASE_URL || "postgres://shopmon:shopmon@localhost:5432/shopmon";

  sqlClient = postgres(connectionString);
  db = drizzle(sqlClient, { schema });

  return db;
}

export async function closeConnection() {
  if (sqlClient) {
    await sqlClient.end();
    sqlClient = undefined;
    db = undefined;
  }
}

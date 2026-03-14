import { HttpClient, type HttpClientResponse } from "#src/modules/shop/http-client.ts";
import { auth } from "#src/auth.ts";
import { scrapeSingleShop } from "#src/modules/shop/jobs/shop-scrape.job.ts";
import { encrypt } from "#src/modules/shop/crypto.ts";
import { closeConnection, getConnection, schema } from "#src/db.ts";
import shops, { generateShopToken } from "#src/modules/shop/shop.repository.ts";
import { eq } from "drizzle-orm";

const user1 = await auth.api.signUpEmail({
  body: {
    email: "owner@fos.gg",
    password: "password",
    name: "Owner",
  },
});

const org = await auth.api.createOrganization({
  body: {
    name: "Acme Corp",
    slug: "acme-corp",
    userId: user1.user.id,
  },
});

const user2 = await auth.api.signUpEmail({
  body: {
    email: "admin@fos.gg",
    password: "password",
    name: "Admin",
  },
});

await auth.api.addMember({
  body: {
    role: "admin",
    userId: user2.user.id,
    organizationId: org.id,
  },
});

const user3 = await auth.api.signUpEmail({
  body: {
    email: "member@fos.gg",
    password: "password",
    name: "Member",
  },
});

await auth.api.addMember({
  body: {
    role: "member",
    userId: user3.user.id,
    organizationId: org.id,
  },
});

const user4 = await auth.api.signUpEmail({
  body: {
    email: "regular@fos.gg",
    password: "password",
    name: "Regular",
  },
});

await getConnection().update(schema.user).set({ emailVerified: true }).execute();

await getConnection()
  .update(schema.user)
  .set({ role: "admin" })
  .where(eq(schema.user.id, user1.user.id))
  .execute();

await getConnection().insert(schema.project).values({
  id: 1,
  name: "Acme Shop",
  organizationId: org.id,
  createdAt: new Date(),
  updatedAt: new Date(),
});

const shopUrl = "http://localhost:3889";
const shopClientId = "SWIAUZL4OXRKEG1RR3PMCEVNMG";
const shopClientSecret = "aXhNQ3NoRHZONmxPYktHT0c2c09rNkR0UHI0elZHOFIycjBzWks";
const shopToken = generateShopToken();

const client = new HttpClient({
  url: shopUrl,
  clientId: shopClientId,
  clientSecret: shopClientSecret,
  shopToken,
});

const resp: HttpClientResponse<{ version: string }> = await client.get("/_info/config");

const shopId = await shops.createShop(getConnection(), {
  name: "Local",
  organizationId: org.id,
  projectId: 1,
  shopUrl,
  clientId: shopClientId,
  clientSecret: await encrypt(process.env.APP_SECRET, shopClientSecret),
  version: resp.body.version,
  shopToken,
});

await scrapeSingleShop(shopId);

console.log("Fixtures applied successfully");
console.log("User 1 (org owner):", user1.user.email);
console.log("User 2 (org admin):", user2.user.email);
console.log("User 3 (org member):", user3.user.email);
console.log("User 4 (regular user without organization):", user4.user.email);
console.log("Organization:", org.name);
console.log("Shop:", shopUrl);

console.log('All users have the password "password".');

await closeConnection();

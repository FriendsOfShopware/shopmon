import {
    HttpClient,
    type HttpClientResponse,
    SimpleShop,
} from '@shopware-ag/app-server-sdk';
import { auth } from './src/auth.ts';
import { scrapeSingleShop } from './src/cron/jobs/shopScrape.ts';
import { encrypt } from './src/crypto/index.ts';
import { getConnection, schema } from './src/db.ts';
import shops from './src/repository/shops.ts';

const user1 = await auth.api.signUpEmail({
    body: {
        email: 'owner@fos.gg',
        password: 'password',
        name: 'Owner',
    },
});

const org = await auth.api.createOrganization({
    body: {
        name: 'Acme Corp',
        slug: 'acme-corp',
        userId: user1.user.id,
    },
});

const user2 = await auth.api.signUpEmail({
    body: {
        email: 'admin@fos.gg',
        password: 'password',
        name: 'Admin',
    },
});

await auth.api.addMember({
    body: {
        role: 'admin',
        userId: user2.user.id,
        organizationId: org.id,
    },
});

const user3 = await auth.api.signUpEmail({
    body: {
        email: 'member@fos.gg',
        password: 'password',
        name: 'Member',
    },
});

await auth.api.addMember({
    body: {
        role: 'member',
        userId: user3.user.id,
        organizationId: org.id,
    },
});

const user4 = await auth.api.signUpEmail({
    body: {
        email: 'regular@fos.gg',
        password: 'password',
        name: 'Regular',
    },
});

await getConnection()
    .update(schema.user)
    .set({ emailVerified: true })
    .execute();

await getConnection().insert(schema.project).values({
    id: 1,
    name: 'Acme Shop',
    organizationId: org.id,
    createdAt: new Date(),
    updatedAt: new Date(),
});

const shop = new SimpleShop('', 'http://localhost:3889', '');
shop.setShopCredentials(
    'SWIAUZL4OXRKEG1RR3PMCEVNMG',
    'aXhNQ3NoRHZONmxPYktHT0c2c09rNkR0UHI0elZHOFIycjBzWks',
);

const client = new HttpClient(shop);

const resp: HttpClientResponse<{ version: string }> =
    await client.get('/_info/config');

const shopId = await shops.createShop(getConnection(), {
    name: 'Local',
    organizationId: org.id,
    projectId: 1,
    shopUrl: shop.getShopUrl(),
    clientId: shop.getShopClientId(),
    clientSecret: await encrypt(
        process.env.APP_SECRET,
        shop.getShopClientSecret(),
    ),
    version: resp.body.version,
});

await scrapeSingleShop(shopId);

console.log('Fixtures applied successfully');
console.log('User 1 (org owner):', user1.user.email);
console.log('User 2 (org admin):', user2.user.email);
console.log('User 3 (org member):', user3.user.email);
console.log('User 4 (regular user without organization):', user4.user.email);
console.log('Organization:', org.name);
console.log('Shop:', shop.getShopUrl());

console.log('All users have the password "password".');

// source
import { connect } from '@planetscale/database'
import credentials from './credentials.js';

const conn = connect(credentials);


// target
import { readdirSync, readFileSync, unlinkSync } from 'fs'
import Database from 'better-sqlite3';
import knex from 'knex';
import { exec } from 'child_process';
try {
    unlinkSync('sqlite.db');
} catch (e) {}
const sqlite = new Database('sqlite.db');

for (const file of readdirSync('../api/drizzle')) {
    if (file.endsWith('.sql')) {
        sqlite.exec(readFileSync(`../api/drizzle/${file}`).toString());
    }
}

const sqliteKnex = knex({
    client: 'better-sqlite3',
    connection: {
        filename: 'sqlite.db',
    },
    useNullAsDefault: false,
})

const structure = {
    user: {
        id: {name: 'id'},
        username: {name: 'displayName'},
        email: {name: 'email'},
        password: {name: 'password'},
        verify_code: {name: 'verify_code'},
        created_at: {name: 'created_at', date: true},
    },

    user_notification: {
        id: {name: 'id'},
        user_id: {name: 'user_id'},
        key: {name: 'key'},
        level: {name: 'level'},
        title: {name: 'title'},
        message: {name: 'message'},
        link: {name: 'link'},
        read: {name: 'read'},
        created_at: {name: 'created_at', date: true},
    },

    team: {
        id: {name: 'id'},
        name: {name: 'name'},
        owner_id: {name: 'owner_id'},
        created_at: {name: 'created_at', date: true},
    },

    user_to_team: {
        user_id: {name: 'user_id'},
        team_id: {name: 'organization_id'},
    },

    shop: {
        id: {name: 'id'},
        team_id: {name: 'organization_id'},
        name: {name: 'name'},
        status: {name: 'status'},
        url: {name: 'url'},
        favicon: {name: 'favicon'},
        client_id: {name: 'client_id'},
        client_secret: {name: 'client_secret'},
        shopware_version: {name: 'shopware_version'},
        last_scraped_at: {name: 'last_scraped_at', date: true},
        last_scraped_error: {name: 'last_scraped_error'},
        ignores: {name: 'ignores'},
        shop_image: {name: 'shop_image'},
        last_updated: {name: 'last_updated'},
        created_at: {name: 'created_at', date: true},
    },

    shop_changelog: {
        id: {name: 'id'},
        shop_id: {name: 'shop_id'},
        extensions: {name: 'extensions'},
        old_shopware_version: {name: 'old_shopware_version'},
        new_shopware_version: {name: 'new_shopware_version'},
        date: {name: 'date', date: true},
    },

    shop_pagespeed: {
        id: {name: 'id'},
        shop_id: {name: 'shop_id'},
        performance: {name: 'performance'},
        accessibility: {name: 'accessibility'},
        bestpractices: {name: 'best_practices'},
        created_at: {name: 'created_at', date: true},
    },
}

try {
    unlinkSync('import.sql');
} catch (e) {}


for (let [table, columns] of Object.entries(structure)) {
    const data = await conn.execute(`SELECT * FROM ${table}`);

    table = table.replace('team', 'organization');

    for (const row of data.rows) {
        const mapped = {};

        for (const [before, after] of Object.entries(columns)) {
            if (typeof row[before] === 'undefined')  {
                throw new Error(`Column ${before} not found in ${table}#${row.id}`);
            }

            if (after.date) {
                row[before] = new Date(row[before]).getTime() / 1000;
            }

            mapped[after.name] = row[before];
        }

        await sqliteKnex
            .insert([mapped])
            .into(table);

        console.log(`Migrated ${table}#${row.id}`);
    }

    
    exec(`sqlite3 sqlite.db .dump | grep '^INSERT INTO ${table} ' >> import.sql`)
}

process.exit(0);

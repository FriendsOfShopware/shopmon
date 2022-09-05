import { Connection } from "@planetscale/database/dist";

interface CreateShopRequest {
    team_id: string;
    name: string;
    version: string;
    shop_url: string;
    client_id: string;
    client_secret: string;
}

export default class Shops {
    static async createShop(con: Connection, params: CreateShopRequest): Promise<string> {
        try {
            const result = await con.execute('INSERT INTO shop (team_id, name, url, client_id, client_secret, created_at, shopware_version) VALUES (?, ?, ?, ?, ?, NOW(), ?)', [
                params.team_id,
                params.name,
                params.shop_url,
                params.client_id,
                params.client_secret,
                params.version,
            ]);

            return result.insertId as string;
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        } catch (e: any) {
            if (e.error?.code == 'ALREADY_EXISTS') {
                throw new Error('Shop already exists.');
            }

            throw e;
        }
    }

    static async deleteShop(con: Connection, id: number): Promise<void> {
        await con.execute('DELETE FROM shop WHERE id = ?', [
            id
        ]);

        await con.execute('DELETE FROM shop_scrape_info WHERE shop_id = ?', [
            id
        ]);
    }
}
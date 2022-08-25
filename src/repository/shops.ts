import { getConnection } from "../db";

interface CreateShopRequest {
    team_id: string;
    name: string;
    version: string;
    shop_url: string;
    client_id: string;
    client_secret: string;
}

export default class Shops {
    static async createShop(params: CreateShopRequest): Promise<string> {
        const result = await getConnection().execute('INSERT INTO shop (team_id, name, url, client_id, client_secret, created_at, shopware_version) VALUES (?, ?, ?, ?, ?, NOW(), ?)', [
            params.team_id,
            params.name,
            params.shop_url,
            params.client_id,
            params.client_secret,
            params.version,
        ]);
    
        if (result.error?.code == 'ALREADY_EXISTS') {
            throw new Error('Shop already exists.');
        }

        return result.insertId as string;
    }

    static async deleteShop(id: number): Promise<void> {
        await getConnection().execute('DELETE FROM shop WHERE id = ?', [
            id
        ]);

        await getConnection().execute('DELETE FROM shop shop_scrape_info shop_id = ?', [
            id
        ]);
   }
}
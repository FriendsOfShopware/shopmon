import { getKv } from "../../db";

export async function validateToken(req: Request): Promise<Response|void> {
    if (req.headers.get('token') === null) {
        return new Response('Invalid token', {
            status: 401
        });
    }

    const token = req.headers.get('token');

    const result = await getKv().get(token as string);

    if (result === null) {
        return new Response('Invalid token', {
            status: 401
        });
    }

    const data = JSON.parse(result);

    req.userId = data.id;
}
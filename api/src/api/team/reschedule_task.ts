import { ErrorResponse, NoContentResponse } from "../common/response";
import { SimpleShop, HttpClient } from "@friendsofshopware/app-server-sdk"
import { decrypt } from "../../crypto";
import { getConnection, schema } from "../../db";
import { eq } from 'drizzle-orm';

export async function reScheduleTask(req: Request, env: Env): Promise<Response> {
    const { shopId, taskId } = req.params as { shopId?: string, taskId?: string };

    if (typeof shopId !== "string") {
        return new ErrorResponse('Missing shopId', 400);
    }

    if (typeof taskId !== "string") {
        return new ErrorResponse('Missing TaskId', 400);
    }

    const con = getConnection(env);
    const shopData = await con.query.shop.findFirst({
        columns: {
            url: true,
            client_id: true,
            client_secret: true,
        },
        where: eq(schema.shop.id, parseInt(shopId))
    });

    if (shopData === undefined) {
        return new ErrorResponse("No shop with this ID found", 404);
    }

    const clientSecret = await decrypt(env.APP_SECRET, shopData.client_secret);
    const shop = new SimpleShop('', shopData.url, '');
    shop.setShopCredentials(shopData.client_id, clientSecret);
    const client = new HttpClient(shop);

    try {
        const nextExecutionTime: string = new Date().toISOString();
        await client.patch(`/scheduled-task/${taskId}`, {
            'status': 'scheduled',
            'nextExecutionTime': nextExecutionTime
        });

        const scrapeResult = await con.query.shopScrapeInfo.findFirst({
            columns: {
                scheduled_task: true,
            },
            where: eq(schema.shopScrapeInfo.shop, parseInt(shopId))
        });

        // If there is no scrape result, we don't need to update the scheduled task
        if (scrapeResult === undefined) {
            return new NoContentResponse();
        }

        const scheduledTasks = JSON.parse(scrapeResult.scheduled_task || '{}');

        for (const task of scheduledTasks) {
            if (task.id === taskId) {
                task.status = 'scheduled';
                task.nextExecutionTime = nextExecutionTime;
                task.overdue = false;
            }
        }

        await con
            .update(schema.shopScrapeInfo)
            .set({ scheduled_task: JSON.stringify(scheduledTasks) })
            .where(eq(schema.shopScrapeInfo.shop, parseInt(shopId)))
            .execute();

    } catch (e: any) {
        return new ErrorResponse(e.response.body.errors[0].detail, e.response.statusCode);
    }

    return new NoContentResponse();
}

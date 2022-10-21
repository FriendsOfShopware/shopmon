import {ErrorResponse, NoContentResponse} from "../common/response";
import {Shop} from "shopware-app-server-sdk/shop";
import {decrypt} from "../../crypto";
import {HttpClient} from "shopware-app-server-sdk/component/http-client";
import {getConnection} from "../../db";
import task from "../../object/status/checks/task";

export async function reScheduleTask(req: Request, env: Env): Promise<Response> {
    const {shopId, taskId} = req.params as { shopId?: string, taskId?: string };

    if (typeof shopId !== "string") {
        return new ErrorResponse('Missing shopId', 400);
    }

    if (typeof taskId !== "string") {
        return new ErrorResponse('Missing TaskId', 400);
    }

    const con = getConnection(env);
    const shopData = await con.execute('SELECT url,client_id,client_secret FROM shop WHERE id = ?', [shopId]);

    if (!shopData.rows.length) {
        return new ErrorResponse("No shop with this ID found", 404);
    }
    const clientSecret = await decrypt(env.APP_SECRET, shopData.rows[0].client_secret);
    const client = new HttpClient(new Shop('', shopData.rows[0].url, '', shopData.rows[0].client_id, clientSecret));

    try {
        const nextExecutionTime: string = new Date().toISOString();
        await client.patch('/scheduled-task/' + taskId, {
            'status': 'scheduled',
            'nextExecutionTime': nextExecutionTime
        });

        const scrapeResult = await con.execute('SELECT scheduled_task FROM shop_scrape_info WHERE shop_id = ?', [shopId]);

        const scheduledTasks = JSON.parse(scrapeResult.rows[0].scheduled_task);

        for (const task of scheduledTasks) {
            if (task.id === taskId) {
                task.status = 'scheduled';
                task.nextExecutionTime = nextExecutionTime;
                task.overdue = false;
            }
        }

        await con.execute('UPDATE shop_scrape_info SET scheduled_task = ? WHERE shop_id = ?', [JSON.stringify(scheduledTasks), shopId]);

    } catch (e) {
        return new ErrorResponse(e.response.body.errors[0].detail, e.response.statusCode);
    }

    return new NoContentResponse();
}
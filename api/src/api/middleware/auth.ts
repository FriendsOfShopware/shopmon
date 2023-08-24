import { Toucan } from "toucan-js";
import { ErrorResponse } from "../common/response";

export async function validateToken(req: Request, env: Env, ctx: ExecutionContext, sentry: Toucan): Promise<Response|void> {
    if (req.headers.get('authorization') === null) {
        return new ErrorResponse('Invalid token', 401);
    }

    let token = req.headers.get('authorization');

    if (token?.toLocaleLowerCase().indexOf('bearer') === 0) {
        token = req.headers.get('authorization')?.split(' ')[1] as string;
    }

    if (token?.indexOf('u-') !== 0) {
        return new ErrorResponse('Invalid token', 401);
    }

    const result = await env.kvStorage.get(token as string);

    if (result === null) {
        return new ErrorResponse('Invalid token', 401);
    }

    const data = JSON.parse(result);

    req.userId = data.id;

    sentry.setUser({ id: req.userId })
}
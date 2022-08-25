export class NoContentResponse extends Response {
    constructor(headers = {}) {
        super(null, { status: 204, headers });
    }
}

export class ErrorResponse extends Response {
    constructor(message: string, statusCode = 500) {
        super(
            JSON.stringify({message}),
        {
            status: statusCode,
            headers: {
                "content-type": 'application/json',
            }
        });
    }
}

export class JsonResponse extends Response {
    constructor(body: any, status: number = 200, headers: {} = {}) {
        super(JSON.stringify(body), {
            status,
            headers: {
                "content-type": 'application/json',
                ... headers
            }
        });
    }
}
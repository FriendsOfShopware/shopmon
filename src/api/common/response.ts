export class NoContentResponse extends Response {
    constructor() {
        super('', { status: 204 });
    }
}

export class ErrorResponse extends Response {
    constructor(message: string) {
        super(
            JSON.stringify({message}),
        {
            status: 500,
            headers: {
                "content-type": 'application/json',
            }
        });
    }
}

export class JsonResponse extends Response {
    constructor(body: any, status = 200) {
        super(JSON.stringify(body), {
            status,
            headers: {
                "content-type": 'application/json',
            }
        });
    }
}
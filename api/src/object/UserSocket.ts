import { WebsocketMessage } from '../../../frontend/src/types/notification';
import type { Bindings } from '../router';

export class UserSocket implements DurableObject {
    state: DurableObjectState;
    env: Bindings;
    sessions: WebSocket[];

    constructor(state: DurableObjectState, env: Bindings) {
        this.state = state;
        this.env = env;

        this.sessions = [];
    }

    async fetch(req: Request): Promise<Response> {
        const { pathname } = new URL(req.url);

        if (pathname === '/api/ws') {
            if (req.headers.get('Upgrade') !== 'websocket') {
                return new Response('Expected websocket', { status: 400 });
            }

            const { 0: client, 1: server } = new WebSocketPair();

            this.sessions.push(server);

            server.accept();

            const closeOrErrorHandler = () => {
                this.sessions = this.sessions.filter(
                    (member) => member !== server,
                );
            };

            // Cleanup when the client closes the connection
            server.addEventListener('close', closeOrErrorHandler);
            server.addEventListener('error', closeOrErrorHandler);

            return new Response(null, { status: 101, webSocket: client });
        }

        if (pathname === '/api/send') {
            const data = await req.text();

            for (const session of this.sessions) {
                session.send(data);
            }

            return new Response(null, { status: 202 });
        }

        return new Response('Not found', { status: 404 });
    }
}

export const UserSocketHelper = {
    async sendNotification(
        namespace: DurableObjectNamespace,
        userId: string,
        notification: WebsocketMessage,
    ) {
        await namespace
            .get(namespace.idFromName(userId))
            .fetch('http://localhost/api/send', {
                method: 'POST',
                body: JSON.stringify(notification),
            });
    },
};

import { useAuthStore } from '@/stores/auth.store';
import { trpcClient } from './trpc';

export const fetchWrapper = {
    get: request('GET'),
    post: request('POST'),
    put: request('PUT'),
    patch: request('PATCH'),
    delete: request('DELETE')
};


interface CustomRequestInit extends RequestInit {
    headers: Record<string, string>
}

function request(method: string) {
    return (url: string, body: null | string | object = null, headers = {}) => {
        url = '/api' + url;
        const requestOptions: CustomRequestInit = {
            method,
            headers: { ...authHeader(url), ...headers },
            body: null,
        };
        if (body) {
            requestOptions.headers['Content-Type'] = 'application/json';
            requestOptions.body = JSON.stringify(body);
        }
        return fetch(url, requestOptions).then(handleResponse(url, requestOptions));
    }
}

function authHeader(url: string): object {
    const { access_token } = useAuthStore();
    const isApiUrl = url.startsWith('/api');
    if (access_token && isApiUrl) {
        return { Authorization: `Bearer ${access_token}` };
    } else {
        return {};
    }
}

function handleResponse(url: RequestInfo | URL, requestOptions: CustomRequestInit) {
    return async (response: Response): Promise<any> => {
        const isJson = response.headers?.get('content-type')?.includes('application/json');
        const data = isJson ? await response.json() : null;

        if (!response.ok) {
            const { refresh_token, logout, setAccessToken } = useAuthStore();
            if (response.status === 401) {
                if (refresh_token) {
                    try {
                        const result = await trpcClient.auth.refreshToken.mutate(refresh_token)

                        setAccessToken(result.accessToken);

                        requestOptions.headers['Authorization'] = `Bearer ${result.accessToken}`;

                        return fetch(url, requestOptions).then(handleResponse(url, requestOptions));
                    } catch (e) {
                        logout();

                        return Promise.reject('Refresh token expired');
                    }
                } else {
                    logout();
                }
            }

            // get error message from body or default to response status
            const error = (data && data.message) || response.status;
            return Promise.reject(error);
        }

        return data;
    }
}

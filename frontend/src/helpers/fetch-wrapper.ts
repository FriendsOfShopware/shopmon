import {useAuthStore} from '@/stores/auth.store';

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
    return (url: string, body: null|string|object = null, headers = {}) => {
        url = '/api' + url;
        const requestOptions: CustomRequestInit = {
            method,
            headers: {... authHeader(url), ...headers},
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
                    const refreshResponse = await fetch('/api/auth/token', {
                        method: 'POST',
                        headers: {
                            'Content-Type': 'application/json',
                        },
                        body: JSON.stringify({
                            client_id: 'shopmon',
                            grant_type: 'refresh_token',
                            refresh_token: refresh_token,
                        }),
                    });
    
                    if (refreshResponse.ok) {
                        const refreshData = await refreshResponse.json();
    
                        setAccessToken(refreshData.access_token);
    
                        requestOptions.headers['Authorization'] = `Bearer ${refreshData.access_token}`;
    
                        return fetch(url, requestOptions).then(handleResponse(url, requestOptions));
                    } else {
                        logout();

                        return Promise.reject('Refresh token expired');
                    }
                } else {
                    logout();
                }
            }

            // get error message from body or default to response status
            const error = (data && data.message) || await response.text() || response.status;
            return Promise.reject(error);
        }

        return data;
    }
}

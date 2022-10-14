export async function shopImage(req: Request, env: Env): Promise<Response> {
    const { uuid } = req.params;

    if (uuid.indexOf('/') !== -1) {
        return new Response('', {
            status: 404
        });
    }

    const file = await env.FILES.get(`pagespeed/${uuid}/screenshot.jpg`)

    if (uuid.indexOf('/') !== -1) {
        return new Response('', {
            status: 404
        });
    }

    return new Response(file?.body, {
        status: 200,
        headers: {
            "content-type": "image/jpeg"
        }
    });
}
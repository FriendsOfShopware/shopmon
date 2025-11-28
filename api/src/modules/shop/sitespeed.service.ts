const sitespeedServiceUrl =
    process.env.APP_SITESPEED_ENDPOINT || 'http://localhost:3001';
const sitespeedPrefix = process.env.APP_SITESPEED_PREFIX || 'local-';

export function getReportName(shopId: number) {
    return `${sitespeedPrefix}shop-${shopId}`;
}

export function getScreenshotUrl(shopId: number) {
    const recordId = getReportName(shopId);
    return `${sitespeedServiceUrl}/screenshot/${recordId}`;
}

export function getReportUrl(shopId: number) {
    const recordId = getReportName(shopId);
    return `${sitespeedServiceUrl}/result/${recordId}/index.html`;
}

export async function runSitespeedReport(shopId: number, urls: string[]) {
    const recordId = getReportName(shopId);

    const response = await fetch(
        `${sitespeedServiceUrl}/api/result/${recordId}`,
        {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                Authorization: `Bearer ${process.env.APP_SITESPEED_API_KEY || 'secret'}`,
            },
            body: JSON.stringify({
                urls: urls,
            }),
        },
    );

    return (await response.json()) as {
        ttfb: number;
        fullyLoaded: number;
        largestContentfulPaint: number;
        firstContentfulPaint: number;
        cumulativeLayoutShift: number;
        transferSize: number;
    };
}

export async function deleteSitespeedReport(shopId: number) {
    const recordId = getReportName(shopId);

    await fetch(`${sitespeedServiceUrl}/api/result/${recordId}`, {
        method: 'DELETE',
        headers: {
            Authorization: `Bearer ${process.env.APP_SITESPEED_API_KEY || 'secret'}`,
        },
    });
}

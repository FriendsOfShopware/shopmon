import * as cron from 'node-cron';
import { lockCleanupJob } from './jobs/lockCleanup.js';
import { pagespeedScrapeJob } from './jobs/pagespeedScrape.js';
import { shopScrapeJob } from './jobs/shopScrape.js';

console.log('Registered cron jobs...');

// Run shop scrape every hour
cron.schedule('0 * * * *', async () => {
    console.log('Running shop scrape job...');
    try {
        await shopScrapeJob();
    } catch (error) {
        console.error('Shop scrape job failed:', error);
    }
});

// Run pagespeed scrape every 24 hours at 2 AM
cron.schedule('0 2 * * *', async () => {
    console.log('Running pagespeed scrape job...');
    try {
        await pagespeedScrapeJob();
    } catch (error) {
        console.error('Pagespeed scrape job failed:', error);
    }
});

// Run lock cleanup job every day at 3 AM
cron.schedule('0 3 * * *', async () => {
    console.log('Running lock cleanup job...');
    try {
        await lockCleanupJob();
    } catch (error) {
        console.error('Lock cleanup job failed:', error);
    }
});

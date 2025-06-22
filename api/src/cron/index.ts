import * as cron from 'node-cron';
import { lockCleanupJob } from './jobs/lockCleanup.ts';
import { shopScrapeJob } from './jobs/shopScrape.ts';
import { scrapeSitespeedForAllShops } from './jobs/sitespeedScrape.ts';

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

// Run sitespeed scrape once daily at 3 AM
cron.schedule('0 3 * * *', async () => {
    console.log('Running sitespeed scrape job...');
    try {
        await scrapeSitespeedForAllShops();
    } catch (error) {
        console.error('Sitespeed scrape job failed:', error);
    }
});

// Run lock cleanup job every day at 4 AM
cron.schedule('0 4 * * *', async () => {
    console.log('Running lock cleanup job...');
    try {
        await lockCleanupJob();
    } catch (error) {
        console.error('Lock cleanup job failed:', error);
    }
});

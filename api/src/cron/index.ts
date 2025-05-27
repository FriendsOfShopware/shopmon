import * as cron from 'node-cron';
import { shopScrapeJob } from './jobs/shopScrape.js';
import { pagespeedScrapeJob } from './jobs/pagespeedScrape.js';
import { sessionCleanupJob } from './jobs/sessionCleanup.js';
import { passwordResetCleanupJob } from './jobs/passwordResetCleanup.js';

console.log('Starting cron jobs...');

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

// Run session cleanup every hour
cron.schedule('0 * * * *', async () => {
  console.log('Running session cleanup job...');
  try {
    await sessionCleanupJob();
  } catch (error) {
    console.error('Session cleanup job failed:', error);
  }
});

// Run password reset cleanup every hour
cron.schedule('0 * * * *', async () => {
  console.log('Running password reset cleanup job...');
  try {
    await passwordResetCleanupJob();
  } catch (error) {
    console.error('Password reset cleanup job failed:', error);
  }
});

console.log('Cron jobs scheduled successfully');

// Keep the process running
process.on('SIGINT', () => {
  console.log('Shutting down cron jobs...');
  process.exit(0);
});

import { onSchedule } from './cron/schedule';
import router from './router'

addEventListener('fetch', (event) => {
  event.respondWith(router.handle(event.request))
})

addEventListener('scheduled', (event) => {
  event.waitUntil(onSchedule())
});
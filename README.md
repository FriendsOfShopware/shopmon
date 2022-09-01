# Shop Monitoring

Shopmon is an hosted application from FriendsOfShopware to manage multiple Shopware instances.

## Features

Overview of all your Shopware instances to see:

- Shopware Version and Security Updates
- Show all installed extension and extension updates

## Ideas

- Alerting based on some criterias like one day no orders has been created
- Track the queue stats and show nice diagrams


## Requirements

- Cloudflare Worker
- Planet Scale Database

## Install

### Frontend

- Go to `frontend`
- Run `npm run dev`
- Open `localhost:3000` to see the page

### Hosting own API

- Install [wrangler](https://developers.cloudflare.com/workers/wrangler/get-started/)
- Create a own Planet Scale Account + Database
- Create [a serverless access key](https://planetscale.com/blog/introducing-the-planetscale-serverless-driver-for-javascript)\
- Import schema `db.sql` using some MySQL cli's
- Go to `api`
- Create a file `.dev.vars`

```text
DATABASE_HOST=aws.connect.psdb.cloud
DATABASE_USER=USER
DATABASE_PASSWORD=PW
MAIL_URL=https://localhost:3000
MAIL_SECRET=foooo
```
- Emails will be not really send, copy the code from the API
- Run `npm install`
- Run `wrangler dev --port 5000 --local`
- Adjust `vite.config.js` from frontend to point to `http://localhost:5000`


## License

MIT
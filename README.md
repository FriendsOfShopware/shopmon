# Shop Monitoring

Shopmon is a hosted application from FriendsOfShopware to manage multiple Shopware instances.

* Credentials are saved on a [Cloudflare D1](https://developers.cloudflare.com/d1/) SQLite database
  * Client secret are encrypted by [web crypto api](https://developer.mozilla.org/en-US/docs/Web/API/Web_Crypto_API) outside the database
* API runs on Cloudflare workers (serverless)
* Mails are sent using [MailChannels](https://www.mailchannels.com)

## Features

Overview of all your Shopware instances to see:

- Shopware Version and Security Updates
- Show all installed extension and extension updates
- Show info on scheduled tasks and queue
- Run a daily check with pagespeed to see decreasing performance
- Clear shop cache

## Requirements (self hosted)

- Cloudflare Worker aka Wrangler

## Managed / SaaS

https://shopmon.fos.gg

## Setup Local

Requirements: 
  - Node 20 or higher
  - PNPM installed as Package manager or Node Corepack enabled

### Install dependencies

```bash
make setup
```

### Run migrations

```bash
make migrate
```

### Run the app

Run the API and the frontend in local development mode

```bash
make dev
```

### Page speed

If you want to trace the performance of your shop you need to activate the Google Pagespeed API.

- Go to https://developers.google.com/speed/docs/insights/v5/get-started

and create a `.dev.vars` file in `api` folder with your API key like:

```text
PAGESPEED_API_KEY=AIzaSyCWNar-IbOaQT1WX_zfAjUxG01x7xErbSc
APP_SECRET=MZRa9lEjACNhNhw40QXwRZANRx8f1WQa
```

## Configuration

### Disable registration
To disable user registrations set the following variables:

`frontend/.env` To disable the frontend registration route:

```text 
VITE_DISABLE_REGISTRATION=1
```

`api/.dev.vars` To disable the app functionality:
```text 
DISABLE_REGISTRATION=1
```

## License

MIT

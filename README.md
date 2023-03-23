# Shop Monitoring

[![Open in Gitpod](https://gitpod.io/button/open-in-gitpod.svg)](https://gitpod.io/#https://github.com/FriendsOfShopware/shopmon)

Shopmon is a hosted application from FriendsOfShopware to manage multiple Shopware instances.

* Credentials are saved on a [planetscale](https://planetscale.com/) database
  * Clientsecret ist encrypted by [web crypto api](https://developer.mozilla.org/en-US/docs/Web/API/Web_Crypto_API)
* API runs on cloudflare workers (serverless)
* Mails are sent via Mailgun

## Features

Overview of all your Shopware instances to see:

- Shopware Version and Security Updates
- Show all installed extension and extension updates
- Show info on scheduled tasks and queue
- Run a daily check with pagespeed to see decreasing performance
- Clear shop cache

## Requirements (selfhosted)

- Cloudflare Worker
- Planet Scale Database
- Mailgun

## Managed / SaaS

https://shopmon.fos.gg

#### Frontend

> If you want to test around a little locally you can run the frontend on your machine, but data is saved on the managed database by friendsofshopware

- Go to `cd frontend`
- Run  `npm i` and `npm run dev`
- Open `localhost:3000` to see the page


## Install Selfhosted

### Setup

#### Planetscale

- Create an own [Planet Scale Account](https://auth.planetscale.com/sign-up) + Database
- Create [a serverless access key](https://planetscale.com/blog/introducing-the-planetscale-serverless-driver-for-javascript)
  - https://app.planetscale.com/PROJECT/DATABASE/settings/beta-features
- Import schema `db.sql` using some MySQL Client or use the console in planetscale (https://app.planetscale.com/PROJECT/DATABASE/main/console)

```
HOST: eu-central.connect.psdb.cloud
PORT: 3306
USERNAME: check /settings/passwords
PASSWORD: check /settings/passwords
```

#### Cloudflare worker
- Install [wrangler](https://developers.cloudflare.com/workers/wrangler/get-started/)
- Set all secrets in the UI or using wrangler cli
- Adjust `routes` inside `wrangler.toml` to your domain


#### Mailgun
- Create an own [Mailgun Account](https://signup.mailgun.com/new/signup) or copy verify code from Database entries
> Currently restricted to EU API


#### Pagespeed
- Activate [Google Pagespeed API](https://developers.google.com/speed/docs/insights/v5/get-started)

#### Enviroment variables
- Go to `api` create a file `.dev.vars` and fill in the just created credentials of each service

```text
DATABASE_HOST=aws.connect.psdb.cloud
DATABASE_USER=USER
DATABASE_PASSWORD=PW
MAILGUN_KEY=KEY
MAILGUN_DOMAIN=DOMAIN
PAGESPEED_API_KEY=AIzaSyCWNar-IbOaQT1WX_zfAjUxG01x7xErbSc
APP_SECRET=MZRa9lEjACNhNhw40QXwRZANRx8f1WQa
```

#### Frontend

- Go to `cd frontend`
- Run  `npm i` and `npm run dev:local`
- Open `localhost:3000` to see the page

#### Backend / API

- Go to `cd api`
- Run  `npm i` and `npm run dev:local`
> Check your console output for request infos and console.logs

## Using Docker

### Docker Compose

See [docker-compose.yml](docker-compose.yml) for an example.

### Environment Variables

| Variable          | Description              | Default                                                                 |
|-------------------|--------------------------|-------------------------------------------------------------------------|
| MAILGUN_KEY       | Mailgun API Key          | NULL                                                                    |
| MAILGUN_DOMAIN    | Mailgun Domain           | NULL                                                                    |
| PAGESPEED_API_KEY | Google Pagespeed API Key | NULL                                                                    |
| APP_SECRET        | Application secret       | automatically generated and persisted inside /app/data if not specified |

### Volumes

All data is stored in the /app/data folder. You can mount this folder to a volume to persist the data.

### Managing Users

You can manage users via the tool.sh script.

```
Usage: ./tool.sh <command> [args]
Commands:
  add <username> <password> <email> - adds a user to the sqlite database
  del <email> - deletes a user from the sqlite database
  activate <email> - activates a user in the sqlite database
  list - lists all users in the sqlite database
  help - prints this help message
```

Example:

```
./tool.sh add TestUser password test@test.org
```

## License

MIT


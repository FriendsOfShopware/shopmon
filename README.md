# Shop Monitoring

Shopmon is an application from FriendsOfShopware to manage multiple Shopware instances.

* Credentials are saved in a SQLite database
  * Client secrets are encrypted using [web crypto api](https://developer.mozilla.org/en-US/docs/Web/API/Web_Crypto_API) outside the database
* API runs on Bun runtime
* Mails are sent via SMTP

## Features

Overview of all your Shopware instances to see:

- Shopware Version and Security Updates
- Show all installed extension and extension updates
- Show info on scheduled tasks and queue
- Run a daily check with sitespeed to see decreasing performance
- Clear shop cache

## Requirements (self hosted)

> [!NOTE]  
> It's not recommended to self-host this application, we don't give any support for self-hosted installations. Please use the managed version at https://shopmon.fos.gg

- Bun runtime (v1.0 or higher)
- SQLite 3
- Node.js 20+ and PNPM (for building frontend)

## Managed / SaaS

https://shopmon.fos.gg

## Local Installation

### Prerequisites

- Bun
- Node.js 22 or higher

### Step 1: Clone the repository

```bash
git clone https://github.com/FriendsOfShopware/shopmon.git
cd shopmon
```

### Step 2: Install dependencies

```bash
make setup
```

This will install dependencies for both the API and frontend.

### Step 3: Configure environment

Copy the example environment file and configure it:

```bash
cp api/.env.example api/.env
```

Edit `api/.env` with your configuration:

```env
# Database
# Security (generate a secure random string)
APP_SECRET=your-secure-random-string-here

# Email configuration
SMTP_SERVER=smtp.example.com
SMTP_PORT=587
SMTP_USER=your-email@example.com
SMTP_PASS=your-password
SMTP_FROM=noreply@yourdomain.com

# Application
FRONTEND_URL=http://localhost:5173
APP_FILES_DIR=./files

# Monitoring (optional)
SENTRY_DSN=
```

### Step 4: Run database migrations

```bash
make migrate
```

### Step 5: Start the application

For development:

```bash
make dev
```

This will start:
- API server on http://localhost:3000
- Frontend dev server on http://localhost:5173

### Additional services

To develop Shopmon easier, you can start a local mail catcher and a local Shopware installation with:

```
make up
```

## Configuration

### Production Deployment

For production deployment, you can use the provided Docker setup:

1. Build the Docker image:
```bash
docker build -t shopmon .
```

2. Run with docker compose:
```bash
docker compose -f compose.deploy.yml up -d
```

Make sure to:
- Use a strong APP_SECRET
- Configure proper email settings
- Set up persistent volumes for database and uploads
- Configure a reverse proxy (nginx, traefik, etc.) for HTTPS

## Contributing

We welcome contributions! Please see our [Contributing Guidelines](./CONTRIBUTING.md) for details on how to get started.

## License

MIT

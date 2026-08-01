# Shopmon Coder template

This template provisions a complete Shopmon development workspace:

- a persistent Ubuntu workspace with Go, Node.js, `mise`, Docker CLI, GitHub CLI, and VS Code Insiders integration
- an isolated Docker-in-Docker daemon and persistent Docker data for the project Compose stack
- an automatic clone of `FriendsOfShopware/shopmon`
- automatic tool and dependency installation through `mise`
- migrations and fixtures on first start, followed by migrations on later starts
- automatically started API, worker, and frontend development servers
- Coder apps for Shopmon, the API, Mailpit, the demo shop, Jaeger, and development logs

The generated workspace-local `mise` config points PostgreSQL, Redis, SMTP, S3, sitespeed, and the packages service at the workspace's isolated Docker daemon.
Docker-backed web apps are reverse-proxied to the workspace agent's loopback interface so Coder can reach them. The proxy preserves the public host and restores `X-Forwarded-Host` and `X-Forwarded-Proto`, which Shopware uses to resolve its storefront sales channel.
The expected public demo-shop URL is available as `SHOPWARE_APP_URL` and is passed to the demo container as `APP_URL`.

## Requirements

- Coder 2.18 or newer
- a Docker daemon reachable by the Coder provisioner
- support for privileged containers on the Docker host
- a Coder wildcard access URL for the Vite frontend and the other subdomain apps
- approximately 8 GB of free RAM and 25 GB of free disk per active workspace

If Coder itself runs in Docker, mount the host Docker socket into the Coder container and grant it access to the socket. See the [Coder Docker template prerequisites](https://registry.coder.com/templates/coder/docker).

> [!WARNING]
> The template uses a privileged Docker-in-Docker sidecar. Coder documents privileged Docker sidecars as unsuitable for untrusted multi-tenant workloads because a container escape can compromise the host. Use a Coder-supported Sysbox setup when stronger workspace isolation is required.

## Publish the template

Log in to your Coder deployment, then run this command from the repository root:

```bash
coder templates push shopmon \
  --directory coder \
  --variable coder_wildcard_access_url="https://*.apps.coder.example.com"
```

The wildcard value must match the `CODER_WILDCARD_ACCESS_URL` configured on the deployment, including its scheme. If the variable is omitted, Shopmon falls back to `http://localhost:3000` for `FRONTEND_URL`; the app button still requires wildcard app routing, and WebAuthn/OAuth callbacks will not work correctly through the remote URL.

For a non-default Docker socket, also pass:

```bash
--variable docker_socket="unix:///path/to/docker.sock"
```

Create a workspace from the `shopmon` template after the upload completes. The first startup pulls the Compose images and installs dependencies, so it can take several minutes. Later starts reuse the persistent home and Docker volumes.

## Workspace behavior

The workspace starts the development stack automatically. Useful commands inside the workspace are:

```bash
cd ~/shopmon

mise run dev          # run the API, worker, and frontend in the foreground
mise run up           # ensure Compose services are running
mise run load-fixtures
mise run lint
mise run test
```

Development process output is stored at `~/.local/state/shopmon/dev.log` and is available through the **Development Logs** workspace app.

The initial fixture credentials are:

```text
Email: admin@fos.gg
Password: password
```

Deleting a Coder workspace deletes its persistent home and nested Docker volumes. Stopping and restarting a workspace preserves both.

## Validate changes

```bash
terraform -chdir=coder fmt -check
terraform -chdir=coder init -backend=false
terraform -chdir=coder validate
```

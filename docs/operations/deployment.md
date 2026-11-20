# Deploy Token Catalog

Requires a Linux host with systemd, Git, Docker Engine, and Docker Compose
2.24.4+. The updater can use up to 4 GiB of RAM. Go, Node.js, and Chromium run
inside the containers.

This setup publishes the site through a Cloudflare Tunnel (see
[Cloudflare setup](cloudflare.md) for DNS, DNSSEC, and agent discovery). Email
alerts via SMTP are optional.

## 1. Install

Keep the checkout at `/opt/tokens` and run:

```sh
sudo sh deploy/install.sh
```

This creates the data directories, config files (preserving existing settings),
and systemd units without starting them.

## 2. Configure the public URL and tunnel

Set the public collection URL (HTTPS with `/md3`, no trailing slash) in
`/etc/tokens/server.env`:

```sh
sudoedit /etc/tokens/server.env
```

```dotenv
TOKENS_PUBLIC_ORIGIN=https://tokencatalog.dev/md3
```

An empty value disables search indexing.

Create a dedicated Cloudflare Tunnel pointing `tokencatalog.dev` (empty path) to
HTTP service `tokens:8080`, then save its token in `/etc/tokens/cloudflared.env`:

```sh
sudoedit /etc/tokens/cloudflared.env
```

```dotenv
TUNNEL_TOKEN='YOUR_TUNNEL_TOKEN'
```

## 3. Configure email alerts (optional)

Edit `/etc/tokens/update.env` (leave empty to disable email):

```sh
sudoedit /etc/tokens/update.env
```

```dotenv
TOKENS_SMTP_HOST=smtp.example.com
TOKENS_SMTP_PORT=465
TOKENS_SMTP_USERNAME=sender@example.com
TOKENS_SMTP_PASSWORD='YOUR_SMTP_PASSWORD'
TOKENS_SMTP_FROM=sender@example.com
TOKENS_SMTP_TO=recipient@example.com
```

Port 465 uses TLS; port 587 uses STARTTLS. Keep `/etc/tokens` private.

## 4. Build and test notifications

```sh
sudo docker compose -f deploy/compose.yaml --profile update build --pull
sudo sh deploy/update.sh test-notification
```

Skip the notification test if email is disabled.

## 5. Capture and approve the first publication

```sh
sudo sh deploy/update.sh run
sudo sh deploy/update.sh status
```

The first capture stops at `needs_review` with the non-blocking
`initial_baseline` issue. Inspect the report using its run ID, then approve if
the counts look right:

```sh
sudo sh deploy/update.sh inspect "<run-id>"
sudo sh deploy/update.sh approve "<run-id>"
```

Do not approve a failed or incomplete capture.

## 6. Start the site and weekly updates

```sh
sudo systemctl enable --now tokens.service tokens-update.timer
sudo docker compose -f deploy/compose.yaml ps
sudo systemctl list-timers --all tokens-update.timer
```

`tokens.service` showing `active (exited)` is normal—it starts the Compose stack
and exits.

The default schedule is Sunday at 04:00 `Europe/Stockholm` (plus up to 15 min
jitter). To change it, run `sudo systemctl edit tokens-update.timer`, override
`OnCalendar=`, and restart the timer.

## Update the application

Pause scheduled updates and wait for any active run to finish:

```sh
sudo systemctl stop tokens-update.timer
sudo systemctl is-active tokens-update.service
```

Then rebuild and restart:

```sh
git pull
sudo sh deploy/install.sh
sudo docker compose -f deploy/compose.yaml --profile update build --pull
sudo systemctl restart tokens.service
sudo systemctl start tokens-update.timer
```

`tokens.service` rebuilds publication assets from the new image before starting
the server. Settings in `/etc/tokens` persist across updates.

## Browser tools

The server image bundles Material Motion Lab at `/tools/motion-lab/` (built from
`sojupie/material-motion-lab`) and Carousel playground at
`/tools/carousel-playground/` (built from `sojupie/carousel-motion-lab`) as
static files served through the same `tokens:8080` tunnel. `/tools/motion/`
redirects to `/tools/motion-lab/`.

- Health checks verify both tool entry pages alongside the catalog.
- GitHub deploys resolve each tool's `main` branch commit before building. For
  manual builds, set `TOKENS_MOTION_SOURCE` and `TOKENS_PLAYGROUND_SOURCE`
  (`<repo-url>#<commit-sha>` or local paths) in a Compose `--env-file`.
- `TOKENS_MOTION_URL` and `TOKENS_MOTION_LAB_URL` override navigation links in
  `/etc/tokens/server.env` (defaults match the bundled paths; reapply overrides
  after `configure-public-origin`).

## Manage from GitHub Actions

Run **Actions → Manage Token Catalog → Run workflow** on `main`:

| Command | Action |
| --- | --- |
| `status` | Show containers, next update timer, and saved run reports |
| `logs` | Show recent website, tunnel, and update logs |
| `deploy` | Pull `main`, build catalog and tool images, build assets, and replace changed containers |
| `restart` | Restart website and tunnel without rebuilding or crawling |
| `crawl` | Run the update service with alerts and restart after publication |
| `inspect` | Show a saved run report (requires run ID) |
| `approve` | Publish a reviewed capture and restart (requires run ID) |
| `retry` | Retry a saved capture and restart if it publishes (requires run ID) |
| `test-notification` | Send a test email or webhook |
| `usage-report` | Generate a 7-day usage report in the summary and email it (leaves weekly checkpoint unchanged) |

Set `TOKENS_PUBLIC_ORIGIN=https://tokencatalog.dev/md3` as a **variable** (not a
secret) under **Settings → Environments → production → Environment variables**.
The deploy job writes it to `/etc/tokens/server.env` before restarting.

- **Push deploy:** A push to `main` whose tip commit starts with
  `Deploy production:` triggers deployment automatically.
- **Fresh capture on deploy:** Select **Capture fresh tokens during deploy** to
  crawl and build assets before swapping containers. If the capture stops at
  `needs_review`, inspect it and rerun `deploy` with that run ID.
- **Safety:** Deploy refuses dirty working trees, local commits, or rollbacks,
  pauses `tokens-update.timer` while running, and prunes unused Docker build
  cache above ~20 GB (`sudo docker builder prune --all --force --keep-storage 20GB`
  for manual runs).

### Self-hosted runner setup

1. In **Settings → Actions → Runners → New self-hosted runner** (Linux / x64),
   install into `~/actions-runner-tokens` as the `/opt/tokens` owner with
   `--name tokens-prod --labels tokens-prod`.
2. Start the runner service:

   ```sh
   sudo ./svc.sh install "$(id -un)"
   sudo ./svc.sh start
   sudo ./svc.sh status
   ```

3. Restrict the `production` environment to `main` and allow only the required
   sudo commands in `/etc/sudoers.d/tokens-actions` (replace `runner` with the
   runner's username; never run pull requests on this runner):

   ```sudoers
   runner ALL=(root) NOPASSWD: /usr/bin/docker, /usr/bin/sh /opt/tokens/deploy/install.sh, /usr/bin/sh /opt/tokens/deploy/update.sh *, /usr/bin/systemctl stop tokens-update.timer, /usr/bin/systemctl start tokens-update.timer, /usr/bin/systemctl start tokens-update.service, /usr/bin/systemctl restart tokens.service, /usr/bin/journalctl -u tokens-update.service -u tokens-update-failure.service -n 100 --no-pager
   ```

## Generated publication assets

Approved generations stay immutable. The offline `assets` Compose service reads
`current`, builds all six DTCG/CSS contexts and the site theme into a new
`.site-assets-*` directory, validates required tokens and embedded CSS bindings,
and atomically swaps `/var/lib/tokens/site-assets`.

`tokens.service`, weekly updates, and GitHub deploys run this automatically. For
a manual data update:

```sh
sudo sh deploy/update.sh build-assets
sudo docker compose -f deploy/compose.yaml restart tokens
```

If a build fails, the previous `site-assets` link and running server stay in
place; fix the build, rebuild assets, and restart. Verify after startup:

```sh
curl --fail https://tokencatalog.dev/md3/downloads/dtcg/light-default/tokens.tokens.json -o /tmp/tokens.tokens.json
curl --fail https://tokencatalog.dev/md3/downloads/dtcg/light-default/tokens.css -o /tmp/tokens.css
curl --fail https://tokencatalog.dev/md3/downloads/dtcg/light-default/report.json -o /tmp/tokens-report.json
```

Remove old `.site-assets-*` directories manually after verifying a restart,
keeping the active target and one previous build.

## Usage reports and README graphic

The server retains 90 days of hourly aggregate totals in
`/var/lib/tokens-usage/usage.sqlite` (no IPs, cookies, query text, or request
bodies; health checks and static hits excluded; edge cache hits don't reach
origin).

```sh
sudo sh deploy/usage.sh --since 7d
sudo sh deploy/usage.sh --since 7d --send
sudo sh deploy/usage.sh --since 7d --svg > usage.svg
```

- **Weekly email digest:** Enable `sudo systemctl enable --now tokens-usage.timer`
  (runs Mondays at 09:00 `Europe/Stockholm`, requires SMTP in `update.env`).
- **README usage graphic:** **Actions → Update usage graphic** runs Mondays at
  09:15 UTC on `tokens-prod` and commits `docs/images/usage.svg` to `main`.

## Troubleshooting

```sh
sudo docker compose -f deploy/compose.yaml ps -a
sudo docker compose -f deploy/compose.yaml logs --tail=80 tokens cloudflared
sudo journalctl -u tokens-update.service -u tokens-update-failure.service -n 100 --no-pager
```

- **Website missing:** Approve the first capture, then start `tokens.service`.
- **Tunnel errors:** Check `TUNNEL_TOKEN` and the `tokens:8080` service URL.
- **Exit 137:** Out-of-memory kill; check kernel logs (updater limit is 4 GiB).
- **Chromium errors:** Run `sudo sh deploy/test-containers.sh` without disabling the sandbox.
- **No scheduled runs:** Enable `tokens-update.timer`.
- **No email:** Run `test-notification` and check `/etc/tokens/update.env`.

See [updates and recovery](updates.md) for review, retries, and backups, and
[Cloudflare setup](cloudflare.md) for DNS-AID, DNSSEC, and bot/caching settings.

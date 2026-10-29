# Deploy Token Catalog

Use a Linux host with systemd, Git, Docker Engine and Docker Compose 2.24.4 or
newer. The updater can use up to 4 GiB of RAM. Go, Node.js and Chromium run inside the
images; they do not need to be installed on the host.

This setup publishes the site through a Cloudflare Tunnel. Have a Cloudflare
account and a domain configured there before starting. Email alerts need an
SMTP account; they are optional.

## 1. Install

Keep the checkout at `/opt/tokens`. Run these commands from that directory:

```sh
sudo sh deploy/install.sh
```

The installer creates the data directory and configuration files. It preserves
existing settings. It installs the services but does not start them.

## 2. Set the public URL and tunnel

```sh
sudoedit /etc/tokens/server.env
```

Set the URL where the site will be reachable:

```dotenv
TOKENS_PUBLIC_ORIGIN=https://tokencatalog.dev/md3
```

Replace the example with your hostname. Use HTTPS with the collection path `/md3`, without a trailing
slash. This controls canonical URLs and the sitemap. An empty setting disables
search indexing.

Create a Cloudflare Tunnel for this site. Add a public hostname with:

- Hostname: `tokencatalog.dev` (no subdomain). Leave the path empty.
- Service type: HTTP.
- Service URL: `tokens:8080`.

Copy the tunnel token into this file:

```sh
sudoedit /etc/tokens/cloudflared.env
```

```dotenv
TUNNEL_TOKEN='YOUR_TUNNEL_TOKEN'
```

Use a separate tunnel token, not a token belonging to another site.

## 3. Set up email

```sh
sudoedit /etc/tokens/update.env
```

Fill in your email provider's SMTP settings and the recipient for alerts:

```dotenv
TOKENS_SMTP_HOST=smtp.example.com
TOKENS_SMTP_PORT=465
TOKENS_SMTP_USERNAME=sender@example.com
TOKENS_SMTP_PASSWORD='YOUR_SMTP_PASSWORD'
TOKENS_SMTP_FROM=sender@example.com
TOKENS_SMTP_TO=recipient@example.com
```

Port 465 uses TLS. Port 587 uses STARTTLS. Both verify the server certificate.
For Purelymail, use `smtp.purelymail.com`, port `465`, and your full email
address as the username. Use an app password if two-factor authentication is enabled.

Leave all SMTP settings empty to disable email. No webhook is needed for email.
Keep the files in `/etc/tokens` private.

## 4. Build and test email

```sh
sudo docker compose -f deploy/compose.yaml --profile update build --pull
sudo sh deploy/update.sh test-notification
```

Check the recipient's inbox and spam folder. SMTP acceptance alone does not
prove inbox delivery. Skip the test if email is disabled.

## 5. Capture and approve the first data

```sh
sudo sh deploy/update.sh run
sudo sh deploy/update.sh status
```

The first complete capture stops at `needs_review`. This is expected.
It does not start the website or publish the data yet.

Copy the run ID from the status output when prompted:

```sh
printf 'Run ID: '
read -r run_id
sudo sh deploy/update.sh inspect "$run_id"
```

Check the capture counts and `issues`. A complete first capture normally has
only the non-blocking `initial_baseline` issue. Do not approve a failed or
incomplete capture.

If the report is acceptable:

```sh
sudo sh deploy/update.sh approve "$run_id"
```

Approval selects the data to serve. It does not crawl again.

## 6. Start the website and weekly updates

```sh
sudo systemctl enable --now tokens.service tokens-update.timer
sudo docker compose -f deploy/compose.yaml ps
sudo systemctl list-timers --all tokens-update.timer
```

The website should be healthy, the tunnel should be running, and the timer
should show its next run. Open the public URL. `tokens.service` showing
`active (exited)` is normal: it starts the containers and then finishes.

The default schedule is Sunday at 04:00 in Europe/Stockholm, plus up to
15 minutes of delay. A missed run is caught up when the timer starts again.
To change the schedule, run `sudo systemctl edit tokens-update.timer` and add
an override with the desired time and timezone:

```ini
[Timer]
OnCalendar=
OnCalendar=Sun *-*-* 04:00:00 Europe/Stockholm
```

Then run `sudo systemctl restart tokens-update.timer`. Overrides survive
reinstalling the service files.

## Update the application

Pause scheduled updates and wait for a running update to finish:

```sh
sudo systemctl stop tokens-update.timer
sudo systemctl is-active tokens-update.service
```

Continue when the update service is no longer active or activating:

```sh
git pull
sudo sh deploy/install.sh
sudoedit /etc/tokens/server.env
sudo docker compose -f deploy/compose.yaml --profile update build --pull
sudo systemctl restart tokens.service
sudo systemctl start tokens-update.timer
```

Server startup first builds publication assets using the matching application
image. This updates the application, token build dependencies and Chromium. Weekly data captures do not update
the images. Settings in `/etc/tokens` survive installation and image rebuilds.
Before restarting, make sure `server.env` contains the public URL. Older
installations may not have had this file.

The Docker build keeps Go's module and compiled-package caches between builds.
Chromium is installed in a layer that application changes do not invalidate.
The first build after changing the Dockerfile or base images can still be slow.

## Browser tools

The server image includes Material Motion Lab at `/tools/motion/` and Carousel
playground at `/tools/motion-lab/`, outside the catalog's `/md3` mount. The existing
Cloudflare Tunnel serves all three destinations through `tokens:8080`. No additional
tunnel, container or processing service is needed.

Docker builds Material Motion Lab from `sojupie/material-motion-lab` and Carousel
playground from `sojupie/carousel-motion-lab`. Both use the catalog checkout's shared
navigation and component code. The tools compile to static files during the image
build. Production accepts GET and HEAD only on their paths; videos, projects and
processing stay in the visitor's browser. Health checks require both tool entry
pages as well as the catalog.

The GitHub deploy command resolves each tool's `main` branch to a commit before
building and records those commits in the run summary. Merge tool changes before
running the catalog deploy. A restart reuses the installed image and its bundled
tools; it does not fetch new tool code. To reproduce a manual build, provide a
Compose environment file containing `TOKENS_MOTION_SOURCE` and
`TOKENS_PLAYGROUND_SOURCE`, each set to its HTTPS Git repository URL followed by
`#` and the full commit SHA, and pass it with `docker compose --env-file FILE`.
Local source directories can be used instead when checking an unreleased build.

The `TOKENS_MOTION_URL` and `TOKENS_MOTION_LAB_URL` settings control navigation
links. Their defaults match the bundled paths. `TOKENS_TOOLS_DIR` points to the
built files inside the image; it does not contain visitor data.

## Manage from GitHub Actions

Open **Actions → Manage Token Catalog → Run workflow**, select `main`, and choose a
command:

| Command | What it does |
| --- | --- |
| `status` | Show containers, the next scheduled update and saved run reports. |
| `logs` | Show recent website, tunnel and update logs. |
| `deploy` | Update `/opt/tokens` to the selected main commit, build catalog and bundled tool images, build publication assets and update changed containers. |
| `restart` | Restart the website and tunnel without rebuilding or crawling. |
| `crawl` | Start the existing update service, including alerts and restart after publication. |
| `inspect` | Show a saved run report. Requires a run ID. |
| `approve` | Publish a reviewed capture and restart the website. Requires a run ID. |
| `retry` | Retry a saved capture and restart the website if it publishes. Requires a run ID. |
| `test-notification` | Send a test email or webhook using the host's settings. |
| `usage-report` | Generate a seven-day usage report, show it in the Actions summary and email it to the existing recipient. Leaves the weekly digest checkpoint unchanged. |

Before running **deploy**, open the repository's **Settings → Environments →
production → Environment variables** and add:

```dotenv
TOKENS_PUBLIC_ORIGIN=https://tokencatalog.dev/md3
```

Use a variable, not a secret: the website URL is public. The deploy action saves
this value in `/etc/tokens/server.env` before restarting the website. Changing
it on GitHub takes effect on the next deploy. It also remains available when
systemd starts the website later. Restart and crawl use the saved value.

Deploy does not crawl. It refuses local edits, local commits or a move to an
older commit. Tunnel and email settings stay in `/etc/tokens`.

Deploy keeps the existing build cache before building. Afterwards, it removes
unused cache above roughly 20 GB. This cleanup affects the host's default
Docker builder; it does not remove images, containers, volumes or token data.
Active build cache can exceed that target while a build is running.
For manual deployments, run the same cleanup when needed:

```sh
sudo docker builder prune --all --force --keep-storage 20GB
```

When the service is already active, deploy uses Compose to replace changed
containers and waits for the website to be healthy. The tunnel can remain
running when its configuration has not changed. Startup health checks run
every two seconds; normal checks still run every thirty seconds.
The Actions summary reports time spent fetching source, installing deployment
configuration, building images, generating publication assets and waiting for containers.

A capture that needs review marks the workflow as failed. Read the run report,
then use `inspect` and `approve` when appropriate. Failed captures must not be
approved. Weekly updates still use `tokens-update.timer`, not GitHub's scheduler.

A complete capture with no review issues is published even when it finds zero
token changes. This refreshes the current snapshot and capture time without
changing the detection dates of earlier token changes. Older runs marked
`unchanged` can be approved if their candidate is intact and the live publication
has not changed since they were built.

Workflows run one at a time. Commands that change the site pause the timer and
refuse to run while an updater is active. The cleanup step restores the timer
if it was running before. If the runner loses power or is forcibly stopped,
check `systemctl list-timers --all tokens-update.timer` afterwards.

### Add the runner

1. Open this repository's **Settings → Actions → Runners → New self-hosted
   runner** and select **Linux / x64**.
2. Download and extract the runner into `~/actions-runner-tokens`.
3. Run GitHub's `./config.sh` command as your normal user. Add
   `--name tokens-prod --labels tokens-prod` to it.
4. Install and start its service from that directory:

```sh
sudo ./svc.sh install "$(id -un)"
sudo ./svc.sh start
sudo ./svc.sh status
```

The runner's user must be able to update `/opt/tokens` and fetch the repository.
Run the service as the checkout's owner. Create a `production` environment in
the repository settings and restrict deployment branches to `main`. Add required
reviewers if your GitHub plan supports them.

### Allow the management commands

Run `sudo visudo -f /etc/sudoers.d/tokens-actions`. Add this rule, replacing
`runner` with the runner's Linux username:

```sudoers
runner ALL=(root) NOPASSWD: /usr/bin/docker, /usr/bin/sh /opt/tokens/deploy/install.sh, /usr/bin/sh /opt/tokens/deploy/update.sh *, /usr/bin/systemctl stop tokens-update.timer, /usr/bin/systemctl start tokens-update.timer, /usr/bin/systemctl start tokens-update.service, /usr/bin/systemctl restart tokens.service, /usr/bin/journalctl -u tokens-update.service -u tokens-update-failure.service -n 100 --no-pager
```

This grants root-level access to the host. Restrict workflow and deployment
script changes to trusted maintainers. Do not run pull-request code on this
runner.

Run `status` to check the setup.

## Configure agent discovery

After deploying, follow [agent discovery and Cloudflare setup](cloudflare.md)
to publish DNS-AID records, enable DNSSEC, and check bot, caching and content-signal
settings. The application serves catalogs, Markdown and a public MCP endpoint;
Cloudflare DNS and account settings require separate configuration.

## Tools links

The Tools page and shared navigation link to Material Motion Lab and Carousel
playground. Both applications are deployed separately from the catalog. Set their URLs in `/etc/tokens/server.env`:

```dotenv
TOKENS_MOTION_URL=/tools/motion/
TOKENS_MOTION_LAB_URL=/tools/motion-lab/
```

These are the defaults. Absolute HTTPS URLs also work during a host migration.
Root-relative paths use the catalog hostname, outside its `/md3` collection
path. Configure the corresponding Cloudflare Tunnel path routes before making
the links public. The catalog does not host or proxy the tool applications.
Restart the server after changing the URLs.

`configure-public-origin` replaces `server.env` with the public-origin setting.
Apply custom tool URLs after running it. The catalog deployment does not provide
visitor isolation or browser-side video processing for Material Motion Lab; its Go
server processes videos and stores projects on the host.

## Check a problem

```sh
sudo docker compose -f deploy/compose.yaml ps -a
sudo docker compose -f deploy/compose.yaml logs --tail=80 tokens cloudflared
sudo journalctl -u tokens-update.service -u tokens-update-failure.service -n 100 --no-pager
```

- Website missing: approve the first capture, then start `tokens.service`.
- Tunnel errors: check the token and the `tokens:8080` service URL.
- Exit 137: the process was killed. Check kernel logs for an out-of-memory
  event. The updater's default memory limit is 4 GiB.
- Chromium startup errors: run `sudo sh deploy/test-containers.sh` to test
  the container setup without contacting Material. Do not disable its sandbox.
- No next scheduled run: enable `tokens-update.timer`.
- No email: run the notification test and check the SMTP settings.

See [updates and recovery](updates.md) for review, retries and backups.

## Usage reports and weekly email

The production server logs requests as JSON and retains hourly totals in
`/var/lib/tokens-usage/usage.sqlite`. The writable usage directory is separate
from the read-only publication. Deploys and Docker log rotation do not erase
these totals. Tracking starts when this version is deployed; it does not recover
older traffic.

Totals expire after 90 days. The SQLite database has a 128 MiB page limit, with
additional small WAL and shared-memory files. Deleted pages are reused. Raw
container logs keep their existing rotation policy. Records use fixed route
patterns and known subject names, so arbitrary URLs cannot create unlimited
categories. Logging runs through a bounded queue and flushes every second;
graceful server shutdown drains it. A forced shutdown can lose the last batch.
Dropped records are reported when storage is available; storage failures appear
in the container logs.

Reports include request kinds, response formats, popular subjects, downloads by
scope, MCP methods and tool outcomes, HTTP errors, response bytes, and mean and
maximum response times. Prefetches are separate. Health checks and successful
static-file requests are excluded.
Browser and Cloudflare cache hits may not reach the server. IP addresses,
cookies, query text, MCP arguments and request bodies are not recorded.

Read a report from `/opt/tokens`:

```sh
sudo sh deploy/usage.sh --since 7d
sudo sh deploy/usage.sh --since 30d
```

The installer installs the weekly digest units. Enable them after deploying:

```sh
sudo systemctl enable --now tokens-usage.timer
sudo systemctl list-timers --all tokens-usage.timer
```

The digest runs Mondays at 09:00 Europe/Stockholm and uses the SMTP settings in
`/etc/tokens/update.env`, including the existing sender and recipient. Weekly
periods end at the last complete UTC hour. Events use request completion times;
the scheduled service waits two seconds for the logging queue to flush. The first report covers the previous
seven days and states when tracking started. Later reports resume from the last
successfully emailed period, including delayed deliveries. Persistent timers
catch missed runs; delivery failures retry hourly. A database lease prevents
overlapping weekly sends. SMTP acceptance is followed by saving the checkpoint;
a crash between those operations can still cause a duplicate email.

Send a current report manually without advancing the weekly checkpoint:

```sh
sudo sh deploy/usage.sh --since 7d --send
```

Inspect delivery failures:

```sh
sudo journalctl -u tokens-usage.service -n 100 --no-pager
sudo docker compose -f deploy/compose.yaml logs --tail=100 tokens
```

SMTP settings are required to enable the email timer. It does not use the
optional update webhook. Changes to the timer schedule can be made with
`sudo systemctl edit tokens-usage.timer`, as with the update timer.

### README usage graphic

**Actions → Update usage graphic → Run workflow** refreshes
`docs/images/usage.svg` from the last seven days of retained totals. The same
workflow runs Mondays at 09:15 UTC. The README embeds the image directly.
Only the image is committed; the workflow does not rewrite README text, send
email, or advance the digest checkpoint.

The workflow uses the existing production runner and the deployed `tokens:local`
image. Deploy the SVG exporter before running it for the first time. It has
`contents: write` permission to push the generated image to `main`; repository
rules must allow that bot commit. No additional token or chart service is needed.
The shared production concurrency group keeps it separate from management runs.

To export a graphic manually:

```sh
sudo sh deploy/usage.sh --since 7d --svg > usage.svg
```

Only aggregate counts appear in the public graphic. Prefetches are listed
separately, and successful MCP tool calls exclude initialization and tool listing.

## Generated publication assets

Approved generations remain immutable: their review digests cover every file.
The `assets` Compose service reads `current`, builds all six DTCG/CSS contexts
and the site theme into a new `.site-assets-*` directory beside it, then swaps
`site-assets` to the completed output. No generated data enters Git.

The asset-builder image contains pinned npm dependencies, the compiled CLI and the
same source revision as the server. It runs without network access. The build
checks required theme values, identical component structure across appearances,
the SQLite checksum, and generated bindings against the CSS embedded in the
server. A source-token change requiring different bindings stops the build;
regenerate the checked-in bindings locally, test and release matching images.
A failed build leaves the previous asset link in place.

`tokens.service` builds assets before starting the server. Weekly updates build
assets after publication and before restarting the website. GitHub deploy also
builds assets before replacing containers; approve/retry run the same startup
step. A manual data update requires these commands before restarting a container:

```sh
sudo sh deploy/update.sh build-assets
sudo docker compose -f deploy/compose.yaml restart tokens
```

The server requires `/var/lib/tokens/site-assets`, validates its manifest against
the loaded publication and embedded bindings, and pins the resolved directory.
Missing, stale or incompatible assets stop startup instead of silently disabling
downloads. A build failure after data activation leaves the existing server
running its previously loaded data; fix the build, rebuild assets, then restart.
The reviewed `current` link is not rolled back automatically.

After startup, check the public page and all three formats for a context:

```sh
curl --fail https://tokencatalog.dev/md3/downloads/dtcg/light-default/tokens.tokens.json -o /tmp/tokens.tokens.json
curl --fail https://tokencatalog.dev/md3/downloads/dtcg/light-default/tokens.css -o /tmp/tokens.css
curl --fail https://tokencatalog.dev/md3/downloads/dtcg/light-default/report.json -o /tmp/tokens-report.json
```

Old `.site-assets-*` directories are retained because a running server may still
serve their downloads. After a successful restart and checks, remove older
asset directories manually, keeping the directory targeted by `site-assets`
and one previous successful build for recovery. Do not remove approved source
generations as part of asset cleanup.

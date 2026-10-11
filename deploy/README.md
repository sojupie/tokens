# Deploy Tokens

Use a Linux host with systemd, Git, Docker Engine and Docker Compose 2.24.4 or
newer. The updater can use up to 4 GiB of RAM. Go and Chromium run inside the
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
TOKENS_PUBLIC_ORIGIN=https://tokens.example.com
```

Replace the example with your hostname. Use HTTPS, without a path or trailing
slash. This controls canonical URLs and the sitemap. An empty setting disables
search indexing.

Create a Cloudflare Tunnel for this site. Add a public hostname with:

- Hostname: the hostname from the public URL.
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

This updates the application and Chromium. Weekly data captures do not update
the images. Settings in `/etc/tokens` survive installation and image rebuilds.
Before restarting, make sure `server.env` contains the public URL. Older
installations may not have had this file.

The Docker build keeps Go's module and compiled-package caches between builds.
Chromium is installed in a layer that application changes do not invalidate.
The first build after changing the Dockerfile or base images can still be slow.

## Manage from GitHub Actions

Open **Actions → Manage Tokens → Run workflow**, select `main`, and choose a
command:

| Command | What it does |
| --- | --- |
| `status` | Show containers, the next scheduled update and saved run reports. |
| `logs` | Show recent website, tunnel and update logs. |
| `deploy` | Update `/opt/tokens` to the selected main commit, build both images and update changed containers. |
| `restart` | Restart the website and tunnel without rebuilding or crawling. |
| `crawl` | Start the existing update service, including alerts and restart after publication. |
| `inspect` | Show a saved run report. Requires a run ID. |
| `approve` | Publish a reviewed capture and restart the website. Requires a run ID. |
| `retry` | Retry a saved capture and restart the website if it publishes. Requires a run ID. |
| `test-notification` | Send a test email or webhook using the host's settings. |

Before running **deploy**, open the repository's **Settings → Environments →
production → Environment variables** and add:

```dotenv
TOKENS_PUBLIC_ORIGIN=https://tokens.marmenlind.com
```

Use a variable, not a secret: the website URL is public. The deploy action saves
this value in `/etc/tokens/server.env` before restarting the website. Changing
it on GitHub takes effect on the next deploy. It also remains available when
systemd starts the website later. Restart and crawl use the saved value.

Deploy does not crawl. It refuses local edits, local commits or a move to an
older commit. Tunnel and email settings stay in `/etc/tokens`.

Deploy removes unused Docker build cache before and after building, retaining
up to roughly 4 GB of unused cache. This cleanup affects the host's default
Docker builder; it does not remove images, containers, volumes or token data.
Active build cache can exceed that target while a build is running.
For manual deployments, run the same cleanup when needed:

```sh
sudo docker builder prune --all --force --keep-storage 4GB
```

When the service is already active, deploy uses Compose to replace changed
containers and waits for the website to be healthy. The tunnel can remain
running when its configuration has not changed. Startup health checks run
every two seconds; normal checks still run every thirty seconds.
The Actions summary reports time spent fetching source, installing deployment
configuration, building images and waiting for containers.

A capture that needs review marks the workflow as failed. Read the run report,
then use `inspect` and `approve` when appropriate. Failed captures must not be
approved. Weekly updates still use `tokens-update.timer`, not GitHub's scheduler.

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

After deploying, follow [agent discovery and Cloudflare setup](../docs/agent-discovery.md)
to publish DNS-AID records, enable DNSSEC, and check bot, caching and content-signal
settings. The application serves catalogs, Markdown and a public MCP endpoint;
Cloudflare DNS and account settings require separate configuration.

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

See [updates and recovery](../docs/updates.md) for review, retries and backups.

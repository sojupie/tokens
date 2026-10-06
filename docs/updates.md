# Scheduled updates

Each update saves its receipt, snapshot, candidate publication and report under
`<state-dir>/runs/<run-id>/`. Reports record the failing phase, capture coverage,
change counts and reasons for review. Only accepted runs enter future history.
Failed and unchanged runs remain available for inspection.

## Review and retry

On the Debian host, define this helper to run administrator commands in the
updater container:

```sh
tokens-admin() {
    sudo docker compose -f /opt/tokens/deploy/compose.yaml run --rm --no-deps updater "$@"
}
```

```sh
tokens-admin update run --state-dir /var/lib/tokens --page-catalog /etc/tokens/page-catalog.json
tokens-admin update status --state-dir /var/lib/tokens
tokens-admin update inspect RUN_ID --state-dir /var/lib/tokens
tokens-admin update approve RUN_ID --state-dir /var/lib/tokens
tokens-admin update retry RUN_ID --state-dir /var/lib/tokens
```

The first complete capture always needs approval. Later runs publish automatically
unless they remove a component, source page, token or contextual value; increase
unresolved values; change the reviewed page catalog; or exceed the change threshold. The default threshold is 25%
of the previous contextual value count. Set `--max-change-fraction` when starting
a run to change it. This is a conservative review policy, not a claim that large
changes are wrong.

Incomplete captures, extraction failures and empty publications cannot be
approved. Catalog pages that were never attempted also block approval. Retry
fixes failed pages and retains successful observations from that
exact receipt. It creates a new run; it does not modify the old candidate.
If capture failed before saving a receipt, retry repeats page selection. A
complete capture with a reviewable change needs approval or a fresh run, not a
failed-page retry. Selective partial publication is not implemented.

A 404 does not automatically retire a page. Confirm the upstream removal, update
the reviewed page catalog, then create and review a new run. Fetch failures alone
never authorize deletions.

Approval checks every candidate and snapshot file, verifies the snapshot, and
requires the live generation to match the one used for comparison. An older
pending candidate cannot overwrite a newer live publication. Inspect
`publication.json` at the report's `candidatePath` for actual changes; its
`changes` entries include token names, contexts and before/after values. Filter
on `snapshotId` equal to the run ID to isolate that observation.

To reuse existing capture evidence without another crawl:

```sh
tokens-admin update run --receipt /var/lib/tokens/import/crawl-receipt.json --state-dir /var/lib/tokens
```

On the first run, use `--history-dir /var/lib/tokens/retained-snapshots` to preserve an
existing timeline. Later runs inherit that directory. Keep those snapshots in
place and include them in backups. Do not point it at pending update runs.
Keep retained history inside the state mount so the updater can read it.

## Publication switching

The updater builds a new generation without changing the live directory. On
Linux it replaces the `current` symlink with one rename, so there is no missing
publication between versions. Existing generations remain on disk.

The server resolves the symlink once at startup. Its database and downloads
therefore stay on the same generation. Restart it after manual approval:

```sh
sudo docker compose -f /opt/tokens/deploy/compose.yaml restart tokens
```

An interrupted run can leave a `running` report. The operating system releases
the update lock when the process exits; do not delete the lock file. If activation
finished before the report write, approve that exact run again to finish recovery.
The next update refuses to proceed until that activation is recorded.

Keep `/var/lib/tokens/runs` private: receipts and reports are administrator data,
not public downloads. Back up it and `current` together. The cache is disposable.
Do not remove a generation while a server may still be using it. Automatic
retention and garbage collection are not included.

## Notifications

Set `TOKENS_NOTIFY_WEBHOOK` to an endpoint that accepts an HTTP JSON POST. The body
contains `runId`, `status`, `phase`, `error`, `issues`, `summary`, `liveGeneration`
and `reportPath`. Credentials in the endpoint URL stay out of saved reports.
Use HTTPS unless the receiver is on a trusted local network.

The updater sends failure and review notifications, not a message for every
successful weekly run. A delivery failure is saved as `notificationError` and
makes the command fail. Resend with:

```sh
tokens-admin update notify RUN_ID --state-dir /var/lib/tokens
```

This is a generic webhook, not a Slack, Discord or ntfy-specific format. Use a
receiver or adapter that accepts this body. If no endpoint is set, reports and
the service journal are the only notifications.

The systemd failure unit covers timeouts, killed processes and service failures
outside the normal report path. It can repeat an alert already sent by the
updater; deduplicate by run ID and status at the receiver. If the receiver or
host is down, no local process can guarantee delivery. Monitor the timer and
notification service separately if missed weekly runs must trigger an alert.

## Debian setup

Use Docker Engine with the Compose plugin (2.24.4 or newer). No host Go or Chromium
installation is needed. Place this checkout at `/opt/tokens`.

If replacing the earlier host deployment, stop `tokens-update.timer` and wait
for any active update to finish. Stop `tokens.service` before installing the
new units. Back up `/var/lib/tokens`, then give the container user (UID/GID
10001) ownership of that directory and its existing contents. Do not move the
state: saved reports and symlinks contain absolute paths.

If the earlier `tokens-tunnel` Compose project is running, stop only that
connector before starting the new one:

```sh
sudo docker stop tokens-tunnel-cloudflared-1
```

Leave all containers and units belonging to other projects alone.

Install the configuration and build both images:

```sh
cd /opt/tokens
sudo sh deploy/install.sh
sudoedit /etc/tokens/cloudflared.env
sudo docker compose -f deploy/compose.yaml --profile update build --pull
```

Set `TUNNEL_TOKEN` to the token of a **new, dedicated Tokens tunnel**.
The installer preserves existing environment files. For notifications, edit
`/etc/tokens/update.env` and set `TOKENS_NOTIFY_WEBHOOK`. Leave it empty if no
receiver is configured. Real environment files stay outside the checkout and
must be root-owned with mode `0600`.

In that tunnel, set the hostname to `tokens.marmenlind.com`, service type to
**HTTP**, and service URL to **`tokens:8080`**. The full origin URL is
`http://tokens:8080`, not localhost. Compose provides the service name on the
private web network. No nginx, host networking or host port publishing is needed.
See [Compose networking](https://docs.docker.com/compose/how-tos/networking/).

### Check Chromium and approve the baseline

Before crawling, test sandboxed Chromium without requesting an upstream page:

```sh
sudo docker compose -f deploy/compose.yaml run --rm --no-deps --entrypoint /usr/bin/chromium updater --headless=new --disable-gpu --disable-dev-shm-usage --dump-dom about:blank
```

This must print an HTML document and exit successfully. The updater runs
non-root and uses a syscall profile that permits Chromium's namespace sandbox.
If it fails, check the host's user-namespace and security policy before enabling
the timer. Do not disable the sandbox. See [container details](../deploy/README.md).

Define `tokens-admin` as shown above, then create and inspect the first run:

```sh
sudo systemctl start tokens-update.service
tokens-admin update status --state-dir /var/lib/tokens
tokens-admin update inspect RUN_ID --state-dir /var/lib/tokens
tokens-admin update approve RUN_ID --state-dir /var/lib/tokens
sudo systemctl enable --now tokens.service tokens-update.timer
systemctl list-timers tokens-update.timer
```

The first update service reports failure because baseline review is required.
This does not publish the candidate. Inspect it before approving it. The web
container cannot start until an approved publication exists.

To include older snapshots in the first run, put them under
`/var/lib/tokens/retained-snapshots` and add
`--history-dir /var/lib/tokens/retained-snapshots` to a manual updater run.
Paths passed to container commands must exist inside the state mount; arbitrary
host paths are not visible. All later runs inherit the retained history path.

### Weekly operation

The timer runs on Sundays at 04:00 Stockholm time, with a random delay of up to
15 minutes. `Persistent=true` runs one missed update after the timer starts
again, not every missed week. See [systemd timers](https://manpages.debian.org/trixie/systemd/systemd.timer.5.en.html).

Each run starts a disposable updater container. A successful run restarts only
the web container to load its accepted generation; this also happens for
unchanged data. Failed or held runs leave the web container and tunnel running.
Manual approval needs the web-container restart shown above.

The application has a one-hour deadline. Systemd enforces a 75-minute limit and
stops/removes the named weekly container on exit, including on timeout. The
failure unit runs a separate alert container. Reports and generations persist
in `/var/lib/tokens`; the server mounts that directory read-only. Back it up.

```sh
journalctl -u tokens-update.service -u tokens-update-failure.service
systemctl status tokens-update.timer tokens.service
sudo docker compose -f /opt/tokens/deploy/compose.yaml ps
sudo docker compose -f /opt/tokens/deploy/compose.yaml logs --tail 50 tokens cloudflared
```

The units call the root-owned Docker daemon; container processes run non-root.
Do not grant the updater Docker socket access. Keep the checkout at
`/opt/tokens` writable only by trusted administrators because systemd reads its
Compose configuration as root.

A container health check detects an unhealthy web server but Docker does not
restart a process merely because it becomes unhealthy. Monitor health and the
timer externally if you need alerts when the host or Docker itself is down.

### Updating the deployment

After pulling reviewed code changes:

```sh
cd /opt/tokens
sudo docker compose -f deploy/compose.yaml --profile update build --pull
sudo systemctl restart tokens.service
```

Build the new images while no update is running. Reinstall the units with
`sudo sh deploy/install.sh` if those files changed. Image rebuilds include
Chromium security updates; weekly data updates do not update images.

The tunnel uses its own token and connector. Reusing another tunnel's token
creates a replica of that tunnel, not a separate tunnel. Keep the token private:
Docker administrators can inspect container environment variables.
See [Cloudflare tunnel tokens](https://developers.cloudflare.com/tunnel/reference/tunnel-tokens/).

## History cache

Historical indexes and consecutive comparisons are cached by manifest hash and
processing version. Every reused snapshot still receives byte/hash checks.
Missing or damaged cache entries are rebuilt; cache write failures do not block
publication. Current token ownership is applied after comparison, so cached
changes do not freeze old display grouping.

For manual publication, `tokens publish --history-cache PATH` selects the cache.
The default is `.cache/publication-history` beside the output directory. Updates
use `<state-dir>/cache/history`. The processing version must change when indexing
or comparison rules change.

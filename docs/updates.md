# Scheduled updates

Each update saves its receipt, snapshot, candidate publication and report under
`<state-dir>/runs/<run-id>/`. Reports record the failing phase, capture coverage,
change counts and reasons for review. Only accepted runs enter future history.
Failed and unchanged runs remain available for inspection.

## Review and retry

```sh
tokens update run --state-dir /var/lib/tokens --page-catalog /etc/tokens/page-catalog.json
tokens update status --state-dir /var/lib/tokens
tokens update inspect RUN_ID --state-dir /var/lib/tokens
tokens update approve RUN_ID --state-dir /var/lib/tokens
tokens update retry RUN_ID --state-dir /var/lib/tokens
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
tokens update run --receipt /path/to/crawl-receipt.json --state-dir /var/lib/tokens
```

On the first run, use `--history-dir /path/to/retained-snapshots` to preserve an
existing timeline. Later runs inherit that directory. Keep those snapshots in
place and include them in backups. Do not point it at pending update runs.
Keep retained history outside home directories, which the service hides.

## Publication switching

The updater builds a new generation without changing the live directory. On
Linux it replaces the `current` symlink with one rename, so there is no missing
publication between versions. Existing generations remain on disk.

The server resolves the symlink once at startup. Its database and downloads
therefore stay on the same generation. Restart it after manual approval:

```sh
sudo systemctl restart tokens.service
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
tokens update notify RUN_ID --state-dir /var/lib/tokens
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

Install Go 1.26 or newer and Chromium, then build:

```sh
go build -o bin/tokens ./cmd/tokens
sudo install -m 0755 bin/tokens /usr/local/bin/tokens
sudo useradd --system --home-dir /var/lib/tokens --shell /usr/sbin/nologin tokens
sudo install -d -o tokens -g tokens -m 0700 /var/lib/tokens
sudo install -d -m 0755 /etc/tokens
sudo install -m 0644 inputs/page-catalog.json /etc/tokens/page-catalog.json
sudo install -m 0644 deploy/systemd/tokens*.service deploy/systemd/tokens-update.timer /etc/systemd/system/
sudo systemctl daemon-reload
```

Skip `useradd` if the account already exists. Confirm Chromium is at
`/usr/bin/chromium`; change the update service if it is elsewhere. Adjust the
server's origin and listen address before starting it. Put your reverse proxy in
front of the loopback address.

For notifications, create `/etc/tokens/update.env` using
`deploy/systemd/update.env.example` as the format. Set its real endpoint and make
the file root-owned with mode `0600`. Do not enable the example endpoint.

Run once, inspect the report, and approve the baseline:

```sh
sudo systemctl start tokens-update.service
sudo -u tokens /usr/local/bin/tokens update status --state-dir /var/lib/tokens
sudo -u tokens /usr/local/bin/tokens update inspect RUN_ID --state-dir /var/lib/tokens
sudo -u tokens /usr/local/bin/tokens update approve RUN_ID --state-dir /var/lib/tokens
sudo systemctl enable --now tokens.service
sudo systemctl enable --now tokens-update.timer
systemctl list-timers tokens-update.timer
```

The first `start` reports failure because review is required; the live site is
not changed. A successful scheduled run restarts an already-running token server,
including when data is unchanged. A held or failed run does not restart it.
The application checks its one-hour deadline during capture and at phase
boundaries; systemd enforces a 75-minute process limit.

The timer runs on Sundays at 04:00 Stockholm time, with a 15-minute random delay
window. `Persistent=true` runs a missed update after the timer starts again; it
does not replay every missed week. See Debian's
[timer documentation](https://manpages.debian.org/trixie/systemd/systemd.timer.5.en.html).

```sh
journalctl -u tokens-update.service -u tokens-update-failure.service
systemctl status tokens-update.timer tokens.service
```

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

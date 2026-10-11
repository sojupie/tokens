# Updates and recovery

Run these commands from `/opt/tokens`. For first-time installation, use
[Deploy Token Catalog](deployment.md).

## Run an update

```sh
sudo sh deploy/update.sh run
sudo sh deploy/update.sh status
```

The website keeps serving its current data while an update runs. A normal
weekly update publishes accepted changes and restarts the website. Failed
updates and updates needing review do not replace live data.

A manual update does not restart the website. After a successful manual
publication, run:

```sh
sudo docker compose -f deploy/compose.yaml restart tokens
```

## Review a held update

Copy its run ID from status when prompted:

```sh
printf 'Run ID: '
read -r run_id
sudo sh deploy/update.sh inspect "$run_id"
```

The report lists capture results, change counts and reasons for review.
Check `issues` before approving. Missing pages, extraction errors and empty
data block approval. Large changes and removals need review.

If the changes are expected:

```sh
sudo sh deploy/update.sh approve "$run_id"
sudo docker compose -f deploy/compose.yaml restart tokens
```

If the website has not started yet, enable `tokens.service` instead of using
restart. Detailed token changes are in `publication.json` at the report's
`candidatePath`. An older candidate cannot replace a newer publication.

## Retry a failed capture

After fixing the network or browser problem:

```sh
sudo sh deploy/update.sh retry "$run_id"
```

Retry creates a new run. It keeps successful results and retries failed pages.
Inspect the new run ID, not the old one.

If all payloads were saved but processing failed, copy `receiptPath` from the
report and use it to skip crawling:

```sh
printf 'Receipt path: '
read -r receipt_path
sudo sh deploy/update.sh run --receipt "$receipt_path"
```

The path must be inside `/var/lib/tokens`. Keep the payload cache when reusing
a receipt.

If a source page returns 404, check whether Material actually removed it.
Update `inputs/page-catalog.json` only after confirming the removal, rebuild
the updater image, and start a new run. A failed request is not proof of deletion.

## Notifications

Email settings are in `/etc/tokens/update.env`. Test them with:

```sh
sudo sh deploy/update.sh test-notification
```

Failures and held updates send alerts. Successful and unchanged runs do not.
The systemd failure handler also covers killed processes and timeouts. It can
send a second alert for the same failed run.

Delivery errors appear as `notificationError` in the saved report.
After fixing the settings, resend the notice:

```sh
sudo sh deploy/update.sh notify "$run_id"
```

Email cannot warn about a host that is powered off or has lost all network
access. Use a separate uptime monitor if those failures also need alerts.

## Back up data and settings

Pause the timer and wait for a running update to finish:

```sh
sudo systemctl stop tokens-update.timer
sudo systemctl is-active tokens-update.service
```

When the update service is no longer active:

```sh
sudo install -d -m 0700 /var/backups/tokens
sudo tar -czf "/var/backups/tokens/$(date +%Y%m%d-%H%M%S).tar.gz" /var/lib/tokens /etc/tokens
sudo systemctl start tokens-update.timer
```

Keep backups private: they include email and tunnel credentials.
Restore the directories to the same paths; saved reports contain absolute paths.

Runs and old publications stay in `/var/lib/tokens/runs`; they are not deleted
automatically. Do not delete a publication while the website is using it.
History caches can be rebuilt, but payload caches are needed to reuse receipts.

## Recover an interrupted activation

If the next update asks you to finish an interrupted approval, inspect that
exact run and approve it again. Restart the website afterwards. Do not delete
the lock file; the operating system releases the lock when the process exits.

## Import retained history

Keep imported snapshots under `/var/lib/tokens/imported-history`, with one
snapshot directory per capture. Copy the complete bundle, including its manifest,
source payloads and checksums. Preserve its original capture and review status.

Verify and compare an imported snapshot from the deployment checkout:

```sh
docker compose -f deploy/compose.yaml run --rm --no-deps updater \
  snapshot verify /var/lib/tokens/imported-history/SNAPSHOT_ID

docker compose -f deploy/compose.yaml run --rm --no-deps updater \
  snapshot compare /var/lib/tokens/imported-history/SNAPSHOT_ID \
  /var/lib/tokens/runs/CURRENT_RUN_ID/snapshot
```

Include the directory in the next update:

```sh
sudo sh deploy/update.sh run --history-dir /var/lib/tokens/imported-history
```

To rebuild from saved capture evidence without opening Chromium, add
`--receipt /var/lib/tokens/runs/CURRENT_RUN_ID/receipt.json`. This creates a new
run from the existing evidence. It does not capture new upstream data.

After activation, later updates inherit the imported-history directory from
the active run. Restart the website to load the publication. The publisher
compares consecutive snapshots and uses the oldest as a baseline.

A partial snapshot can provide useful historical values. Missing payloads can
also appear as additions or removals in later comparisons. Inspect capture
coverage before interpreting those entries as upstream changes.

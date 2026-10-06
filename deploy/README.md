# Container deployment

See [setup and scheduled updates](../docs/updates.md).

After building both images, run `sh deploy/test-containers.sh` on Linux. It tests
Chromium and an offline capture fixture, approves its baseline, checks serving
and removes its own test containers and volume. It does not start cloudflared or
read the notification environment file. Do not run two copies at once.

`compose.yaml` defines the web server, its dedicated tunnel connector and a
one-shot updater. The updater has Chromium; the server does not. Both run as UID
10001. Keep `/var/lib/tokens` owned by that UID. The web server mounts it read-only;
only the updater writes it. No container receives the Docker socket.

The updater uses a separate Docker network. Only the web server and cloudflared
share the web network. Neither network joins another project's deployment.
No host ports or browser debugging ports are published.

## Chromium sandbox

`chromium-seccomp.json` is Docker's default profile from
[moby/profiles at 2ceae35d351c156cb5a8efc0fdc4a08cf94569d8](https://github.com/moby/profiles/blob/2ceae35d351c156cb5a8efc0fdc4a08cf94569d8/seccomp/default.json),
with an allow rule for `clone`, `setns`, `unshare` and `chroot`. These calls let
Chromium create its user-namespace sandbox and isolate its filesystem. Allowing
`chroot` here is needed because all container capabilities are dropped; this
does not grant the container the `SYS_CHROOT` capability. The upstream `clone3`
ENOSYS fallback remains.
The upstream Apache 2.0 license is in `MOBY-LICENSE`.

This deliberately widens the updater's syscall policy. It does not grant
`SYS_ADMIN`, disable seccomp or disable Chromium's sandbox. Keep it in sync with
the Docker profile when upgrading Docker. The kernel and host security policy
must also permit unprivileged user namespaces. Test Chromium before enabling the
timer; do not add `--no-sandbox` or privileged mode to silence a startup failure.

The updater has a 2 GiB memory limit, two CPUs and a 256-process limit. Increase
these if measured crawl failures show they are insufficient. Its writable
temporary directory is limited to 1 GiB. The current crawler uses
`--disable-dev-shm-usage`, so browser shared-memory files go there.

## Image updates

Go 1.26 and Debian trixie image tags follow their patch releases. The tunnel
image uses an explicit release tag. Rebuild with `--pull` to get Debian security
updates, including Chromium, then recreate the web container. Review and update
the cloudflared tag separately. The weekly crawl does not rebuild images.

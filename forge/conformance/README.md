# Browser conformance checks

The Go host proxies Catalog and serves checks on the same origin. `pages/` owns
review-page markup, `browser/` owns reusable assertions and the extracted page programs, and ignored `fixtures/` holds
local baseline captures. Existing flat URLs such as `/motion.html` and
`/check-component-rendering.mjs` remain stable through the host's routing.

Run `go run ./forge/conformance --catalog http://127.0.0.1:8085 --listen 127.0.0.1:8082`.
These checks inspect real computed styles and interaction behavior against captured
token values. The pages remain useful for visual investigation; CLI benchmarks
and CI checks belong in `cmd/` rather than new report pages.

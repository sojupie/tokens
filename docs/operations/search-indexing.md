# Search indexing

Set the public collection origin (HTTPS with `/md3`, no trailing slash, query,
or fragment) via `--public-origin` or `TOKENS_PUBLIC_ORIGIN`:

```sh
./bin/tokens serve --published published --listen 127.0.0.1:8081 \
  --public-origin https://tokencatalog.dev/md3
```

The server uses this value for canonical URLs and the sitemap rather than
incoming request headers. When unset, HTML pages emit `noindex`, `/robots.txt`
disallows crawling, and `/sitemap.xml` returns `503`.

## Canonical and `noindex` rules

- **Subject pages & token sets:** The default table and single-table
  `/sets/all` views declare the main subject URL as canonical. Distinct
  token-set pages (`/components/<slug>/sets/...`) keep their own canonicals,
  and legacy `?set=...` URLs canonicalize to the corresponding `/sets/...` URL.
- **Filtered views:** Search queries, context filters, and filtered changelog
  views emit `noindex, follow`.
- **Pagination:** Unfiltered component changelog pages keep their own canonical
  URLs (`?page=2` is not canonicalized to page 1).
- **APIs and downloads:** JSON API and download responses send
  `X-Robots-Tag: noindex`.

## Sitemap and robots.txt

- `/sitemap.xml` lists the home page, component and style indexes, subject
  pages, distinct token sets, each subject's first changelog page, and
  same-origin browser tools. Default-table duplicates and filtered views are
  omitted.
- `/robots.txt` advertises `/sitemap.xml` when `TOKENS_PUBLIC_ORIGIN` is set and
  allows crawlers to fetch filtered pages so they can read their `noindex` tags.

## Post-deploy checks

1. Confirm ordinary crawlers can fetch pages without a Cloudflare Access login
   or bot challenge.
2. Check `<link rel="canonical">` and meta descriptions on a subject page and a
   token-set page.
3. Submit `/sitemap.xml` in Google Search Console and inspect representative
   URLs.

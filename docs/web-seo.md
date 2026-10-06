# Search indexing

Build and run the server with its public origin:

```powershell
go build -o .\bin\sourceloom.exe .\cmd\sourceloom
.\bin\sourceloom.exe serve --published published --listen 127.0.0.1:8081 --public-origin https://YOUR-DOMAIN
```

Replace `https://YOUR-DOMAIN` with the site's real public origin. The origin must not include a path, query, fragment or credentials. Use the public HTTPS origin even when cloudflared connects to a local HTTP listener. The server does not trust request or proxy headers to construct SEO URLs.

Without `--public-origin`, HTML pages have `noindex`, robots.txt disallows crawling, and the sitemap returns 503. This keeps local and preview instances out of search results. Set the flag when deploying the public site.

## Pages and filters

- Component and style pages link to each token set using ordinary HTML links.
- A token-set page such as `/components/buttons/sets/button/color/elevated` renders its tokens on the server. JavaScript is not needed to load the content.
- Existing `?set=...` links still work and declare the corresponding token-set URL as canonical.
- Search, context and changelog-filter views have `noindex, follow`. These views contain selected subsets, so they do not claim to duplicate a full page.
- Unfiltered component changelog pages have their own canonical URLs. Page two keeps `?page=2`; it is not canonicalized to page one.
- API and JSON download responses have an `X-Robots-Tag: noindex` header.

## Sitemap and deployment checks

`/sitemap.xml` lists the home page, component and style indexes, component and style pages, token sets, and each subject's changelog entry page. Further changelog pages are reachable through normal pagination links. Filter combinations are not listed. No fabricated modification dates are included.

`/robots.txt` advertises the sitemap when the public origin is configured. It does not block filtered HTML pages: crawlers must be able to read their `noindex` instructions.

After deployment:

1. Check that the public site is accessible without a Cloudflare Access login or bot challenge for ordinary crawlers.
2. Verify the canonical URL and description in the HTML response for a component and a token-set page.
3. Add the domain to Google Search Console and submit `/sitemap.xml`.
4. Inspect representative URLs in Search Console. Check which canonical Google selected and whether the HTML contains the expected tokens.

Canonical URLs and sitemaps help discovery and consolidation; they do not guarantee indexing or rankings. See Google's [canonical URL guidance](https://developers.google.com/search/docs/crawling-indexing/consolidate-duplicate-urls), [pagination guidance](https://developers.google.com/search/docs/specialty/ecommerce/pagination-and-incremental-page-loading), and [sitemap guidance](https://developers.google.com/search/docs/crawling-indexing/sitemaps/build-sitemap).

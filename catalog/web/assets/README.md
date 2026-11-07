# Browser assets

Edit browser sources in `scripts/` and Catalog CSS in `styles/`. Compiled scripts
belong in `generated/`; run `npm run generate:web` to rebuild them. Headers identify
their source. `vendor/` holds upstream libraries with their license notices.

`static.go` maps embedded files to existing public names. A folder move must not
change a public URL. Images and fonts retain their licenses beside the assets.

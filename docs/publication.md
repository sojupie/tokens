# Offline publication

Source Loom publishes a verified snapshot into a local read-only data set.
The published data is derived from the snapshot. The snapshot remains the
canonical source and is not modified.

## Create a publication

Using the newest verified snapshot:

```text
sourceloom publish
```

Using an explicit snapshot and output directory:

```text
sourceloom publish \
  --snapshot-dir snapshots/material-token-data/material-2026-09-05-1816 \
  --output published
```

The command verifies the snapshot before reading it. The output directory must
be empty or not exist. A failed publication must not be treated as valid
published data.

For local contributor development, the same command can be run with:

```text
go run ./cmd/sourceloom publish
```

Normal documentation and deployment should use the compiled `sourceloom`
binary.

## Generated output

```text
published/
  publication.json
  tokens.sqlite
```

`publication.json` identifies the publication, current snapshot, snapshot
manifest hash, capture status, review status, subjects, source payloads, and
the public token view.

`tokens.sqlite` is a generated read index. It stores subjects, tokens, context-
specific values, lifecycle notices, source payloads, and provenance links.
It can be deleted and rebuilt from the verified snapshot.

The current publication does not replace or copy the immutable snapshot's raw
evidence. Raw graph data and source payloads remain available from the
snapshot directory.

## What is preserved

The publication keeps the distinction between:

- component, style, and foundation subjects;
- source token metadata and resolved values;
- values under different context tags;
- lifecycle notices and ordinary token state;
- source pages, source payloads, and source record IDs.

A token with several context-specific resolved values is stored as one token
with several values. It is not collapsed into whichever value happened to be
read first.

## Current limitations

- The publication currently contains one snapshot.
- Change history across snapshots is not generated yet.
- Normalized download files are not generated yet.
- The web server is not implemented yet.

These are later publication steps. They do not change the immutable snapshot
contract.

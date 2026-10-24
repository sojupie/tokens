# Source

Source captures upstream token data and turns verified snapshots into publications.
The current capture and conversion code supports Material Design 3.

| Directory | Responsibility |
| --- | --- |
| `crawl/` | Chromium capture, discovery, payload cache, and receipts |
| `guidance/` | Versioned rendered documentation, text and Markdown |
| `tokensnapshot/` | Immutable snapshots, verification, and semantic comparisons |
| `publish/` | Published views, SQLite, retained history, and JSON downloads |
| `dtcg/` | Material value conversion, context selection, and DTCG exports |
| `update/` | Saved runs, review, activation, retry, and notifications |

The CLI remains in `cmd/tokens/`. Inputs, schemas, and saved-data paths remain at
the repository root. See [development](../docs/development.md) and
[the snapshot format](../docs/reference/snapshot-format.md).

A second design system needs its own capture and conversion rules. Preserve its
native names and contexts, and map supported values into DTCG without discarding
source evidence. Material-specific payloads and shape interpretation belong here.

[Documentation capture](../docs/reference/guidance-capture.md) preserves prose and
tables independently of the token snapshot pipeline.

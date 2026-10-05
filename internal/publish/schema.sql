PRAGMA foreign_keys = ON;

CREATE TABLE publication (
    schema_version INTEGER NOT NULL,
    publication_id TEXT NOT NULL,
    generated_at TEXT NOT NULL,
    current_snapshot_id TEXT NOT NULL
);

CREATE TABLE snapshots (
    id TEXT PRIMARY KEY,
    created_at TEXT NOT NULL,
    manifest_hash TEXT NOT NULL,
    capture_status TEXT NOT NULL,
    review_status TEXT NOT NULL
);

CREATE TABLE token_changes (
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL,
    subject TEXT NOT NULL,
    source_payload_kind TEXT NOT NULL,
    kind TEXT NOT NULL,
    name TEXT NOT NULL,
    before_json TEXT,
    after_json TEXT,
    previous_snapshot_id TEXT NOT NULL,
    snapshot_id TEXT NOT NULL,
    FOREIGN KEY (previous_snapshot_id) REFERENCES snapshots(id),
    FOREIGN KEY (snapshot_id) REFERENCES snapshots(id)
);

CREATE TABLE subjects (
    kind TEXT NOT NULL,
    slug TEXT NOT NULL,
    name TEXT NOT NULL,
    source_page_urls_json TEXT NOT NULL,
    PRIMARY KEY (kind, slug)
);

CREATE TABLE tokens (
    subject_kind TEXT NOT NULL,
    subject_slug TEXT NOT NULL,
    source_name TEXT NOT NULL,
    display_name TEXT NOT NULL,
    category TEXT NOT NULL,
    raw_json TEXT,
    lifecycle TEXT NOT NULL,
    PRIMARY KEY (subject_kind, subject_slug, source_name),
    FOREIGN KEY (subject_kind, subject_slug)
        REFERENCES subjects(kind, slug)
);

CREATE TABLE token_values (
    subject_kind TEXT NOT NULL,
    subject_slug TEXT NOT NULL,
    source_name TEXT NOT NULL,
    context_tags_json TEXT NOT NULL,
    context_labels_json TEXT NOT NULL,
    is_default_context INTEGER NOT NULL,
    value_json TEXT NOT NULL,
    references_json TEXT NOT NULL,
    raw_json TEXT,
    sources_json TEXT NOT NULL,
    PRIMARY KEY (subject_kind, subject_slug, source_name, context_tags_json),
    FOREIGN KEY (subject_kind, subject_slug, source_name)
        REFERENCES tokens(subject_kind, subject_slug, source_name)
);

CREATE TABLE source_payloads (
    id TEXT PRIMARY KEY,
    url TEXT NOT NULL,
    hash TEXT NOT NULL,
    kind TEXT NOT NULL,
    area TEXT NOT NULL,
    source_subject TEXT NOT NULL,
    page_type TEXT NOT NULL,
    page_urls_json TEXT NOT NULL
);

CREATE TABLE token_sources (
    subject_kind TEXT NOT NULL,
    subject_slug TEXT NOT NULL,
    source_name TEXT NOT NULL,
    payload_id TEXT NOT NULL,
    source_record_id TEXT NOT NULL,
    PRIMARY KEY (subject_kind, subject_slug, source_name, payload_id, source_record_id),
    FOREIGN KEY (subject_kind, subject_slug, source_name)
        REFERENCES tokens(subject_kind, subject_slug, source_name),
    FOREIGN KEY (payload_id)
        REFERENCES source_payloads(id)
);

CREATE TABLE lifecycle_notices (
    id TEXT PRIMARY KEY,
    subject_kind TEXT NOT NULL,
    subject_slug TEXT NOT NULL,
    state TEXT NOT NULL,
    subject TEXT NOT NULL,
    message TEXT NOT NULL,
    source_payload_id TEXT NOT NULL,
    source_record_id TEXT NOT NULL,
    FOREIGN KEY (subject_kind, subject_slug)
        REFERENCES subjects(kind, slug),
    FOREIGN KEY (source_payload_id)
        REFERENCES source_payloads(id)
);

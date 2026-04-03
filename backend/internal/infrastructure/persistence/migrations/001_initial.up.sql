CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email           TEXT NOT NULL UNIQUE,
    name            TEXT NOT NULL,
    avatar_url      TEXT NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_login_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ NULL
);

CREATE TABLE IF NOT EXISTS organizations (
    id                     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name                   TEXT NOT NULL,
    slug                   TEXT NOT NULL UNIQUE,
    ai_api_key_encrypted   TEXT NOT NULL DEFAULT '',
    created_at             TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at             TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at             TIMESTAMPTZ NULL
);

CREATE TABLE IF NOT EXISTS org_memberships (
    org_id      UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role        TEXT NOT NULL CHECK (role IN ('owner','admin','member')),
    joined_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (org_id, user_id)
);

CREATE TABLE IF NOT EXISTS workspaces (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id      UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    slug        TEXT NOT NULL,
    visibility  TEXT NOT NULL DEFAULT 'private' CHECK (visibility IN ('private','org-visible')),
    created_by  UUID NOT NULL REFERENCES users(id),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ NULL,
    UNIQUE (org_id, slug)
);

CREATE TABLE IF NOT EXISTS workspace_memberships (
    workspace_id    UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role            TEXT NOT NULL CHECK (role IN ('admin','editor','viewer')),
    joined_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (workspace_id, user_id)
);

CREATE TABLE IF NOT EXISTS models (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id    UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    description     TEXT NOT NULL DEFAULT '',
    source_format   TEXT NOT NULL DEFAULT 'yaml' CHECK (source_format IN ('dsl','yaml')),
    created_by      UUID NOT NULL REFERENCES users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ NULL
);

CREATE TABLE IF NOT EXISTS model_versions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    model_id        UUID NOT NULL REFERENCES models(id) ON DELETE CASCADE,
    version         INT NOT NULL,
    raw_content     TEXT NOT NULL,
    source_format   TEXT NOT NULL DEFAULT 'yaml' CHECK (source_format IN ('dsl','yaml')),
    commit_message  TEXT NOT NULL DEFAULT '',
    committed_by    UUID NOT NULL REFERENCES users(id),
    committed_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ NULL,
    UNIQUE (model_id, version)
);

CREATE TABLE IF NOT EXISTS changesets (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    model_id        UUID NOT NULL REFERENCES models(id) ON DELETE CASCADE,
    base_version    INT NOT NULL DEFAULT 0,
    actions_json    JSONB NOT NULL DEFAULT '[]',
    status          TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft','committed','rejected')),
    title           TEXT NOT NULL DEFAULT '',
    created_by      UUID NOT NULL REFERENCES users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ NULL
);

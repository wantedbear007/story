-- +goose Up
-- +goose StatementBegin

CREATE TYPE tweet_status AS ENUM ('draft', 'reviewing', 'approved', 'scheduled', 'posted', 'archived');

CREATE TABLE prompt_templates (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(100) NOT NULL,
    version     INT NOT NULL DEFAULT 1,
    template    TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(name, version)
);

CREATE TABLE tweets (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entry_id       UUID NOT NULL REFERENCES entries(id) ON DELETE CASCADE,
    user_id        UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content        TEXT NOT NULL DEFAULT '',
    status         tweet_status NOT NULL DEFAULT 'draft',
    version        INT NOT NULL DEFAULT 1,
    prompt_id      UUID REFERENCES prompt_templates(id),
    provider_name  VARCHAR(50) NOT NULL DEFAULT '',
    model_name     VARCHAR(100) NOT NULL DEFAULT '',
    input_tokens   INT NOT NULL DEFAULT 0,
    output_tokens  INT NOT NULL DEFAULT 0,
    cost_usd       NUMERIC(10,8) NOT NULL DEFAULT 0,
    retry_count    INT NOT NULL DEFAULT 0,
    latency_ms     INT NOT NULL DEFAULT 0,
    error_message  TEXT NOT NULL DEFAULT '',
    scheduled_for  TIMESTAMPTZ,
    posted_at      TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tweets_entry_id ON tweets (entry_id);
CREATE INDEX idx_tweets_user_id ON tweets (user_id);
CREATE INDEX idx_tweets_status ON tweets (user_id, status);
CREATE INDEX idx_tweets_created_at ON tweets (user_id, created_at DESC);

CREATE TABLE generation_audits (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tweet_id          UUID NOT NULL REFERENCES tweets(id) ON DELETE CASCADE,
    action            VARCHAR(50) NOT NULL,
    user_id           UUID REFERENCES users(id),
    previous_content  TEXT NOT NULL DEFAULT '',
    new_content       TEXT NOT NULL DEFAULT '',
    previous_status   tweet_status,
    new_status        tweet_status,
    metadata          JSONB DEFAULT '{}',
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_generation_audits_tweet_id ON generation_audits (tweet_id);
CREATE INDEX idx_generation_audits_created_at ON generation_audits (tweet_id, created_at DESC);

-- Seed default prompt templates
INSERT INTO prompt_templates (id, name, version, template, description) VALUES
    (gen_random_uuid(), 'tweet-summarize', 1,
     'Summarize the following into a tweet under 280 characters:\n\nTitle: {{.Title}}\n\nContent: {{.Content}}',
     'Summarize an entry as a single tweet'),
    (gen_random_uuid(), 'tweet-thread', 1,
     'Convert the following into a Twitter thread. Each tweet must be under 280 characters. Separate tweets with ---:\n\nTitle: {{.Title}}\n\nContent: {{.Content}}',
     'Convert an entry into a multi-tweet thread'),
    (gen_random_uuid(), 'blog-summarize', 1,
     'Summarize the following into a short blog post with a headline and 2-3 paragraphs:\n\nTitle: {{.Title}}\n\nContent: {{.Content}}',
     'Summarize an entry as a short blog post');

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS generation_audits;
DROP TABLE IF EXISTS tweets;
DROP TABLE IF EXISTS prompt_templates;
DROP TYPE IF EXISTS tweet_status;
-- +goose StatementEnd

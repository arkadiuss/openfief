-- The Ask box's guards (docs/DESIGN.md D-208): what the day's model calls cost, for the daily budget, and which
-- sessions and addresses asked the model a question, for the limits per session and per address. Only the modelrouter
-- context reads or writes them. Neither table holds a question, an answer or an amount of the owner's.

-- What the calls of each day cost, per model entry and currency: the tokens the provider reported and what they cost
-- at the entry's price per million tokens. A day is a calendar date in the time zone the core counts "today" in. The
-- budget is the sum of a day's cost, read before each call, so it survives a restart and is the same for every
-- replica of the core.
CREATE TABLE modelrouter.daily_usage (
    day           date    NOT NULL,
    provider      text    NOT NULL CHECK (provider <> ''),
    model         text    NOT NULL CHECK (model <> ''),
    currency      text    NOT NULL CHECK (currency ~ '^[A-Z]{3}$'),
    calls         bigint  NOT NULL DEFAULT 0 CHECK (calls >= 0),
    input_tokens  bigint  NOT NULL DEFAULT 0 CHECK (input_tokens >= 0),
    output_tokens bigint  NOT NULL DEFAULT 0 CHECK (output_tokens >= 0),
    cost          numeric NOT NULL DEFAULT 0 CHECK (cost >= 0),
    PRIMARY KEY (day, provider, model, currency)
);

-- The questions the Ask box sent a model, newest last, kept for the longest window of the limits (a day) and pruned
-- as new ones arrive. address_key is a salted SHA-256 of the asker's address (an IPv4 address, or an IPv6 /64), never
-- the address itself: it can tell that two questions came from one address, and nothing more. It is null when the
-- address was unknown, which leaves the address limit out.
CREATE TABLE modelrouter.ask_questions (
    id          bigserial PRIMARY KEY,
    asked_at    timestamptz NOT NULL,
    session_id  text        NOT NULL CHECK (session_id <> ''),
    address_key text CHECK (address_key IS NULL OR address_key ~ '^[0-9a-f]{64}$')
);
CREATE INDEX ask_questions_by_session ON modelrouter.ask_questions (session_id, asked_at DESC);
CREATE INDEX ask_questions_by_address ON modelrouter.ask_questions (address_key, asked_at DESC) WHERE address_key IS NOT NULL;
CREATE INDEX ask_questions_by_time ON modelrouter.ask_questions (asked_at);

-- The salt of address_key: one random value per database, made here and never changed, so a stored key can't be
-- matched against a list of addresses without it.
CREATE TABLE modelrouter.guard_salt (
    singleton boolean PRIMARY KEY DEFAULT true CHECK (singleton),
    salt      text NOT NULL CHECK (salt <> '')
);
INSERT INTO modelrouter.guard_salt (salt) VALUES (gen_random_uuid()::text || gen_random_uuid()::text);

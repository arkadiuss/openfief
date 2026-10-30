-- D-105: the owner confirms or rejects each suggested transfer between two of their accounts, and the decision is
-- stored in the ledger, append-only like the transaction log (D-18): a correction is the pair's next version, which
-- supersedes the earlier ones. A pair names its withdrawal and its deposit by their transaction keys. The legs are
-- checked by the ledger's use case (ledger.CheckTransfer), not here: a key names every version of a transaction.
CREATE TABLE ledger.transfer_links (
    withdrawal_source      text NOT NULL CHECK (withdrawal_source <> ''),
    withdrawal_external_id text NOT NULL CHECK (withdrawal_external_id <> ''),
    deposit_source         text NOT NULL CHECK (deposit_source <> ''),
    deposit_external_id    text NOT NULL CHECK (deposit_external_id <> ''),
    version                integer NOT NULL CHECK (version >= 1),
    decision               text NOT NULL CHECK (decision IN ('confirmed', 'rejected')),
    recorded_at            timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT transfer_links_pair_version
        PRIMARY KEY (withdrawal_source, withdrawal_external_id, deposit_source, deposit_external_id, version)
);

CREATE TRIGGER transfer_links_append_only
    BEFORE UPDATE OR DELETE ON ledger.transfer_links
    FOR EACH ROW EXECUTE FUNCTION ledger.refuse_change();
CREATE TRIGGER transfer_links_no_truncate
    BEFORE TRUNCATE ON ledger.transfer_links
    FOR EACH STATEMENT EXECUTE FUNCTION ledger.refuse_change();

-- A version must be the next one of its pair: version n needs version n-1.
CREATE FUNCTION ledger.require_previous_link() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF NEW.version > 1 AND NOT EXISTS (
        SELECT 1 FROM ledger.transfer_links
        WHERE withdrawal_source = NEW.withdrawal_source AND withdrawal_external_id = NEW.withdrawal_external_id
          AND deposit_source = NEW.deposit_source AND deposit_external_id = NEW.deposit_external_id
          AND version = NEW.version - 1
    ) THEN
        RAISE EXCEPTION 'ledger.transfer_links: version % of %/% -> %/% has no version %',
            NEW.version, NEW.withdrawal_source, NEW.withdrawal_external_id, NEW.deposit_source,
            NEW.deposit_external_id, NEW.version - 1
            USING ERRCODE = 'OFA02';
    END IF;
    RETURN NEW;
END;
$$;

CREATE TRIGGER transfer_links_in_sequence
    BEFORE INSERT ON ledger.transfer_links
    FOR EACH ROW EXECUTE FUNCTION ledger.require_previous_link();

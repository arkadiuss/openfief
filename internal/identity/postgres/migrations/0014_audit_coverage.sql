-- Audit coverage of every other change of state (docs/DESIGN.md D-211): the owner's changes in the other contexts, each
-- an entry naming what changed by the records' IDs and never an amount, a name or a secret, and the operator's installs,
-- upgrades and removals of plugins, which the core records when it sees a Plugin's generation change. The action column
-- is a closed set, so adding a value is a migration; this replaces the CHECKs, which is not a row change, so the
-- append-only trigger and the withheld privileges are untouched. The list keeps everything 0004 to 0013 added,
-- enrollment_link_cancelled and model_settings_changed included.
ALTER TABLE identity.audit_log DROP CONSTRAINT audit_log_action_check;
ALTER TABLE identity.audit_log ADD CONSTRAINT audit_log_action_check CHECK (action IN (
    'sign_in', 'sign_out', 'sign_out_everywhere', 'passkey_added', 'passkey_removed',
    'enrollment_link_issued', 'enrollment_link_used', 'script_token_created', 'script_token_revoked',
    'grant_created', 'grant_revoked', 'model_call', 'model_call_result',
    'person_created', 'person_revoked', 'maintainer_run', 'maintainer_proposal_dropped', 'maintainer_proposal_applied', 'maintainer_proposal_dismissed',
    'enrollment_link_cancelled', 'model_settings_changed',
    'settings_changed', 'account_created', 'account_changed', 'account_bound', 'account_deleted', 'broker_account_remapped',
    'sleeve_added', 'sleeve_changed', 'sleeve_deleted', 'sleeve_kind_added', 'sleeve_kind_renamed', 'sleeve_kind_deleted',
    'transfer_confirmed', 'transfer_rejected', 'cash_checkpoint_recorded', 'cash_difference_booked', 'instrument_renamed',
    'statement_uploaded', 'ticker_confirmed', 'ticker_corrected', 'bond_lot_added', 'bond_lot_deleted', 'sync_retried',
    'attention_acknowledged', 'attention_resolved', 'price_check_acknowledged',
    'plugin_installed', 'plugin_upgraded', 'plugin_removed'));

-- The operator is an actor of its own, with the ID 'operator', as the core is 'core'.
ALTER TABLE identity.audit_log DROP CONSTRAINT audit_log_actor_kind_check;
ALTER TABLE identity.audit_log ADD CONSTRAINT audit_log_actor_kind_check CHECK (actor_kind IN (
    'owner', 'viewer', 'enrollment_code', 'system', 'operator', 'anonymous'));
ALTER TABLE identity.audit_log DROP CONSTRAINT audit_log_actor_id_check;
ALTER TABLE identity.audit_log ADD CONSTRAINT audit_log_actor_id_check CHECK (actor_id ~ '^([0-9a-f]{16}|core|operator)?$');
ALTER TABLE identity.audit_log ADD CONSTRAINT audit_log_operator_actor CHECK ((actor_kind = 'operator') = (actor_id = 'operator'));

-- The generation of each plugin instance the core last entered in the audit log, so that observing a Plugin's
-- generation again, as the core does every sweep, and after a restart, writes nothing: an entry is made once for an
-- instance and a generation, and the row and the entry commit together. A removed instance's row is deleted with its
-- entry, so a later install of the same name is an install again.
CREATE TABLE identity.plugin_generations (
    instance    text PRIMARY KEY CHECK (instance ~ '^[a-z0-9]([-a-z0-9]{0,61}[a-z0-9])?$'),
    generation  bigint NOT NULL CHECK (generation > 0),
    observed_at timestamptz NOT NULL
);

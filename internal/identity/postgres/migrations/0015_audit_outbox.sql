-- The audit outbox (docs/DESIGN.md D-211, D-198, D-67): a context that changes state enqueues the entry of its change
-- as a River job in its own transaction, and a worker writes it here, once, keyed by the job's ID, with the change's
-- own time. outbox_job_id is that key: NULL for an entry written directly, in the same transaction as the change (the
-- identity context's own), and unique for one the worker wrote, so a worker that runs twice, or a retry after a commit
-- that was not acknowledged, writes nothing the second time. Adding a column and a constraint changes no row, so the
-- append-only trigger and the withheld privileges are untouched.
ALTER TABLE identity.audit_log ADD COLUMN outbox_job_id bigint;
ALTER TABLE identity.audit_log ADD CONSTRAINT audit_log_outbox_job_id_key UNIQUE (outbox_job_id);

-- A sync the core started that changed state, and the administrator's changes at the core's command line. The list
-- keeps everything 0014 allowed.
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
    'plugin_installed', 'plugin_upgraded', 'plugin_removed',
    'connector_synced'));

-- The administrator at the core's command line is an actor of its own, with the ID 'admin', as the core is 'core'.
ALTER TABLE identity.audit_log DROP CONSTRAINT audit_log_actor_kind_check;
ALTER TABLE identity.audit_log ADD CONSTRAINT audit_log_actor_kind_check CHECK (actor_kind IN (
    'owner', 'viewer', 'enrollment_code', 'system', 'operator', 'admin', 'anonymous'));
ALTER TABLE identity.audit_log DROP CONSTRAINT audit_log_actor_id_check;
ALTER TABLE identity.audit_log ADD CONSTRAINT audit_log_actor_id_check CHECK (actor_id ~ '^([0-9a-f]{16}|core|operator|admin)?$');
ALTER TABLE identity.audit_log ADD CONSTRAINT audit_log_admin_actor CHECK ((actor_kind = 'admin') = (actor_id = 'admin'));

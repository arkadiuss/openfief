-- A script token's calls of the MCP server's tools (docs/DESIGN.md D-213, D-211): usage only, each an entry whose actor
-- is the token, by its ID, and that names the tool called. The action and actor columns are closed sets, so adding a
-- value is a migration; this replaces the CHECKs, which is not a row change, so the append-only trigger and the
-- withheld privileges are untouched. The action list keeps everything 0015 allowed, connector_synced included.
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
    'connector_synced', 'mcp_call'));

-- A script token is an actor of its MCP calls, named by its 16-digit ID, which the actor_id pattern already allows.
ALTER TABLE identity.audit_log DROP CONSTRAINT audit_log_actor_kind_check;
ALTER TABLE identity.audit_log ADD CONSTRAINT audit_log_actor_kind_check CHECK (actor_kind IN (
    'owner', 'viewer', 'enrollment_code', 'system', 'operator', 'admin', 'script_token', 'anonymous'));
ALTER TABLE identity.audit_log ADD CONSTRAINT audit_log_script_token_actor CHECK (actor_kind <> 'script_token' OR actor_id ~ '^[0-9a-f]{16}$');

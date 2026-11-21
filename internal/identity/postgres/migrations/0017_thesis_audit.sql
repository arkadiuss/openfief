-- A draft thesis the owner saved or edited (docs/DESIGN.md D-229, D-211): each an entry that names the thesis and its
-- version, never its text. The action column is a closed set, so adding a value is a migration; this replaces the
-- CHECK, which is not a row change, so the append-only trigger and the withheld privileges are untouched. The list
-- keeps everything 0016 allowed.
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
    'connector_synced', 'mcp_call',
    'thesis_saved', 'thesis_edited'));

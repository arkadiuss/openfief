-- The owner's changes to the agents-and-models settings (docs/DESIGN.md D-212, D-211): the switch, a feature's toggle,
-- a role's route and a data tier's models, each an entry that names the setting. The action column is a closed set, so
-- adding a value is a migration; this replaces the CHECK, which is not a row change, so the append-only trigger and the
-- withheld privileges are untouched. The list keeps everything 0004 to 0012 added, enrollment_link_cancelled included.
ALTER TABLE identity.audit_log DROP CONSTRAINT audit_log_action_check;
ALTER TABLE identity.audit_log ADD CONSTRAINT audit_log_action_check CHECK (action IN (
    'sign_in', 'sign_out', 'sign_out_everywhere', 'passkey_added', 'passkey_removed',
    'enrollment_link_issued', 'enrollment_link_used', 'script_token_created', 'script_token_revoked',
    'grant_created', 'grant_revoked', 'model_call', 'model_call_result',
    'person_created', 'person_revoked', 'maintainer_run', 'maintainer_proposal_dropped', 'maintainer_proposal_applied', 'maintainer_proposal_dismissed',
    'enrollment_link_cancelled', 'model_settings_changed'));

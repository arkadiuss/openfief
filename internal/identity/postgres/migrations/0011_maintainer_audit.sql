-- The model-portfolio maintainer's audit entries (docs/DESIGN.md D-204, D-211): a run that was refused or failed with
-- its class, a proposal dropped for breaking a rule, and the owner applying or dismissing one. The action column is a
-- closed set, so adding values is a migration; this replaces the CHECK, which is not a row change, so the append-only
-- trigger and the withheld privileges are untouched. The list keeps everything 0004 to 0010 added.
ALTER TABLE identity.audit_log DROP CONSTRAINT audit_log_action_check;
ALTER TABLE identity.audit_log ADD CONSTRAINT audit_log_action_check CHECK (action IN (
    'sign_in', 'sign_out', 'sign_out_everywhere', 'passkey_added', 'passkey_removed',
    'enrollment_link_issued', 'enrollment_link_used', 'script_token_created', 'script_token_revoked',
    'grant_created', 'grant_revoked', 'model_call', 'model_call_result',
    'person_created', 'person_revoked', 'maintainer_run', 'maintainer_proposal_dropped', 'maintainer_proposal_applied', 'maintainer_proposal_dismissed'));

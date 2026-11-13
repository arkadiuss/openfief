-- A model call is two audit entries linked by a call ID (docs/DESIGN.md D-211): `model_call`, written before the call is
-- sent, with the outcome `sent`, and `model_call_result` after it, with ok, failed or refused. A call the router refuses
-- before it chooses a model has only the result entry. The action and outcome columns are closed sets, so adding a value
-- is a migration (0004); this replaces the two CHECKs, which is not a row change, so the append-only trigger and the
-- withheld privileges are untouched. The action list keeps everything 0006 and 0007 added.
ALTER TABLE identity.audit_log DROP CONSTRAINT audit_log_action_check;
ALTER TABLE identity.audit_log ADD CONSTRAINT audit_log_action_check CHECK (action IN (
    'sign_in', 'sign_out', 'sign_out_everywhere', 'passkey_added', 'passkey_removed',
    'enrollment_link_issued', 'enrollment_link_used', 'script_token_created', 'script_token_revoked',
    'grant_created', 'grant_revoked', 'model_call', 'model_call_result'));
ALTER TABLE identity.audit_log DROP CONSTRAINT audit_log_outcome_check;
ALTER TABLE identity.audit_log ADD CONSTRAINT audit_log_outcome_check CHECK (outcome IN ('ok', 'refused', 'failed', 'sent'));

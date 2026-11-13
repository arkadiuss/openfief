-- Model calls enter the audit log (docs/DESIGN.md D-205, D-211): the action `model_call` names the role and the model
-- entry by name, and the outcome. Adding an action is a migration (0004); the table stays append-only, and this only
-- replaces the action's CHECK, which is not a row change. The list keeps the two actions of a grant (0006, D-209), so that
-- this migration removes no action another one added.
ALTER TABLE identity.audit_log DROP CONSTRAINT audit_log_action_check;
ALTER TABLE identity.audit_log ADD CONSTRAINT audit_log_action_check CHECK (action IN (
    'sign_in', 'sign_out', 'sign_out_everywhere', 'passkey_added', 'passkey_removed',
    'enrollment_link_issued', 'enrollment_link_used', 'script_token_created', 'script_token_revoked',
    'grant_created', 'grant_revoked', 'model_call'));

-- The owner can cancel a viewer's invitation before it is used (D-183, D-210). A cancelled code is refused like an
-- expired one, and the cancellation is an audit entry.

ALTER TABLE identity.enrollment_codes
    ADD COLUMN cancelled_at timestamptz CHECK (cancelled_at >= created_at),
    -- Only an invitation that was not used is cancelled, and only a viewer's is an invitation.
    ADD CONSTRAINT enrollment_codes_cancel_invite CHECK (cancelled_at IS NULL OR (used_at IS NULL AND viewer_name IS NOT NULL));

ALTER TABLE identity.audit_log DROP CONSTRAINT audit_log_action_check;
ALTER TABLE identity.audit_log ADD CONSTRAINT audit_log_action_check CHECK (action IN (
    'sign_in', 'sign_out', 'sign_out_everywhere', 'passkey_added', 'passkey_removed',
    'enrollment_link_issued', 'enrollment_link_used', 'enrollment_link_cancelled', 'script_token_created', 'script_token_revoked',
    'grant_created', 'grant_revoked', 'model_call', 'model_call_result', 'person_created', 'person_revoked', 'maintainer_run', 'maintainer_proposal_dropped', 'maintainer_proposal_applied', 'maintainer_proposal_dismissed'));

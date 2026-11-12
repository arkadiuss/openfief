-- A registration begun by an owner session names that session, so the audit log (docs/DESIGN.md D-211) can say
-- which sign-in created the passkey when the ceremony finishes. Null for a registration begun with an enrollment
-- code, and for the other ceremonies.
ALTER TABLE identity.ceremonies
    ADD COLUMN session_id text REFERENCES identity.sessions (id) CHECK (session_id ~ '^[0-9a-f]{16}$');

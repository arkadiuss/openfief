---
title: Identity Enrollment Code Expiry At Use
status: active
fact_kind: subject
subject_key: identity.enrollment_code_expiry_at_use
origin:
  kind: agent
  ref: 'Security review before the alpha.2 cut, known limit #4 (2026-11-21)'
  recorded_at: '2026-11-21T18:38:30.508Z'
id: FACT-SUBJECT-IDENTITY-ENROLLMENT-CODE-EXPIRY-AT-USE
type: fact
---
Whether an enrollment code, the owner's own link or a viewer's invitation, has expired at the moment a finished registration spends it, checked in the same statement that marks it used. It does not cover the check when the registration begins, the code's lifetime, its single use or its cancellation.
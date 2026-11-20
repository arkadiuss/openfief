---
title: Turnstile Token Verification
status: active
fact_kind: subject
subject_key: turnstile.token_verification
origin:
  kind: agent
  ref: Author's decision to upgrade to kibi 2.45 and review the subjects, 2026-11-20
  recorded_at: '2026-11-20T12:51:08.000Z'
id: FACT-SUBJECT-3AF40C86588D5F97
type: fact
---
The server-side check that a Turnstile token is accepted by Cloudflare's siteverify for the asking host before the core counts a question, and what a refused or impossible check does. It does not cover the browser-side messages.
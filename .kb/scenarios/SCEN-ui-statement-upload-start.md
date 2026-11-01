---
title: The owner chooses how an account starts in the upload form
status: active
origin:
  kind: agent
  recorded_at: '2026-11-01T07:35:38.991Z'
id: SCEN-ui-statement-upload-start
type: scenario
---
Given an account that hasn't started, when the owner opens its page, then the upload form asks how it starts, from the holdings of a current statement with the cash at the start or from a statement of its whole history, and the cash entered reaches the core with the upload in upper-case currency; once the account has started, the form preselects transactions and shows no cash field.
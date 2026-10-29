---
title: The owner manages sleeves and accounts in the built-in UI
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T12:45:42.320Z'
id: SCEN-ui-sleeve-and-account-pages
type: scenario
---
Given an owner session, when the owner opens Sleeves or Accounts, then the pages list kinds, sleeves, accounts with their connectors and broker accounts, mapped and unmapped, with forms to add and change them, the broker account picked from a list; a posted form goes to the core and back to its page, and a refusal shows the core's reason; a read session sees the lists without forms and can't post.
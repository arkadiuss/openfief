---
title: The owner grants a script token or a view plugin an account, a sleeve or the household, each grant and revocation is an audit entry, and the grant is a relationship that only lets its holder view
status: active
origin:
  kind: agent
  recorded_at: '2026-11-13T10:00:17.217Z'
id: SCEN-identity-grants-per-target
type: scenario
---
Given an owner session and a script token and a view plugin that exist, when the owner grants the token an account, a sleeve and the household and the plugin the household, then each grant is stored, its audit entry by the owner names the holder and the target by kind and ID in the same transaction, and SpiceDB holds one relationship for it: a viewer of the account or the sleeve, or a reader of the household, never a relation that feeds change. Granting again changes nothing. Revoking removes the relationship, keeps the grant as history with its end and writes its entry; a holder or target that does not exist is not found. With the audit log down no grant or revocation is made. Assumes D-209 and D-211.
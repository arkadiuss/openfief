---
title: A grant ends, with an audit entry, when its script token is revoked, its view plugin is uninstalled or its sleeve or account is deleted
status: active
origin:
  kind: agent
  recorded_at: '2026-11-13T10:00:18.976Z'
id: SCEN-identity-grants-end-with-their-records
type: scenario
---
Given grants of a script token, of a view plugin, of a sleeve and of an account, when the token is revoked, the plugin instance is uninstalled, the sleeve is deleted (only an empty one is) or the account is deleted, then every grant that named it is ended, kept as history, removed from SpiceDB and written to the audit log, the token's in the transaction of its revocation and the others as the core; a sleeve created later under the same ID inherits nothing; and at its start the core ends the grants of anything that went while it was down. Assumes D-209, D-120 and the author's answers that a plugin's grant is its own and ends with its uninstall.
---
title: The database refuses to change or remove a stored transaction version
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T06:29:12.975Z'
id: SCEN-ledger-pg-append-only
type: scenario
---
Given stored transaction versions and postings, when any role, the schema's owner included, tries to update, delete or truncate them, or to add a posting to a stored version, then the statement fails and the log is unchanged; when a version is stored that is not the next of its key, then it is refused; when a row is changed with the protection switched off, then loading the log fails because the version no longer matches its fingerprint.
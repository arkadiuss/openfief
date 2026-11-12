---
title: SpiceDB failing refuses everyone, the owner too
status: active
origin:
  kind: agent
  recorded_at: '2026-11-12T22:51:42.041Z'
id: SCEN-identity-spicedb-fails-closed
type: scenario
---
Given the core started with --authz=spicedb and a SpiceDB that errors, times out or can't be reached, when any holder, the owner included, asks to view or change an account or to list the accounts it may view, then the answer is a refusal with the error, never the owner's view and never every account; the views fail the request. Assumes the answer is read through identityapp.Authorizer.
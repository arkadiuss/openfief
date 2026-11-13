---
title: A token is created with its grants atomically, and a failure keeps no token and shows no secret
status: active
origin:
  kind: agent
  ref: 'S5 task brief: the Access page, script tokens and view plugins (D-209, D-210), 2026-11-13'
  recorded_at: '2026-11-13T18:00:00.000Z'
id: SCEN-identity-token-created-with-grants
type: scenario
---
Given the owner with a recent passkey check, CreateScriptToken with the MCP flag and grants of sleeves and accounts stores the token, its grants and an entry for each, then writes the grants to the model; a repeated target is granted once; a target that does not exist or is malformed is refused before anything is stored; a model that does not answer refuses a token that is to hold grants and stores nothing; a model whose write fails revokes the token with its own entry and returns no secret; a token with no grants needs nothing of the model; in PostgreSQL an audit log that refuses an entry leaves no token and no grant.
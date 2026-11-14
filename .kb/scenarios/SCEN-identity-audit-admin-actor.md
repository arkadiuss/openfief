---
title: The administrator's changes at the command line are entered as the admin actor
status: active
origin:
  kind: agent
  ref: the S9 brief, from D-211, D-198 and D-67, 2026-11-14
  recorded_at: '2026-11-14T04:22:54.525Z'
id: SCEN-identity-audit-admin-actor
type: scenario
---
Given the core's command line, when the administrator issues or revokes a token, issues an enrollment link or binds a connector instance to an account, then the entry's actor is admin; the core's own work stays the system's.

## Context

D-211: every change of state is an entry of the audit log, failing closed. Asked for in the S9 brief.

## Source

> An append-only audit log records every change of state and every access event: sign-ins and sign-outs, passkeys, enrollment links, script tokens, grants, settings, the owner's entries and corrections, decisions on attention items, plugin installs and upgrades the operator applies, and model and MCP calls. Each entry names who, what, which records by ID, when and the outcome, never an amount, a secret or a question's text (GR-12). Settings shows it to the owner, filterable.

Source: the S9 brief, from D-211, 2026-11-14
---
title: Each owner change in the other contexts is one audit entry, and a change the log can't record is not made
status: active
origin:
  kind: agent
  ref: the S8 brief, from D-211, 2026-11-14
  recorded_at: '2026-11-14T02:39:45.004Z'
id: SCEN-identity-audit-owner-changes
type: scenario
---
Given the owner signed in, when the owner changes a setting, an account, a sleeve or a kind, a transfer, a cash checkpoint, an instrument's name or an attention item, then one ok entry of the owner's session names the action and the records by ID, a failing log leaves the change unmade and answers unavailable, and a change that fails writes no entry.

## Context

D-211: every change of state is an entry of the audit log, failing closed.

## Source

> An append-only audit log records every change of state and every access event: sign-ins and sign-outs, passkeys, enrollment links, script tokens, grants, settings, the owner's entries and corrections, decisions on attention items, plugin installs and upgrades the operator applies, and model and MCP calls. Each entry names who, what, which records by ID, when and the outcome, never an amount, a secret or a question's text (GR-12). Settings shows it to the owner, filterable.

Source: the S8 brief, from D-211, 2026-11-14
---
title: A change and its entry commit together through the outbox, and the log being down loses no entry
status: active
origin:
  kind: agent
  ref: the S9 brief, from D-211, D-198 and D-67, 2026-11-14
  recorded_at: '2026-11-14T04:22:49.242Z'
id: SCEN-identity-audit-outbox
type: scenario
---
Given an owner's change, when it is made, then its entry is a job enqueued in the same transaction and a worker writes it once with the change's time; when the change fails, no entry exists; when the entry can't be enqueued, the change is not made; when the log is down, the change is made and the entry waits and is written when the log is back.

## Context

D-211: every change of state is an entry of the audit log, failing closed. Asked for in the S9 brief.

## Source

> An append-only audit log records every change of state and every access event: sign-ins and sign-outs, passkeys, enrollment links, script tokens, grants, settings, the owner's entries and corrections, decisions on attention items, plugin installs and upgrades the operator applies, and model and MCP calls. Each entry names who, what, which records by ID, when and the outcome, never an amount, a secret or a question's text (GR-12). Settings shows it to the owner, filterable.

Source: the S9 brief, from D-211, 2026-11-14
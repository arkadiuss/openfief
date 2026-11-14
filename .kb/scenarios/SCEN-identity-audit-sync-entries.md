---
title: A sync that changed state is one entry of the system, a sync that changed nothing is none
status: active
origin:
  kind: agent
  ref: the S9 brief, from D-211, D-198 and D-67, 2026-11-14
  recorded_at: '2026-11-14T04:22:52.490Z'
id: SCEN-identity-audit-sync-entries
type: scenario
---
Given a connector sync the core started, when it wrote new transaction versions, prices or bond terms, then one entry of the system names the instance, the batch, the accounts and the counts, in the transaction of its writes; when it brought nothing new, or failed, then no entry exists.

## Context

D-211: every change of state is an entry of the audit log, failing closed. Asked for in the S9 brief.

## Source

> An append-only audit log records every change of state and every access event: sign-ins and sign-outs, passkeys, enrollment links, script tokens, grants, settings, the owner's entries and corrections, decisions on attention items, plugin installs and upgrades the operator applies, and model and MCP calls. Each entry names who, what, which records by ID, when and the outcome, never an amount, a secret or a question's text (GR-12). Settings shows it to the owner, filterable.

Source: the S9 brief, from D-211, 2026-11-14
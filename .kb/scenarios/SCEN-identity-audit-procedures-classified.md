---
title: A procedure of an owner service that changes something and has no entry fails the test
status: active
origin:
  kind: agent
  ref: the S8 brief, from D-211, 2026-11-14
  recorded_at: '2026-11-14T02:39:47.096Z'
id: SCEN-identity-audit-procedures-classified
type: scenario
---
Given the procedures of the owner services, when one is in no list and one of the audited changes leaves no entry, then the test fails, naming it.

## Context

D-211: every change of state is an entry of the audit log, failing closed.

## Source

> An append-only audit log records every change of state and every access event: sign-ins and sign-outs, passkeys, enrollment links, script tokens, grants, settings, the owner's entries and corrections, decisions on attention items, plugin installs and upgrades the operator applies, and model and MCP calls. Each entry names who, what, which records by ID, when and the outcome, never an amount, a secret or a question's text (GR-12). Settings shows it to the owner, filterable.

Source: the S8 brief, from D-211, 2026-11-14
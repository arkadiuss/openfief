---
title: Sweeps, pushes, owner actions and the relay keep one item per subject and publish each change
status: active
sourceFile: internal/attention/app/items_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-10-31T18:30:33.687Z'
id: TEST-attention-app-items
type: test
---
Runs attention/app.Items over an in-memory store that keeps an InTx's writes only when it succeeds and writes an event with every change, as the PostgreSQL store does. Two sweeps of the same findings raise each item once with one event; a cleared finding resolves its item; Raise and Clear are idempotent; the owner acknowledges and resolves only allowed kinds while a read or missing session is refused; a failed append keeps nothing; the relay publishes in order, stops at a failure and resumes there; the log publisher never writes the subject or detail.
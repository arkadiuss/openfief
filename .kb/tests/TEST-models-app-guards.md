---
title: The router refuses before calling past the budget, counts what it costs, gates calls at once, and the Ask door admits questions
status: active
sourceFile: internal/modelrouter/app/guards_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-14T20:37:51.603Z'
id: TEST-models-app-guards
type: test
---
Router tests over fakes: the budget refuses before any provider call with the day's end and an audit entry of class budget, counts across days and in the core's zone, reads and records nothing without a cap, counts a priceless entry as free, refuses a candidate priced in another currency, fails closed on a usage that can't be read and withholds an answer it can't count; the gate refuses a third call as busy after the queue wait, queues for a freed slot, frees on failure and ends a wait with the caller's context; Ask.Admit counts per session and address, says when it resets, keys addresses (IPv4, IPv6 /64, mapped, zone, junk), checks the budget before counting, and allows without counting when no limit is set.
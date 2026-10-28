---
title: Everything the ledger stores in PostgreSQL reads back exactly
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T06:29:08.149Z'
id: SCEN-ledger-pg-round-trip
type: scenario
---
Given a fresh database with the ledger migrations applied, when sleeves, accounts, versions of every transaction kind and assertion checks are stored, then reading them back gives equal values: amounts and quantities exact to every digit, dates, descriptions and the order of versions; running the migrations again, even concurrently, changes nothing.
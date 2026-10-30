---
title: Keys, sleeve IDs and account IDs are made from the owner's names
status: active
sourceFile: internal/portfolio/key_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T11:35:20.854Z'
id: TEST-portfolio-key
type: test
---
Table test of portfolio.KeyFromName: words become lowercase hyphenated keys, Polish letters are spelled in ASCII, punctuation and spaces collapse to one hyphen with none at either end, digits are kept, other scripts are dropped, a long name is cut to 64 characters without a trailing hyphen, every result is a valid KindKey, and an empty or punctuation-only name fails with ErrNoKey.
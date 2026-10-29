---
title: The owner adds and changes sleeve kinds and sleeves; nobody else can
status: active
sourceFile: internal/portfolio/app/sleeves_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T12:45:14.484Z'
id: TEST-portfolio-app-sleeves
type: test
---
Use-case tests of portfolio app Sleeves over an in-memory store that applies the portfolio rules: adding a kind makes its key from the name, Polish letters included, and refuses a taken key, a name without a key and a kind without a role; renaming keeps the key and role and refuses an unknown kind; adding a sleeve makes its ID from the name and takes the kind's role, refuses another role without the override, a taken ID and an unknown kind; updating renames and moves a default sleeve and refuses an unknown one; a refused or read session and a failing store store nothing.
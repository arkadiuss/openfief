---
title: The kernel's LogSafe names known errors only, else the innermost type
status: active
verification_scope: unit
sourceFile: internal/kernel/logsafe_test.go
origin:
  kind: agent
  recorded_at: '2026-11-03T03:24:08.949Z'
id: TEST-kernel-logsafe
type: test
---
Tables errors that wrap synthetic values (an amount, an ISIN, an account number) through kernel.LogSafe with two known errors: known ones in the chain are named, joined; an unknown one, alone or joined, is described only by its innermost type; nil is empty. A description that quoted the error's text fails it.
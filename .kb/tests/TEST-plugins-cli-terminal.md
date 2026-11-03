---
title: Everything the CLI says is written with control characters escaped
status: active
sourceFile: internal/plugins/cli/terminal_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-03T09:18:05.605Z'
id: TEST-plugins-cli-terminal
type: test
---
TestSayEscapesControls writes through say, the CLI's one writer of prompts and screens: plain text with a tab and newline is unchanged, while the review's payload, a C1 control, a right-to-left override and invalid UTF-8 come out as \u escapes, so nothing that reaches say unescaped can steer the terminal.
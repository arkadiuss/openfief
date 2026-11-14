---
title: The bench's report traces each round with what was sent and what the host reported
status: active
sourceFile: tools/askbench/finalround_test.go
verification_scope: integration
origin:
  kind: agent
  ref: 'Live bench finding on R1 (gpt-oss-120b on deepinfra: two model calls, 1694 input tokens, a greeting), 2026-11-14'
  recorded_at: '2026-11-14T23:20:12.489Z'
id: TEST-askbench-rounds
type: test
---
Runs the command against a fake OpenAI-compatible server and checks that each run of R1 has two rounds with their roles, message counts, lengths and hashes (and the question's text nowhere in report.json), the host's token counts, finish reason and reply text, and that a reply with cached tokens, a tool call and over 500 characters of text is traced with the cached count, the tool's name and the text cut to 500.
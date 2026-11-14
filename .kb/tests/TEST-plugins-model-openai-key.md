---
title: The plugin reads its key from the mounted Secret
status: active
origin:
  kind: agent
  ref: 'AB1 brief from the author: the Ask box bench (D-217, D-218, D-219), 2026-11-14'
  recorded_at: '2026-11-14T08:24:02.861Z'
sourceFile: plugins/model-openai/key_test.go
verification_scope: unit
id: TEST-plugins-model-openai-key
type: test
---
Reads the key from a temporary directory: trimmed, absent as no key, empty mount as no key, an empty key as an error that holds no key.
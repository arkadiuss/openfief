---
title: Malicious Flex statement files are refused before they cost memory or leak files
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T06:23:28.686Z'
id: SCEN-ibkrflex-hostile
type: scenario
---
Given statement files with a DTD defining nested entities, an external entity naming a local file, an undeclared entity, a non-UTF-8 encoding, more bytes, depth, elements or attributes than the configured limits, or an attribute longer than allowed, when they are mapped, then mapping fails with a malformed or too-large error and no data; given an amount written as 1e999999999, then that row is an unsupported problem rather than a huge number.
---
title: A number format is a pair of separators the owner picks from a fixed list
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T21:47:53.365Z'
id: SCEN-jurisdiction-number-format-choices
type: scenario
---
Given the formats space-comma, space-dot, comma-dot, dot-comma, apostrophe-dot, none-comma and none-dot, when one is stored by its key and read back, then it is the same pair; a format with one separator for both jobs or no decimal separator is refused.
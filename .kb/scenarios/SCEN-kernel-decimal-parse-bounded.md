---
title: A decimal with a huge exponent or too much text is refused at once
status: active
origin:
  kind: agent
  recorded_at: '2026-11-02T20:57:31.154Z'
id: SCEN-kernel-decimal-parse-bounded
type: scenario
---
Given text such as 1E+2000000000, an exponent past the range of an int, or more than 256 bytes of digits, when it is parsed as a decimal, then ParseDecimal refuses it with ErrInvalidDecimal at once, without building the number, and the error quotes none of the text; a spreadsheet's 1.5E-7, 1E+100 and 1E-100 and 256 bytes of digits still parse exactly.
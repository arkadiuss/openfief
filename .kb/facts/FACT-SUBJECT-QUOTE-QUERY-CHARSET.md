---
title: Quote Query Charset
status: active
fact_kind: subject
subject_key: quote.query_charset
origin:
  kind: agent
  ref: 'Security review before the alpha.2 cut, known limit #5 (2026-11-21)'
  recorded_at: '2026-11-21T18:46:03.061Z'
id: FACT-SUBJECT-QUOTE-QUERY-CHARSET
type: fact
---
The characters a market quote's query, the ticker or ISIN someone asks for, may hold where the core, the Ask box's market price tool and the EODHD connector check it before it goes into a provider's URL. It does not cover a listing's symbol a connector answers with, the period of the quote or its length limit.
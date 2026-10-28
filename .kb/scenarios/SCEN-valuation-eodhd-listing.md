---
title: An ISIN resolves to one EODHD listing on a supported exchange, or a clear error
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T21:37:08.080Z'
id: SCEN-valuation-eodhd-listing
type: scenario
---
Given EODHD's search results for synthetic ISIN XX00000SYNT3 on Xetra, the LSE and Frankfurt, when no exchange is named, then picking fails as ambiguous naming SYNT.XETRA and SYNT.LSE; when LSE is named, then the LSE listing is picked; given an ISIN listed only on Frankfurt or a US exchange, then picking fails as not listed, because only Warsaw, LSE, Xetra and Euronext markets are supported.
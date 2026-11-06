---
title: The tickers page lists confirmed tickers and takes a correction
status: active
origin:
  kind: agent
  recorded_at: '2026-11-02T07:00:24.746Z'
id: SCEN-ui-tickers-correct
type: scenario
---
Given confirmed tickers, when the owner opens the tickers page, then each is listed once per institution with its ISIN, trading currency and rows booked and a form to correct the ISIN, with a currency choice on London; the owner's correction reaches the core trimmed and in upper case and the page says how many rows were booked anew, or shows the core's refusal; a read session can't post one.
---
title: A test pinning the embedded catalog fails the architecture test
status: active
origin:
  kind: agent
  recorded_at: '2026-11-06T17:50:00.614Z'
id: SCEN-tooling-archtest-catalog-pinned
type: scenario
---
Given tests that compare len of catalog.Load() entries with 0, read entry 0, spell out a pinned digest, or keep the result of a function that reads the catalog without calling catalog.Load() themselves, when the catalog-tests rule runs, then each is reported by file and test; a test deriving its expectation from catalog.Load() through a helper, and one checking only the error, are not.

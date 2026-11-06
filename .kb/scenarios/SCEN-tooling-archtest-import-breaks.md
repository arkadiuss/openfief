---
title: An import the map or the layers forbid fails the architecture test
status: active
origin:
  kind: agent
  recorded_at: '2026-11-06T17:49:43.373Z'
id: SCEN-tooling-archtest-import-breaks
type: scenario
---
Given a package of ingest's app layer importing portfolio, a domain package importing its own app package, an app package importing an adapter, an adapter importing another context's adapter, a plugin importing internal/kernel, or a package named util, when the architecture rules run, then each is reported by its import edge or package. Allowed arrows, cmd importing adapters, and an adapter's own subpackages are not.

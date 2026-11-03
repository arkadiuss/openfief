---
title: An instance name or a Secret role holding two hyphens in a row is refused
status: active
origin:
  kind: agent
  recorded_at: '2026-11-03T18:23:41.472Z'
id: SCEN-plugins-double-hyphen-refused
type: scenario
---
Given a Plugin or a grant named sample--a, a manifest with Secret role flex--token, or an image whose repository is named sample--plugin, when the plugin is validated, applied to the API server, installed or bound, then the domain refuses it with ErrSecretNameSeparator, the CRDs refuse it as Invalid, the default instance name is empty and a connector binding refuses the instance name.
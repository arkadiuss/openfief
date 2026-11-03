---
title: The consent screens show vendor text escaped
status: active
origin:
  kind: agent
  recorded_at: '2026-11-03T09:18:10.803Z'
id: SCEN-plugins-cli-escapes-vendor-text
type: scenario
---
Given a manifest, catalog entry or cluster status whose text holds a control or bidi character, when the CLI shows an install, an upgrade, an uninstall or the instance list, then each such string is printed Go-quoted with visible escapes, and whatever the CLI writes through say has every remaining control character escaped, so the screen the owner confirms can't be rewritten.
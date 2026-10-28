---
title: Trust documents are read strictly
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T08:51:01.854Z'
id: SCEN-plugins-catalog-strict
type: scenario
---
Given a catalog or owner-signer document, when it pins a publisher by pattern, names an image by tag, lacks a name or identity, or has a field the format does not define, then the whole document is refused and nothing in it is trusted.
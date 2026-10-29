---
title: Every UI page draws from the design system and writes no color of its own
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T09:49:42.730Z'
id: SCEN-ui-design-system
type: scenario
---
Given the built-in UI's embedded templates, scripts and stylesheet, when any page is rendered, then it loads the design system's favicon, fonts and tokens from the UI itself before the UI's stylesheet, the stylesheet uses only defined tokens, no file writes a color outside design/tokens.css, and Kasztan appears only in one of his four poses. Assumes the vendored libraries are not the UI's own code.
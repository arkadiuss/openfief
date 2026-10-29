---
title: The built-in UI uses only the design system's tokens, fonts, marks and Kasztan, and groups plugin trust into D-87's badges
status: active
sourceFile: internal/views/ui/design_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T09:49:40.148Z'
id: TEST-ui-design
type: test
---
Walks the embedded templates, scripts and stylesheet (vendored libraries excluded) and fails on any hex code or color function, so every color comes from design/tokens.css; checks every custom property style.css uses is defined by tokens.css or style.css; checks every page's layout loads the favicon, fonts.css, tokens.css and style.css in that order; serves the design files through the UI's handler and checks their content types and that every font fonts.css names exists; checks templates use only the four Kasztan poses; and table-tests badgeOf over every signature trust, an unknown one included. A false pass would need a color written in a way the pattern misses, such as a named color.
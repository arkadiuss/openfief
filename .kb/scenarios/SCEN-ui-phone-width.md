---
title: A page at phone width keeps its figures and never scrolls sideways
status: active
origin:
  kind: agent
  recorded_at: '2026-10-30T08:07:50.776Z'
id: SCEN-ui-phone-width
type: scenario
---
Given the owner signed in on a phone 360 or 390 px wide, when any page opens, then the page scrolls only vertically; the navigation sits behind a Menu button that opens it as a full-width list; each table row reads as a card of header and value lines with every amount still monospaced, right-aligned and in its currency; charts take the screen's width; fields and buttons span the form, at least 44 px tall. Without JavaScript the navigation stays open and wide tables scroll inside their box.
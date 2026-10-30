---
title: A view plugin reads and changes the owner's settings through the core
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T22:59:02.376Z'
id: SCEN-valuation-settings-service
type: scenario
---
Given an owner session, when the UI asks for the settings, then the core answers the reporting currency, the number format with whether it is the default, the formats on offer, the wrappers and the marginal rate; the owner sets an offered format and a rate between 0 and 1; anything else, or another caller, is refused with its code.
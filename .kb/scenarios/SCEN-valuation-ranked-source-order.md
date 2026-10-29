---
title: The owner's order first, then each source in its default place
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T18:00:50.034Z'
id: SCEN-valuation-ranked-source-order
type: scenario
---
Given price sources prices-eu from eodhd, ibkr-main from ibkr and zz-prices from a vendor, with no order set, prices-eu ranks first and the rest follow by name; when the owner places zz-prices and a source that never delivered, those two come first in that order; a source named twice is refused.
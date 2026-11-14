---
title: An entry pins hosts and sets reasoning for its own requests on OpenRouter only
status: active
origin:
  kind: agent
  ref: 'M4 prep brief from the design session: the demo''s model behind the switch (D-207 amended, D-205, D-208, D-212, D-203), 2026-11-14'
  recorded_at: '2026-11-14T17:59:30.654Z'
id: SCEN-models-entry-hosts-reasoning
type: scenario
---
Given an instance at openrouter.ai with one entry that lists hosts and another that sets reasoning low, when each sends a request, then the body holds provider.order with that entry's hosts, allow_fallbacks false and require_parameters, or reasoning effort low, for that entry alone; an entry with neither sends what it always did, the shared client is not changed, and a config that sets hosts or reasoning on any other base URL, a host that is no slug, a host twice, too many hosts or an unknown reasoning value is refused, as the manifest's schema refuses them.
---
title: The core's front door finds the browser's address from the connection, or from the header of a configured reverse proxy only when the connection comes from one of the proxies it is told to trust (X-Forwarded-For read from its right end past the trusted proxies, or one header with a single address), hands it to the view plugin in X-Openfief-Client-Ip after dropping any value the browser sent, gives no address when a trusted proxy says nothing usable, and refuses at start a header without proxies and proxies without a header.
status: open
firmness: default
semantic_text: The core's front door finds the browser's address from the connection, or from the header of a configured reverse proxy only when the connection comes from one of the proxies it is told to trust (X-Forwarded-For read from its right end past the trusted proxies, or one header with a single address), hands it to the view plugin in X-Openfief-Client-Ip after dropping any value the browser sent, gives no address when a trusted proxy says nothing usable, and refuses at start a header without proxies and proxies without a header.
origin:
  kind: agent
  ref: 'M5 brief from the design session: the demo''s spend guards (D-208), plus the M4 follow-ups, 2026-11-14'
  recorded_at: '2026-11-14T20:37:48.020Z'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 126976f9587288f9e13bb717b91024836909d0ff5924dccb727621cc007aebe0
semantic_inventory:
  - claim_key: CLAIM-7370A8AFDD246278
    claim_text: The core's front door finds the browser's address from the connection, or from the header of a configured reverse proxy only when the connection comes from one of the proxies it is told to trust (X-Forwarded-For read from its right end past the trusted proxies, or one header with a single address), hands it to the view plugin in X-Openfief-Client-Ip after dropping any value the browser sent, gives no address when a trusted proxy says nothing usable, and refuses at start a header without proxies and proxies without a header
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 528
    subjects:
      - front_door.client_address
logic_claims:
  - CLAIM-7370A8AFDD246278
id: REQ-views-client-address-trusted-proxy
type: req
---
The core's front door finds the browser's address from the connection, or from the header of a configured reverse proxy only when the connection comes from one of the proxies it is told to trust (X-Forwarded-For read from its right end past the trusted proxies, or one header with a single address), hands it to the view plugin in X-Openfief-Client-Ip after dropping any value the browser sent, gives no address when a trusted proxy says nothing usable, and refuses at start a header without proxies and proxies without a header.

## Context

The per-address limit rests on this address, and a header any browser can write counts for nothing unless only the proxy in front can have written it. The core strips client-sent forwarding headers toward plugins today; this adds the one value a plugin may trust, and only from proxies the owner named.

## Source

> The design session's M5 corrections (2026-11-14): Where the IP comes from behind the demo's ingress: trusted proxy header config, only trusting the header from the configured proxy, documented (M5 brief, 2026-11-14).
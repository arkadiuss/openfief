---
title: The end-to-end suite's institution fixture server must serve only synthetic answers, over HTTPS with a certificate and key it is given, for the IBKR Flex Web Service, EODHD and NBP table A endpoints the connectors call, each below its own path that a connector base URL can name, so that the connectors' own clients read them and no end-to-end test calls a real institution.
status: open
text_ref: design-session:connector-endpoint-ca
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-session-connector-endpoint-ca
  - lane:strict
semantic_text: The end-to-end suite's institution fixture server must serve only synthetic answers, over HTTPS with a certificate and key it is given, for the IBKR Flex Web Service, EODHD and NBP table A endpoints the connectors call, each below its own path that a connector base URL can name, so that the connectors' own clients read them and no end-to-end test calls a real institution.
logic_claims:
  - CLAIM-03537CCD6AEE3528
semantic_clauses:
  - The end-to-end suite's institution fixture server must serve only synthetic answers, over HTTPS with a certificate and key it is given, for the IBKR Flex Web Service, EODHD and NBP table A endpoints the connectors call, each below its own path that a connector base URL can name, so that the connectors' own clients read them and no end-to-end test calls a real institution.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 9cb5b9e55650767e7ca11867e434d261e67c39e83e2feab3f1350781a1fb8091
semantic_inventory:
  - claim_key: CLAIM-03537CCD6AEE3528
    claim_text: The end-to-end suite's institution fixture server must serve only synthetic answers, over HTTPS with a certificate and key it is given, for the IBKR Flex Web Service, EODHD and NBP table A endpoints the connectors call, each below its own path that a connector base URL can name, so that the connectors' own clients read them and no end-to-end test calls a real institution
    role: normative
    status: modeled
    span:
      start: 0
      end: 373
origin:
  kind: agent
  recorded_at: '2026-11-06T22:44:09.915Z'
id: REQ-e2e-institution-fixture-server
type: req
---
The end-to-end suite's institution fixture server must serve only synthetic answers, over HTTPS with a certificate and key it is given, for the IBKR Flex Web Service, EODHD and NBP table A endpoints the connectors call, each below its own path that a connector base URL can name, so that the connectors' own clients read them and no end-to-end test calls a real institution.

## Context

Design session decision for M0 exit end-to-end tests (D-200) and later model providers (D-205), given to the implementation agent by the orchestrator on 2026-11-06: connectors get two config fields, a schema-validated https base URL defaulting to the real host, and an optional extra CA bundle from a Secret added to the system roots, with no test-only switches. docs/dev/testing.md: external institutions are fixture servers in the cluster; end-to-end tests never call a real broker. The data-path tests that use it come next.

## Source

> add an e2e-only fixture server image (synthetic responses only, never real account data): one small Go HTTPS server serving the IBKR Flex, EODHD and NBP endpoints the connectors call, with a per-run CA, built by scripts/e2e-images.sh.

design-session:connector-endpoint-ca

---
title: task grafana:apply routes the demo's alerts to Slack through a Grafana IRM integration, idempotently, and keeps the integration's link secret
status: open
firmness: default
origin:
  kind: agent
  ref: 'Author request: route the demo''s Grafana alerts to Slack #openfief-demo through Grafana IRM, provisioned by task grafana:apply (docs/dev/demo.md#grafana), 2026-11-20'
  recorded_at: '2026-11-20T15:00:00.000Z'
semantic_text: 'task grafana:apply provisions the route of the demo''s alerts to Slack: it finds the Grafana IRM integration named in deploy/grafana/notifications.json through the OnCall API, creates it when absent and sets its default route to the Slack channel when that differs, upserts a contact point of type oncall at the integration''s link, and puts one first-level route for the demo folder into the notification policy tree, replacing only a route with the same receiver and keeping every other route and the default receiver, so a second run changes nothing; it fails with a clear error and a non-zero exit when the IRM API fails; the integration''s link is read from the API into the contact point''s request and is in no argument, file, output or git; task grafana:check validates that notifications.json names a Slack channel ID and holds no URL.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 782a7cf4403118b6f5570255bee89e48aa644c9cce31a2b2720b22ff4e6f2bb2
semantic_inventory:
  - claim_key: CLAIM-D191D582DEE40AD9
    claim_text: 'task grafana:apply provisions the route of the demo''s alerts to Slack: it finds the Grafana IRM integration named in deploy/grafana/notifications.json through the OnCall API, creates it when absent and sets its default route to the Slack channel when that differs, upserts a contact point of type oncall at the integration''s link, and puts one first-level route for the demo folder into the notification policy tree, replacing only a route with the same receiver and keeping every other route and the default receiver, so a second run changes nothing'
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 550
    subjects:
      - demo.grafana_notifications
  - claim_key: CLAIM-378BBEB70068968F
    claim_text: it fails with a clear error and a non-zero exit when the IRM API fails
    role: normative
    status: ontology_gap
    span:
      start: 552
      end: 622
    subjects:
      - demo.grafana_notifications
  - claim_key: CLAIM-712A6100D74BC2BB
    claim_text: the integration's link is read from the API into the contact point's request and is in no argument, file, output or git
    role: descriptive
    status: missing
    span:
      start: 624
      end: 743
  - claim_key: CLAIM-16C50A7E12A53DF5
    claim_text: task grafana:check validates that notifications.json names a Slack channel ID and holds no URL
    role: descriptive
    status: missing
    span:
      start: 745
      end: 839
logic_claims:
  - CLAIM-D191D582DEE40AD9
  - CLAIM-378BBEB70068968F
  - CLAIM-712A6100D74BC2BB
  - CLAIM-16C50A7E12A53DF5
id: REQ-demo-grafana-notifications
type: req
---
task grafana:apply provisions the route of the demo's alerts to Slack: it finds the Grafana IRM integration named in deploy/grafana/notifications.json through the OnCall API, creates it when absent and sets its default route to the Slack channel when that differs, upserts a contact point of type oncall at the integration's link, and puts one first-level route for the demo folder into the notification policy tree, replacing only a route with the same receiver and keeping every other route and the default receiver, so a second run changes nothing; it fails with a clear error and a non-zero exit when the IRM API fails; the integration's link is read from the API into the contact point's request and is in no argument, file, output or git; task grafana:check validates that notifications.json names a Slack channel ID and holds no URL.

## Context

The alert rules of REQ-demo-grafana-as-code had no contact point, so a firing rule only showed in Grafana's alerting page. The author set up an IRM integration, a contact point and a policy route by hand and asked for apply to reproduce them. The integration's link lets whoever holds it raise alerts in the channel, so it is a credential that only Grafana keeps.

## Source

> Make task grafana:apply also provision the demo alerts' route to Slack (the author's request, 2026-11-20).

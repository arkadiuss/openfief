---
title: Only labelled keys from the allowed CIDRs reach SSH
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T08:59:29.662Z'
id: SCEN-tooling-demo-ssh-access
type: scenario
---
Given no SSH key labelled openfief-ssh=true, when task demo:up runs, then it refuses and explains how to label a key. Given a labelled key and DEMO_SSH_CIDRS=203.0.113.7/32, then the server gets that key and a firewall whose only inbound rule is 22/tcp from that CIDR; with no CIDRs the firewall has no inbound rule at all.
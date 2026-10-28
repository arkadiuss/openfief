---
title: SSH only from allowed CIDRs with labelled keys, HTTP(S) as configured
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T08:59:29.662Z'
id: SCEN-tooling-demo-ssh-access
type: scenario
---
Given no SSH key labelled openfief-ssh=true, when task demo:up runs, then it refuses and explains how to label a key. Given a labelled key, DEMO_SSH_CIDRS=203.0.113.7/32 and no DEMO_HTTP_CIDRS, then the server gets that key and a firewall whose inbound rules are 22/tcp from that CIDR and 80 and 443/tcp from everyone; with DEMO_HTTP_CIDRS=none only the SSH rule is left, and with no SSH CIDRs there is no SSH rule.
---
title: '--print renders the Plugin and PluginGrant for GitOps'
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T23:13:58.294Z'
id: SCEN-plugins-cli-print-yaml
type: scenario
---
Given an install of sample that requests the connector capability, api.example.com:443 and the Secret sample-credentials at /etc/openfief/sample with a config, when the owner asks to print instead of apply, then the CLI writes one YAML stream with an openfief.io/v1alpha1 Plugin and PluginGrant named sample in namespace openfief, the grant approving exactly the request, with no status and nothing the API server sets, and changes nothing in the cluster.
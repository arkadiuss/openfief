---
status: active
origin:
  kind: agent
  recorded_at: '2026-11-13T02:55:10.085Z'
title: The staged check writes no throwaway store and the hook clears dead ones
id: SCEN-tooling-kibi-staged-store-unsaved
type: scenario
---
Given a staged change that touches a changed symbol, when kibi check --staged projects the branch's entities and relationships into its throwaway store, then nothing is written to that store's kb.rdf and the store is removed on exit; given throwaway stores of processes that are gone in the temporary directory, when task kibi:staged runs, then those are removed before the checks, stores of running processes are left alone, and both checks still run in order.
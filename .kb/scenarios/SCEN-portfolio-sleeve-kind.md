---
title: A sleeve kind needs a key, a name and a default role, and a sleeve needs a stored kind
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T09:34:22.408Z'
id: SCEN-portfolio-sleeve-kind
type: scenario
---
Given the built-in kinds, when the owner adds a kind with key crypto, name Crypto and default role offensive, then it is valid; when the key is empty, has uppercase letters, spaces or empty hyphen-separated words, or the name or default role is missing, then it is rejected as an invalid sleeve kind; when a sleeve names no kind or a kind that is not stored, then it is rejected as an unknown sleeve kind.
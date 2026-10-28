---
title: The IBKR Flex plugin maps a statement in the mode the core asks for
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T17:06:58.799Z'
id: SCEN-plugins-ibkr-flex-modes
type: scenario
---
Given an IBKR Flex plugin whose Secret holds a Flex token and query ID and whose config maps one IBKR account to acc-ibkr, when the core syncs it in opening mode, then the Flex statement is fetched with those credentials and mapped into opening lots, opening cash and balance assertions; in transactions mode the same statement maps into transactions without opening lots; an uploaded statement is mapped the same way; and a missing credential fails without revealing the token.
---
title: The daily pull fetches a Flex statement once IBKR has generated it
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T06:34:04.891Z'
id: SCEN-ibkrflex-web-service-pull
type: scenario
---
Given a Flex token and query ID, when the client calls SendRequest and IBKR answers with reference code 1111111111, then it waits and calls GetStatement with that code; while IBKR answers that the statement is still being generated it waits longer and retries, up to its attempt limit, and when IBKR returns the FlexQueryResponse it hands back that document; an error code that does not mean try again, an HTTP failure or an oversized statement fails the pull with an error that never contains the token.
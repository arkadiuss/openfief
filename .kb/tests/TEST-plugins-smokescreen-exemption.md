---
title: The proxy dials the exempt host at its address, refuses every other private host, and warns of the exemption at start
status: active
sourceFile: internal/plugins/smokescreen/exemption_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-12T23:20:13.216Z'
id: TEST-plugins-smokescreen-exemption
type: test
---
TestProxyExemptsOneHost starts the proxy on a local listener with a fake resolver and an exemption, and sends CONNECTs: the exempt host at its private address is answered 200 and dialed; another host at the same address, another host in the same range, a host with a public and a private address, the exempt host at another address or at a further one, and the exempt host with no exemption set are answered 502 and never dialed; an unlisted host is answered 407. TestProxyExemptionIsRefusedForAnIPLiteral sends the exempt address as the host and gets 407. TestProxyWarnsOfAnExemption reads the start log for the warning naming the exempt host and its address, and none without an exemption. TestProxyLogsExemptionDecisions reads the decision lines: allowed with no error for the exempt host, an error naming a forbidden destination address for the other.
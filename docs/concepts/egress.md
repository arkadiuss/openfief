---
title: Egress
description: How the egress proxy lets each plugin reach only the hostnames its grant allows, and never a private address.
covers:
  - internal/plugins/egress/**
  - internal/plugins/smokescreen/**
  - cmd/egress-proxy/**
  - deploy/egress-proxy/**
---

A connector that pulls IBKR Flex statements needs to reach IBKR's Flex service and nothing else. OpenFief makes that a rule of the network rather than a promise of the plugin: a plugin can't resolve names or reach the internet directly, and every outbound connection goes through an egress proxy that checks who is asking and where to. This page explains why that takes a proxy, how the proxy is built on Stripe's Smokescreen, and how a plugin proves who it is to it. The [Plugins](plugins.md) page covers how a plugin's grant becomes its allow-list, and the Go SDK that sends its token.

## Concepts

**Why a proxy.** Kubernetes [NetworkPolicy](../glossary.md) allows traffic by pod, namespace and IP block, never by hostname (R-13). The owner approves hostnames, but the addresses behind a name change, and a CDN serves many unrelated sites from the same addresses, so an IP rule either goes stale or lets in more than intended. Hostname rules exist only in some network plugins (Cilium's `toFQDNs`, GKE's `FQDNNetworkPolicy`, Calico's commercial editions), and the upstream standard is still provisional, while OpenFief promises any conformant cluster (D-12). An [egress proxy](../glossary.md) works with any CNI that enforces NetworkPolicy: the policy only has to say "plugins may reach the proxy", and the proxy checks hostnames.

**HTTP CONNECT.** With [HTTP CONNECT](../glossary.md), a client asks the proxy for a tunnel to `api.example.com:443` and then runs TLS through it end to end. The proxy sees the hostname and port, never the content, so it holds no certificates and can't read the credentials a plugin sends to its institution.

**Smokescreen.** [Smokescreen](../glossary.md) is Stripe's CONNECT proxy, written in Go and open-sourced around 2016. Stripe built it against server-side request forgery (SSRF): a service that fetches URLs on someone else's behalf, such as one that sends webhooks, can be tricked into fetching the cloud metadata address `169.254.169.254` or an internal admin page. Smokescreen resolves names itself, refuses private and internal addresses, and checks each client against its own allow-list. A replaceable "role function" decides who the client is. R-14 records its state when OpenFief chose it; D-32 and D-63 embed it as a Go library.

**Projected ServiceAccount tokens.** Kubernetes can mount into a pod a short-lived token issued for one audience and bound to that pod; the kubelet replaces the file before it expires. These [projected ServiceAccount tokens](../glossary.md) have been stable since Kubernetes 1.20 (2020). Any service can check one with the [TokenReview](../glossary.md) API, which answers whether it is valid, for which audiences, and whose it is (`system:serviceaccount:<namespace>:<name>`). That is how the proxy learns a request's [plugin identity](../glossary.md) without holding any secret per plugin.

**A token of its own.** The audience says who a token is for. A plugin gets one token for the core (`openfief-core`, D-27) and, only when it has egress, a second one for the proxy (`openfief-egress`, D-63). A token the proxy sees is then worthless at the core, so even a compromised proxy can't replay it there.

**Token in the proxy URL.** HTTP clients already know how to authenticate to a proxy: given `http://user:password@proxy:port`, Go's `net/http` and undici's `ProxyAgent` (Node) send the credentials as a Basic `Proxy-Authorization` header on every CONNECT (R-14). A plugin puts its egress token in the password. Projected tokens rotate, so the URL has to be rebuilt from the token file on each request (D-63).

**Per-plugin allow-lists.** Each plugin has its own [egress allow-list](../glossary.md): the [egress hosts](../glossary.md) of its effective permissions. An entry is one exact hostname and port. There are no wildcards, no subdomains and no IP literals, so `x.api.example.com` and `api.example.com.evil.example.net` don't match `api.example.com`.

**Refusing private and internal ranges.** An allowed hostname is still dangerous if it resolves to an address inside the cluster or to cloud metadata, through a mistake, a hijacked DNS record, or DNS rebinding (a name that answers with a public address once and a private one the next time). So the proxy resolves the name itself, checks every address it gets, and dials only an address it checked.

**Failing closed.** Every unknown is a refusal: no token, an unknown identity, a missing or unreadable allow-list, a forbidden address, a Kubernetes API that can't review a token. Smokescreen's own default is the opposite: with no allow-list configured it allows every destination (R-14), so OpenFief's proxy never runs without one (D-63).

## How it works

| Piece | Code | Job |
|---|---|---|
| Use case | [app/egress.go](../../internal/plugins/app/egress.go) | `app.Egress.Identify(ctx, token)` returns a plugin name; `Decide(plugin, host, port)` returns an [egress decision](../glossary.md), `app.EgressDecision` |
| Token cache | [app/tokencache.go](../../internal/plugins/app/tokencache.go) | `app.CachingTokenReviewer` in front of TokenReview |
| Plugins package | [allowlist.go](../../internal/plugins/allowlist.go), [identity.go](../../internal/plugins/identity.go) | `EgressAllowList.Allows`, `PluginForServiceAccount`, `EgressAudience` |
| Address check | [egress/address.go](../../internal/plugins/egress/address.go) | `ForbiddenPrefixes`, `CheckAddress` |
| Smokescreen adapter | [smokescreen/proxy.go](../../internal/plugins/smokescreen/proxy.go) | `New` and `Proxy.Serve`: `BuildProxy`'s handler in OpenFief's own server, role function, `EgressACL`, deny ranges, checking resolver |
| Kubernetes adapter | [kube/tokenreview.go](../../internal/plugins/kube/tokenreview.go), [kube/allowlists.go](../../internal/plugins/kube/allowlists.go) | `TokenReviewer` for one audience, `WatchAllowLists` |
| Binary | [cmd/egress-proxy/main.go](../../cmd/egress-proxy/main.go) | Wiring, signals, logging |
| Install | [deploy/egress-proxy/](../../deploy/egress-proxy/) | Kustomize base: ServiceAccount, RBAC, Deployment, Service |
| Plugin side | [sdk/go/openfief/egress.go](../../sdk/go/openfief/egress.go) | `NewEgressClient`, `EgressTransport` |

The `app` package declares two ports: `TokenReviewer` (token in, Kubernetes username out) and `AllowListSource` (plugin name in, its current list out). The `kube` package implements both, `CachingTokenReviewer` wraps the first, and the tests use fakes. The `kube` adapter is described with the [Operator](operator.md), which writes the ConfigMap it follows.

### A request through the proxy

```mermaid
sequenceDiagram
  autonumber
  participant P as Plugin (SDK EgressTransport)
  participant X as Egress proxy (Smokescreen)
  participant K as Kubernetes API
  participant R as DNS resolver
  participant E as api.example.com
  Note over P,X: NetworkPolicy lets the plugin reach only the core and the proxy, by ClusterIP, with no DNS
  P->>P: read the egress token file, build the proxy URL with user plugin and the token
  P->>X: CONNECT api.example.com:443, Proxy-Authorization Basic
  X->>X: role function decodes the token from the password
  X->>X: CachingTokenReviewer, fresh entry for this token?
  X->>K: if not cached, TokenReview with audience openfief-egress
  K-->>X: authenticated as system:serviceaccount:openfief:plugin-sample
  X->>X: PluginForServiceAccount gives plugin sample
  X->>X: decider checks CONNECT, no upstream proxy, then app.Egress.Decide
  X->>R: look up api.example.com
  R-->>X: addresses
  X->>X: checkingResolver runs egress.CheckAddress on every address
  X->>E: TCP to a checked address
  X-->>P: 200, tunnel open
  P->>E: TLS through the tunnel
```

Smokescreen asks the role function first, then the decider, and resolves the name only when the decider allows it (`checkIfRequestShouldBeProxied` in [Smokescreen](https://github.com/stripe/smokescreen)). An unlisted host is therefore never looked up.

### Where a request is refused

```mermaid
flowchart TD
  start(["request from a plugin"]) --> hdr{"Proxy-Authorization is Basic,<br/>with a token as the password?"}
  hdr -- no --> deny["407: refused"]
  hdr -- yes --> tr{"cached, or TokenReview says:<br/>authenticated, audience openfief-egress?"}
  tr -- "no, or API error" --> deny
  tr -- yes --> sa{"username is a plugin<br/>ServiceAccount in openfief?"}
  sa -- no --> deny
  sa -- yes --> conn{"CONNECT to host:port,<br/>no X-Upstream-Https-Proxy?"}
  conn -- no --> deny
  conn -- yes --> has{"plugin has an allow-list?"}
  has -- no --> deny
  has -- yes --> on{"exact host and port on it?"}
  on -- no --> deny
  on -- yes --> addr{"resolves, and every address<br/>passes CheckAddress?"}
  addr -- no --> bad["502: refused before dialing"]
  addr -- yes --> ok["dial a checked address,<br/>open the tunnel"]
```

Smokescreen answers a refusal by identity or allow-list with `407 Proxy Authentication Required` and logs one decision line with the role, the host and the reason; the reason is `EgressDecision.Reason`. A forbidden address comes back as a `net.Error`, which Smokescreen answers with `502 Bad Gateway`. OpenFief's errors never contain the token.

### Identifying the plugin

`identify` in [proxy.go](../../internal/plugins/smokescreen/proxy.go) is the role function. `proxyToken` reads `Proxy-Authorization: Basic base64(user:token)`, matching the scheme case-insensitively as RFC 9110 allows. It ignores the user name and takes the password as the token, which must be non-empty and free of whitespace. A missing header is `app.ErrNoToken`. A `Bearer` header, credentials that aren't base64, credentials without a colon, or a token put in the user name instead are all `ErrBadProxyAuth`. `app.Egress.Identify` then sends the token to its `TokenReviewer` and maps the username back with `PluginForServiceAccount`, which accepts only `system:serviceaccount:openfief:plugin-<name>`; any other identity is `ErrNotAPlugin`.

`kube.TokenReviewer` is built for one audience; [main.go](../../cmd/egress-proxy/main.go) passes `plugins.EgressAudience`. It accepts a review only when it is authenticated *and* lists `openfief-egress`, so a plugin's core token is refused here. TokenReview is a cluster-scoped API, so the proxy has a ClusterRole that may only create TokenReviews, narrower than the usual `system:auth-delegator`; [rbac.yaml](../../deploy/egress-proxy/rbac.yaml) notes it is the only cluster-scoped permission OpenFief's components hold.

`app.CachingTokenReviewer` sits in front so that not every CONNECT costs an API call (D-63):

- An accepted token is trusted until the earlier of the TTL (`DefaultTokenCacheTTL`, five minutes) and the token's own `exp` claim. The claim is read without checking the signature, and only from a token the API server has just accepted.
- A refused token is never cached, and neither is one already past its expiry.
- Entries are keyed by the token's SHA-256 digest, so the cache holds no token. It holds at most 1024; when full it drops expired entries, and if that frees nothing it stops caching until it can.
- The TTL must be positive and below a projected token's one-hour lifetime, or `NewCachingTokenReviewer` refuses it.

The trade-off is stated in the code: a token the API server would now refuse, because its pod or ServiceAccount was deleted, stays accepted for up to the TTL.

### Following the allow-lists

The operator publishes every running plugin's list in the ConfigMap `openfief-egress-allowlist` (format on the [Plugins](plugins.md#handing-allow-lists-to-the-egress-proxy) page). `kube.WatchAllowLists` watches that one ConfigMap with an informer limited to the `openfief` namespace and to that name, and the proxy's Role can read nothing else.

- On every add or update it decodes the data with `plugins.DecodeAllowLists` and swaps the whole map in with an `atomic.Pointer`, so each CONNECT is decided against one complete set of lists, and the next CONNECT after a change sees the new one.
- If the data doesn't decode, every list is dropped and the error is logged: "egress allow-lists refused; denying all egress until fixed". Deleting the ConfigMap drops every list too.
- It returns only after its first read. If that read does not finish before the context ends, it returns `ErrNotSynced` and the binary exits instead of serving.

`app.Egress.Decide` then gives one of three reasons. A plugin with no key at all has no list (`ReasonNoAllowList`), and with no `AllowListSource` at all nobody has one. A plugin whose key is empty, such as the planned XTB parser, has a list with no hosts, and every destination is `ReasonNotListed`. Only an exact match is `ReasonListed`. `EgressAllowList.Allows` ignores ASCII case and one trailing dot, as DNS does. A host with any non-ASCII byte never matches, because Unicode case folding would let the Kelvin sign stand in for `k`, and an IP literal never matches because no entry is one. Smokescreen itself also refuses an IPv6 literal destination.

### Checking addresses

`egress.ForbiddenPrefixes` lists the ranges the proxy never dials:

- IPv4: `0.0.0.0/8`, private `10/8`, `172.16/12` and `192.168/16`, carrier-grade NAT `100.64/10`, loopback `127/8`, link-local `169.254/16` (cloud metadata), `192.0.0/24`, the three documentation ranges, `192.88.99/24`, benchmarking `198.18/15`, multicast `224/4` and reserved `240/4`.
- IPv6: `::/96` (unspecified, loopback, IPv4-compatible), local-use NAT64 `64:ff9b:1::/48`, discard `100::/64`, `2001::/23` (including Teredo), documentation `2001:db8::/32` and `3fff::/20`, unique-local `fc00::/7`, link-local `fe80::/10`, site-local `fec0::/10` and multicast `ff00::/8`.

`CheckAddress` refuses an invalid or zoned address. It checks the IPv4 address inside an IPv4-mapped (`::ffff:`), NAT64 (`64:ff9b::/96`) or 6to4 (`2002::/16`) address as IPv4, so a private address can't hide inside an IPv6 one. Anything that passes the list must still be global unicast.

`checkingResolver` wraps Smokescreen's resolver: if *any* address of a name fails `CheckAddress`, the whole name is refused, rather than dialed at its public address. The same prefixes are also handed to Smokescreen as deny ranges, a second layer. `New` turns on Smokescreen's self-connection detection, so a destination on any of the proxy's own addresses is refused on every port, and the dial hook connects to exactly the address Smokescreen picked from the checked set.

### Refusing other uses of the proxy

The `decider` (Smokescreen's `EgressACL`) refuses anything that is not a CONNECT: Smokescreen can also forward plain HTTP requests, in which the proxy would see, and could change, the whole request. It refuses a request that asks Smokescreen to chain through another proxy with `X-Upstream-Https-Proxy`. And `New` refuses to build a proxy when `HTTPS_PROXY` or `https_proxy` is set in the proxy's own environment, because Smokescreen's underlying goproxy would then send every CONNECT through that proxy and skip the resolver check. It also refuses to build one without the egress use case, so Smokescreen never runs without an ACL.

### Running it

`smokescreen.New` does the setup Smokescreen's `StartWithConfig` would do (validating the config, the self-connection guard, a connection tracker), then serves `ss.BuildProxy`'s handler from an `http.Server` it owns, wrapped in Smokescreen's health-check middleware (D-63). Unlike `StartWithConfig`, it never exits the process and installs no signal handlers. A tunnel with no traffic for five minutes is closed. `Proxy.Serve(ctx)` runs until its context ends, then stops accepting, lets open tunnels finish for at most `DefaultDrainTimeout` (25 seconds, inside Kubernetes' default 30-second grace period), and closes the rest itself, since `http.Server.Shutdown` doesn't close hijacked connections.

[main.go](../../cmd/egress-proxy/main.go) listens on `-listen` (default `:4750`), logs JSON, and uses in-cluster credentials. It cancels its context on SIGTERM or an interrupt, starts the allow-list watch, wraps `kube.NewTokenReviewer(client, plugins.EgressAudience)` in a `CachingTokenReviewer`, opens the listener, and serves with a `/healthcheck` that answers 204. [go.mod](../../go.mod) pins Smokescreen to commit 68e7b82 (`v0.1.1-0.20261022111441-68e7b8218e22`), after v0.1.0, as D-63 asks.

The [Dockerfile](../../cmd/egress-proxy/Dockerfile) builds a static binary for amd64 and arm64 on digest-pinned base images, and runs it on distroless as non-root. The kustomize base runs one replica with a read-only root filesystem, no capabilities, no privilege escalation and the runtime's default seccomp profile, behind a ClusterIP Service on port 4750. The Service's `targetPort` is a number because the operator copies it into every plugin's NetworkPolicy. The base names the image without a digest; an overlay pins it. The base defines no NetworkPolicy for the proxy's own pod.

### The plugin's side

When a plugin has granted egress hosts, `plugins.Compile` gives it `HTTPS_PROXY=http://<proxy ClusterIP>:4750`, `NO_PROXY=<core ClusterIP>`, a NetworkPolicy egress rule to the proxy's pod port, and the second projected token at `/var/run/secrets/openfief-egress/token`, whose path the operator puts in `OPENFIEF_EGRESS_TOKEN_FILE`. A plugin without egress hosts, such as the planned XTB parser, gets none of these, so it can reach only the core. No plugin gets a DNS rule.

`HTTPS_PROXY` itself carries no credentials. The SDK's `NewEgressClient` reads both variables, and its `EgressTransport` builds the proxy URL for every request with user `plugin` and the token read from the file at that moment. It refuses plain `http` requests and refuses to send anything when no proxy is set, rather than go around the proxy. A plugin that doesn't use the SDK must do the same: `TestProxyGoClientProxyURL` shows that Go's own client, given such a URL, gets through.

## Decisions

Implemented, fully or in the part named:

- **D-32** Smokescreen embedded as a library, custom role function.
- **D-32** Granted hostnames compiled into per-plugin allow-lists.
- **D-32** Private and internal address ranges refused.
- **D-32** Plugins reach only core and proxy, no DNS.
- **D-63** `BuildProxy` served from OpenFief's own server, draining on shutdown.
- **D-63** Never without an allow-list; lists swapped atomically.
- **D-63** Token as proxy URL password, Basic header.
- **D-63** TokenReview for `openfief-egress`, results cached.
- **D-63** SDK rebuilds the proxy URL per request.
- **D-63** Smokescreen pinned to a commit after v0.1.0.

The code now follows D-63 in every point above. One stale comment remains: `EnvEgressIdentityFile` in [sdk/go/openfief/env.go](../../sdk/go/openfief/env.go), lines 25 to 28, says the operator doesn't mount the egress token yet, but it does, under that name.

## Verification

Run `task test` from the repository root. The egress packages need no cluster or network: `go test ./internal/plugins/egress/... ./internal/plugins/smokescreen/... ./internal/plugins/kube/... ./internal/plugins/app/...` runs them alone, and `go test ./...` inside `sdk/go` runs the SDK's. The proxy test starts a real Smokescreen on `127.0.0.1` with a fake resolver (names map to `8.8.4.4` or `10.0.0.7`), a dialer that records each address and connects to a local TLS server instead, and synthetic tokens. The kube tests use client-go's fake clientset; the cache tests use a fake clock. Requirement IDs come from kibi (`node_modules/.bin/kibi search "<question>"`).

| Behavior | Tests | Requirement |
|---|---|---|
| Only a Basic proxy password that TokenReview accepts for `openfief-egress` as a plugin identifies a request | [proxy_test.go](../../internal/plugins/smokescreen/proxy_test.go): `TestProxyConnect` ("no token", "token not accepted", "bearer instead of basic", "token as the user name", "no user name"); [egress_test.go](../../internal/plugins/app/egress_test.go): `TestEgressIdentify`; [kube_test.go](../../internal/plugins/kube/kube_test.go): `TestTokenReviewerReviewToken` ("authenticated for the core audience only") | REQ-plugins-egress-proxy-identifies-plugin |
| Go's client, given the token as the proxy URL's password, gets through | `TestProxyGoClientProxyURL` | REQ-plugins-egress-proxy-identifies-plugin |
| Accepted reviews cached within TTL and expiry; refused never; bounded | [tokencache_test.go](../../internal/plugins/app/tokencache_test.go): `TestCachingTokenReviewer`, `TestCachingTokenReviewerBounded`, `TestNewCachingTokenReviewerTTL` | REQ-plugins-egress-token-review-cache |
| Only CONNECT, only to the plugin's list, no chaining | `TestProxyConnect` ("unlisted host", "upstream proxy requested", "plain HTTP proxying"); `TestEgressDecide` | REQ-plugins-egress-proxy-connect-only-allow-list |
| Exact host and port match | [allowlist_test.go](../../internal/plugins/allowlist_test.go): `TestEgressAllowListAllows` | REQ-plugins-allow-list-exact-match |
| The proxy resolves, refuses a name with any forbidden address, dials only checked ones | `TestProxyConnect` ("listed host resolves to a private address", "... public and private addresses"); `TestProxyDialsOnlyCheckedAddresses` | REQ-plugins-egress-proxy-checks-every-address |
| Private, internal and embedded IPv4 ranges refused | [address_test.go](../../internal/plugins/egress/address_test.go): `TestCheckAddress`, `TestCheckAddressInvalid`, `TestForbiddenPrefixes` | REQ-plugins-egress-forbidden-addresses |
| No list, or no source of lists, reaches nothing | `TestProxyConnect` ("plugin without an allow-list"); `TestProxyWithoutAllowListSource`; `TestEgressDecide`, `TestEgressDecideWithoutSource`; `TestNewNeedsListenerAndEgress` | REQ-plugins-egress-fail-closed |
| A swapped list applies to the next CONNECT | `TestProxyFollowsSwappedAllowList` | REQ-plugins-egress-allow-lists-swapped-atomically |
| Lists follow the ConfigMap; missing, undecodable or deleted means none | `TestWatchAllowListsFollowsConfigMap`, `TestWatchAllowListsWithoutConfigMap`, `TestWatchAllowListsNotSynced` | REQ-plugins-egress-allow-lists-fail-closed |
| Serve stops on its context, closes leftover tunnels, stops accepting | `TestProxyServeStops`, `TestProxyHealthcheck` | REQ-plugins-egress-proxy-embedded-buildproxy |
| No proxy with `HTTPS_PROXY` set in its own environment | `TestNewRefusesProxyEnvironment` | |
| The SDK sends the current token on every CONNECT | [sdk egress_test.go](../../sdk/go/openfief/egress_test.go): `TestEgressSendsCurrentTokenOnEveryConnect`, `TestEgressRefusals` | REQ-plugins-sdk-token-per-request, REQ-sdk-go-egress-proxy-url |

`TestProxyDialsOnlyCheckedAddresses` is a good one to read first: it sends four CONNECTs (private, mixed, unlisted, allowed) and asserts that exactly one address was dialed, the public one.

Not covered by tests: `cmd/egress-proxy` and the deploy manifests, and anything that needs a real cluster, such as NetworkPolicy actually blocking a direct connection or a DNS query, or the kubelet rotating a real token.

## Not built yet

- **A TypeScript SDK (D-11).** Only Go plugins get the token handling for free; R-14 notes Node's built-in client reads `HTTPS_PROXY` only with `NODE_USE_ENV_PROXY=1`.
- **End-to-end checks.** There is no `test/e2e` suite. P-10's acceptance items 10 and 16 (a refused host, failed bypass attempts, destinations checked in the proxy's logs) are proposals, not decisions.

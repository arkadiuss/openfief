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

**Per-plugin allow-lists.** Each plugin has its own [egress allow-list](../glossary.md): the [egress hosts](../glossary.md) of its effective permissions. An entry is one exact hostname and port. There are no wildcards, no subdomains and no IP literals, so `x.api.example.com` and `api.example.com.evil.example.net` don't match `api.example.com`. A host can also come from the instance's config: a manifest may ask for "the host of config field `url`" instead of a fixed one, and the grant approves the host that config names, which then sits in the allow-list like any other (D-75). It must be the host of an `https` URL, never a wildcard or an IP address.

**The one exception: an in-cluster service.** A plugin may also reach one service in the same cluster that its owner picked at install, such as a Ghostfolio running next to OpenFief (D-74). That traffic doesn't go through the proxy: the proxy refuses private addresses, and a ClusterIP is one. Instead the plugin's NetworkPolicy allows exactly that service's pods on that port, and the plugin gets the service's ClusterIP, so it still needs no DNS. The traffic may be plain HTTP; the install warns about it. See [Operator](operator.md#the-resources-and-what-the-operator-owns).

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
| Log audit | [smokescreen/audit.go](../../internal/plugins/smokescreen/audit.go) | `DecisionLines`, `AuditLog`: the proxy's decision lines checked against each instance's approved hosts |
| Kubernetes adapter | [kube/tokenreview.go](../../internal/plugins/kube/tokenreview.go), [kube/allowlists.go](../../internal/plugins/kube/allowlists.go), [kube/egressaudit.go](../../internal/plugins/kube/egressaudit.go) | `TokenReviewer` for one audience, `WatchAllowLists`; `EgressProxyLog` and `GrantedEgress` for the audit |
| Binary | [cmd/egress-proxy/main.go](../../cmd/egress-proxy/main.go), [audit.go](../../cmd/egress-proxy/audit.go) | Wiring, signals, logging; the `audit` subcommand |
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

`app.CachingTokenReviewer` sits in front so that not every CONNECT costs an API call (D-63), and so that no caller can make the proxy, or the core, which uses the same cache, ask the API server at will (D-184):

- A token that isn't shaped like a JSON Web Token (`JWTShaped`: at most 8 KiB, three base64url parts whose header and claims are JSON objects) is refused without asking. A projected ServiceAccount token always is one.
- An accepted token is trusted until the earlier of the TTL (`DefaultTokenCacheTTL`, five minutes) and the token's own `exp` claim. The claim is read without checking the signature, and only from a token the API server has just accepted.
- A refused token stays refused for `RefusalCacheTTL`, ten seconds, without asking again; a token already past its expiry isn't cached, and neither is a review that failed, such as one the API server didn't answer.
- Entries are keyed by the token's SHA-256 digest, so the cache holds no token. Accepted and refused tokens are kept apart, each at most 1024, so refusals never crowd out accepted tokens; when one is full it drops expired entries, and if that frees nothing it stops caching until it can.
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

The proxy also refuses the node it runs on and, when the owner names them, a cluster's own ranges that the list can't know (D-184): `egress.ParseDenyList` reads `-deny-cidr`, comma-separated ranges such as a cluster's public or global IPv6 pod, Service and node ranges, and `OPENFIEF_NODE_IP`, which the Deployment sets to the pod's `status.hostIP`, refused as a single address. On a single-node install with a public address, such as the demo's, that keeps a plugin from reaching the node's own ports, the API server's and the kubelet's among them, through an allowed name that resolves to it. `CheckAddressDenying` is `CheckAddress` that also refuses those ranges, IPv4-mapped, NAT64 and 6to4 forms included. A value that doesn't parse stops the proxy at start.

`checkingResolver` wraps Smokescreen's resolver: if *any* address of a name fails `CheckAddressDenying`, the whole name is refused, rather than dialed at its public address. The same prefixes, the extra ones included, are also handed to Smokescreen as deny ranges, a second layer. `New` turns on Smokescreen's self-connection detection, so a destination on any of the proxy's own addresses is refused on every port, and the dial hook connects to exactly the address Smokescreen picked from the checked set.

### Refusing other uses of the proxy

The `decider` (Smokescreen's `EgressACL`) refuses anything that is not a CONNECT: Smokescreen can also forward plain HTTP requests, in which the proxy would see, and could change, the whole request. It refuses a request that asks Smokescreen to chain through another proxy with `X-Upstream-Https-Proxy`. And `New` refuses to build a proxy when `HTTPS_PROXY` or `https_proxy` is set in the proxy's own environment, because Smokescreen's underlying goproxy would then send every CONNECT through that proxy and skip the resolver check. It also refuses to build one without the egress use case, so Smokescreen never runs without an ACL.

### Running it

`smokescreen.New` does the setup Smokescreen's `StartWithConfig` would do (validating the config, the self-connection guard, a connection tracker), then serves `ss.BuildProxy`'s handler from an `http.Server` it owns, wrapped in Smokescreen's health-check middleware (D-63). Unlike `StartWithConfig`, it never exits the process and installs no signal handlers. A tunnel with no traffic for five minutes is closed. `Proxy.Serve(ctx)` runs until its context ends, then stops accepting, lets open tunnels finish for at most `DefaultDrainTimeout` (25 seconds, inside Kubernetes' default 30-second grace period), and closes the rest itself, since `http.Server.Shutdown` doesn't close hijacked connections.

[main.go](../../cmd/egress-proxy/main.go) listens on `-listen` (default `:4750`), refuses the ranges `-deny-cidr` and `OPENFIEF_NODE_IP` name, logs JSON, and uses in-cluster credentials. It cancels its context on SIGTERM or an interrupt, starts the allow-list watch, wraps `kube.NewTokenReviewer(client, plugins.EgressAudience)` in a `CachingTokenReviewer`, opens the listener, and serves with a `/healthcheck` that answers 204. [go.mod](../../go.mod) pins Smokescreen to commit 68e7b82 (`v0.1.1-0.20261022111441-68e7b8218e22`), after v0.1.0, as D-63 asks.

The [Dockerfile](../../cmd/egress-proxy/Dockerfile) builds a static binary for amd64 and arm64 on digest-pinned base images, and runs it on distroless as non-root. The kustomize base runs one replica with a read-only root filesystem, no capabilities, no privilege escalation and the runtime's default seccomp profile, behind a ClusterIP Service on port 4750. The Service's `targetPort` is a number because the operator copies it into every plugin's NetworkPolicy. The base names the image without a digest; an overlay pins it.

The proxy's own NetworkPolicy, [networkpolicy.yaml](../../deploy/egress-proxy/networkpolicy.yaml), admits only its callers (D-184): plugin pods, labelled `app.kubernetes.io/name: openfief-plugin`, on port 4750. Its egress is DNS, the API server on 6443 (TokenReview and the allow-list ConfigMap, at any address for the reason D-92 gives the core), and the internet with every private, shared, link-local and loopback range left out: `10/8`, `172.16/12`, `192.168/16`, `100.64/10`, `169.254/16`, `127/8`, and `fc00::/7`, `fe80::/10`, `::1`. Most clusters put pods, Services and nodes in those ranges, so even a proxy turned against its own checks can't open a connection to another pod, a Service or a node. A cluster that uses public or global IPv6 ranges for any of them adds them to each `except` list with an overlay, and passes them with `-deny-cidr`.

### Auditing the log

Smokescreen logs one JSON decision line per request it decides, with the message `CANONICAL-PROXY-DECISION`: who asked (`role`, the plugin instance), where to (`proxy_type` and `requested_host`, a host and port), the decision (`allow`, `decision_reason`) and its timing. It never logs a path, a query, a header, a body or the token, since a CONNECT tunnel carries those inside TLS and the token goes to TokenReview only. D-200's criterion 16 asks that outbound traffic go only to configured destinations, verified from these lines.

`smokescreen.AuditLog` checks a log against the hosts each instance may reach: every decision line may hold only the fields above, `requested_host` must be a bare host and port, `role` and `allow` must be there, and every allowed request must be to a host the asking instance's hosts include. It returns the counts, the allowed destinations by instance and host, and one problem per broken rule, naming the line's id, the instance and the field, never a value that could hold a request's contents.

`egress-proxy audit` runs it against a cluster, and serves nothing: it reads every egress proxy pod's log (`kube.EgressProxyLog`) and every `PluginGrant` in `openfief` (`kube.GrantedEgress`), with in-cluster credentials in a pod and the kubeconfig (`-kubeconfig`, `-context`) elsewhere, prints the counts and destinations, and exits non-zero on a problem. It only reads. Three places run the check:

- The demo, after each deploy: a Job with the proxy's own image and a Role that may only list pods and grants and read pod logs ([deploy/demo/audit](../../deploy/demo/audit/kustomization.yaml)). There the proxy reaches the real NBP API, so this is the check on real traffic.
- The release smoke test, on kind after the install guide's steps, with the binary built from the checkout ([pipeline.md](../dev/pipeline.md#releases)).
- The end-to-end suite's isolation test, against each instance's allow-list rather than its grant, which is stricter, since the allow-list is the grant cut down to what the plugin requests.

The demo and the smoke test read grants, not allow-lists: a grant is what the owner approved, and an allowed request to a host off the grant means the proxy itself went wrong. Both read the logs as they are when they run, so a proxy pod that restarted has lost what it logged before.

### The plugin's side

When a plugin has granted egress hosts, `plugins.Compile` gives it `HTTPS_PROXY=http://<proxy ClusterIP>:4750`, `NO_PROXY=<core ClusterIP>`, a NetworkPolicy egress rule to the proxy's pod port, and the second projected token at `/var/run/secrets/openfief-egress/token`, whose path the operator puts in `OPENFIEF_EGRESS_TOKEN_FILE`. A plugin without egress hosts, such as the planned XTB parser, gets none of these, so it can reach only the core. No plugin gets a DNS rule.

`HTTPS_PROXY` itself carries no credentials. The SDK's `NewEgressClient` reads both variables, and its `EgressTransport` builds the proxy URL for every request with user `plugin` and the token read from the file at that moment. It refuses plain `http` requests and refuses to send anything when no proxy is set, rather than go around the proxy. A plugin that doesn't use the SDK must do the same: `TestProxyGoClientProxyURL` shows that Go's own client, given such a URL, gets through.

### The test-only exemption

A cluster's Services have private addresses, so with every private range refused no plugin in the end-to-end suite could ever open an allowed tunnel, and D-200's criterion 16 would be checked on refusals alone. D-215 makes one exception to D-32, for tests alone. `-refused-range-exempt-host=<host>=<address>[,<address>...]`, or `OPENFIEF_REFUSED_RANGE_EXEMPT_HOST` when the flag is empty, exempts exactly one host name: `egress.ParseExemption` reads it, and `egress.CheckResolved`, which `checkingResolver` calls, lets that name through when it resolves to exactly the listed addresses, and checks every other name as before. Smokescreen's own range check is skipped for that exact name, which the resolver has already vouched for. There is no range and no wildcard form, and everything else stays refused: another host at the exempt address, another address in the same range, the exempt host at another or an additional address, a subdomain of it, and the exempt address as an IP literal, since the exemption is the name together with its addresses. The exempt host can't itself be an IP address. The proxy logs a warning at start whenever an exemption is set, naming the host and its addresses (`smokescreen.ExemptionWarning`).

Only the end-to-end suite sets it: its install starts a fixture in a namespace of its own, reads the Service's ClusterIP, adds the flag to the proxy's Deployment, and an overlay patch ([overlays/openfief](../../test/e2e/overlays/openfief/kustomization.yaml)) adds to the proxy's NetworkPolicy one egress rule to that fixture's pods on their port. `deploy/` never sets it, and [release-bundle.sh](../../scripts/release-bundle.sh) refuses a bundle whose manifests do, as a flag or as the variable, so a shipped install never has one.

## Decisions

Implemented, fully or in the part named:

- **D-32** Smokescreen embedded as a library, custom role function.
- **D-32** Granted hostnames compiled into per-plugin allow-lists.
- **D-32** Private and internal address ranges refused.
- **D-32** Plugins reach only core and proxy, no DNS; **D-74** adds the one in-cluster service the owner approved, outside the proxy.
- **D-75** A host from the instance's config (an https URL's host, no wildcard, no IP) is approved by the grant and reached through the proxy like a fixed one.
- **D-215** A test-only exemption of one host, at exactly its addresses, from the refused ranges, with a warning at start and a release bundle that can't carry it.
- **D-63** `BuildProxy` served from OpenFief's own server, draining on shutdown.
- **D-63** Never without an allow-list; lists swapped atomically.
- **D-63** Token as proxy URL password, Basic header.
- **D-63** TokenReview for `openfief-egress`, results cached.
- **D-63** SDK rebuilds the proxy URL per request.
- **D-63** Smokescreen pinned to a commit after v0.1.0.

The code follows D-63 in every point above.

## Verification

Run `task test` from the repository root. The egress packages need no cluster or network: `go test ./internal/plugins/egress/... ./internal/plugins/smokescreen/... ./internal/plugins/kube/... ./internal/plugins/app/...` runs them alone, and `go test ./...` inside `sdk/go` runs the SDK's. The proxy test starts a real Smokescreen on `127.0.0.1` with a fake resolver (names map to `8.8.4.4` or `10.0.0.7`), a dialer that records each address and connects to a local TLS server instead, and synthetic tokens. The kube tests use client-go's fake clientset; the cache tests use a fake clock. Requirement IDs come from kibi (`node_modules/.bin/kibi search "<question>"`).

| Behavior | Tests | Requirement |
|---|---|---|
| Only a Basic proxy password that TokenReview accepts for `openfief-egress` as a plugin identifies a request | [proxy_test.go](../../internal/plugins/smokescreen/proxy_test.go): `TestProxyConnect` ("no token", "token not accepted", "bearer instead of basic", "token as the user name", "no user name"); [egress_test.go](../../internal/plugins/app/egress_test.go): `TestEgressIdentify`; [kube_test.go](../../internal/plugins/kube/kube_test.go): `TestTokenReviewerReviewToken` ("authenticated for the core audience only") | REQ-plugins-egress-proxy-identifies-plugin |
| Go's client, given the token as the proxy URL's password, gets through | `TestProxyGoClientProxyURL` | REQ-plugins-egress-proxy-identifies-plugin |
| Only JWT-shaped tokens are reviewed; accepted reviews cached within TTL and expiry, refusals for ten seconds, failures never; each bounded | [tokencache_test.go](../../internal/plugins/app/tokencache_test.go): `TestCachingTokenReviewer`, `TestCachingTokenReviewerBounded`, `TestCachingTokenReviewerRefusalsBounded`, `TestJWTShaped` | REQ-plugins-token-review-not-amplified |
| Accepted reviews cached within TTL and expiry; bounded | [tokencache_test.go](../../internal/plugins/app/tokencache_test.go): `TestCachingTokenReviewer`, `TestCachingTokenReviewerBounded`, `TestNewCachingTokenReviewerTTL` | REQ-plugins-egress-token-review-cache |
| Only CONNECT, only to the plugin's list, no chaining | `TestProxyConnect` ("unlisted host", "upstream proxy requested", "plain HTTP proxying"); `TestEgressDecide` | REQ-plugins-egress-proxy-connect-only-allow-list |
| Exact host and port match | [allowlist_test.go](../../internal/plugins/allowlist_test.go): `TestEgressAllowListAllows` | REQ-plugins-allow-list-exact-match |
| The proxy resolves, refuses a name with any forbidden address, dials only checked ones | `TestProxyConnect` ("listed host resolves to a private address", "... public and private addresses"); `TestProxyDialsOnlyCheckedAddresses` | REQ-plugins-egress-proxy-checks-every-address |
| The node's own address and the ranges the owner names refused too; a bad value is refused at start; the proxy policy admits only plugin pods and keeps private ranges out of its egress | [deny_test.go](../../internal/plugins/egress/deny_test.go): `TestParseDenyList`, `TestCheckAddressDenying`; [proxy_test.go](../../internal/plugins/smokescreen/proxy_test.go): `TestProxyDeniesExtraRanges`; [network-policies.test.sh](../../scripts/tests/network-policies.test.sh) | REQ-plugins-egress-denies-node-and-cluster-ranges, REQ-deploy-network-policies |
| The test-only exemption parses and covers one host at exactly its addresses; another host, another address, a further address and the address as an IP literal stay refused; the proxy dials the exempt host, warns at start and logs the decision | [exemption_test.go](../../internal/plugins/egress/exemption_test.go): `TestParseExemption`, `TestCheckResolved`; [exemption_test.go](../../internal/plugins/smokescreen/exemption_test.go): `TestProxyExemptsOneHost`, `TestProxyExemptionIsRefusedForAnIPLiteral`, `TestProxyWarnsOfAnExemption`, `TestProxyLogsExemptionDecisions` | REQ-plugins-egress-test-exemption-one-host, REQ-plugins-egress-test-exemption-warns-at-start |
| A release bundle that sets the exemption is refused | [release-bundle.test.sh](../../scripts/tests/release-bundle.test.sh) | REQ-plugins-egress-test-exemption-never-shipped |
| Private, internal and embedded IPv4 ranges refused | [address_test.go](../../internal/plugins/egress/address_test.go): `TestCheckAddress`, `TestCheckAddressInvalid`, `TestForbiddenPrefixes` | REQ-plugins-egress-forbidden-addresses |
| No list, or no source of lists, reaches nothing | `TestProxyConnect` ("plugin without an allow-list"); `TestProxyWithoutAllowListSource`; `TestEgressDecide`, `TestEgressDecideWithoutSource`; `TestNewNeedsListenerAndEgress` | REQ-plugins-egress-fail-closed |
| A swapped list applies to the next CONNECT | `TestProxyFollowsSwappedAllowList` | REQ-plugins-egress-allow-lists-swapped-atomically |
| Lists follow the ConfigMap; missing, undecodable or deleted means none | `TestWatchAllowListsFollowsConfigMap`, `TestWatchAllowListsWithoutConfigMap`, `TestWatchAllowListsNotSynced` | REQ-plugins-egress-allow-lists-fail-closed |
| Serve stops on its context, closes leftover tunnels, stops accepting | `TestProxyServeStops`, `TestProxyHealthcheck` | REQ-plugins-egress-proxy-embedded-buildproxy |
| No proxy with `HTTPS_PROXY` set in its own environment | `TestNewRefusesProxyEnvironment` | |
| Decision lines hold only who, where and the decision; every allowed request is to an approved host; problems quote no value | [audit_test.go](../../internal/plugins/smokescreen/audit_test.go): `TestAuditLog`, `TestDecisionLines`; [egressaudit_test.go](../../internal/plugins/kube/egressaudit_test.go): `TestEgressProxyLog`, `TestGrantedEgress`; [cmd/egress-proxy/audit_test.go](../../cmd/egress-proxy/audit_test.go): `TestReport` | REQ-plugins-egress-audit-log |
| The demo audits each deploy's proxy log in a Job at the proxy's verified digest; the release smoke test audits after its install | [demo-render-audit.test.sh](../../scripts/tests/demo-render-audit.test.sh), [release-smoke.test.sh](../../scripts/tests/release-smoke.test.sh) | REQ-plugins-egress-audit-runs |
| The SDK sends the current token on every CONNECT | [sdk egress_test.go](../../sdk/go/openfief/egress_test.go): `TestEgressSendsCurrentTokenOnEveryConnect`, `TestEgressRefusals` | REQ-plugins-sdk-token-per-request, REQ-sdk-go-egress-proxy-url |

`TestProxyDialsOnlyCheckedAddresses` is a good one to read first: it sends four CONNECTs (private, mixed, unlisted, allowed) and asserts that exactly one address was dialed, the public one.

Not covered by unit tests: `cmd/egress-proxy`'s wiring, and anything that needs a real cluster, such as the kubelet rotating a real token. The end-to-end suite ([test/e2e](../../test/e2e/main_test.go)) installs the proxy with the demo's manifests on kind or k3d. `TestOperatorDeploysSignedPlugin` checks that the operator publishes exactly the granted host as an operator-deployed plugin's allow-list entry, and removes it when the `Plugin` is deleted. `TestPluginIsolation` sends CONNECTs through the proxy that it must refuse, checks that direct connections and DNS lookups fail, and audits the run's log with `AuditLog`. It also makes the one allowed connection, through the test-only exemption: an instance granted the exempt fixture host gets 200 and an allowed decision line, a second instance granted another Service of the same range gets 502 and a line naming its forbidden address, the start log holds the warning, and `AuditLog` finds the allowed destination on its instance's allow-list.

## Not built yet

- **A TypeScript SDK (D-11).** Only Go plugins get the token handling for free; R-14 notes Node's built-in client reads `HTTPS_PROXY` only with `NODE_USE_ENV_PROXY=1`.
- **Allowed traffic to a real destination in CI.** The end-to-end suite opens an allowed tunnel, to its test-only exempt fixture (D-215), and calls no real host. Allowed traffic to a real destination is audited only on the demo, which reaches NBP, and the release smoke test runs no connector, so its audit sees refusals at most.
- **Nobody is told when the demo's audit fails.** The demo's Job fails the Flux layer `demo-egress-audit`, which the owner sees on the machine; the deploy workflow holds no credential for the cluster, so CI never sees the result.

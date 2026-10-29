---
title: Operator
description: How the Kubernetes operator turns Plugin and PluginGrant resources into isolated plugin workloads, and how much of that is built.
covers:
  - internal/plugins/operator/**
  - internal/plugins/kube/**
  - cmd/operator/**
  - deploy/crds/**
  - deploy/operator/**
  - internal/plugins/cosign/**
  - cmd/sigstore-trusted-root/**
---

The operator runs OpenFief's plugins on Kubernetes (D-13). It watches two custom resources in the `openfief` namespace, `Plugin` and `PluginGrant`, and runs each plugin with only what both allow, as a locked-down [workload](../glossary.md). The operator makes no decisions of its own: the `internal/plugins` package decides, and the operator maps each decision onto Kubernetes objects. It deploys an image only when a [trusted signer](../glossary.md) signed its digest and the image's signed manifest label matches what the `Plugin` declares, checked offline with sigstore-go.

## Concepts

**Custom resources and CRDs.** Kubernetes stores objects of built-in types such as Pods and Services. A CustomResourceDefinition (CRD) adds a new type to the API server, with an OpenAPI schema that the server checks on every write. CRDs replaced ThirdPartyResources in Kubernetes 1.7 (2017) and reached `apiextensions.k8s.io/v1` in 1.16. OpenFief defines exactly two, `Plugin` and `PluginGrant`, in group `openfief.io`, version `v1alpha1` (D-55). The Go types are the source of truth. controller-gen, from the Kubebuilder project, reads markers in their comments and generates the CRD YAML, the deepcopy code and the operator's Role.

**The operator pattern and the reconcile loop.** CoreOS introduced operators in 2016: a controller that encodes how to run one piece of software, driven by its own custom resources. The loop underneath comes from Google's Borg and Omega, as described in "Borg, Omega, and Kubernetes" (Burns et al., 2016). Controllers repeatedly compare the desired state with the observed state and act to remove the difference. The loop is *level-triggered*: it reacts to the current state, not to the event that changed it. A missed or repeated event does no harm, because the next pass looks at everything again. So `Reconcile` gets only a name, re-reads everything and is safe to run any number of times.

**Desired and observed state.** By convention, `spec` holds what someone wants and `status` holds what the controller saw and did. Here the vendor's request (image, [manifest](../glossary.md), config) is the `Plugin`'s spec. The owner's approval is a separate object, the `PluginGrant`, so editing the vendor's half can't change what the owner consented to (D-16). The `Plugin`'s status records which version actually runs and with which [effective permissions](../glossary.md).

**Owner references and garbage collection.** An object can name its owners in `metadata.ownerReferences`. When an owner is deleted, the cluster's garbage collector deletes what it owns, whether or not the operator is running. The operator makes each `Plugin` the controlling owner of every object it creates for it. controller-runtime also uses these references to route an event on an owned object back to its `Plugin`.

**Status conditions.** A condition is one typed status line: `type`, `status` (`True`, `False` or `Unknown`), a CamelCase `reason`, a readable `message`, the `observedGeneration` it describes and when it last changed. Kubernetes standardized this shape as `metav1.Condition` in 1.19. `kubectl get plugins` shows a `Plugin`'s three conditions as columns.

**Namespace-scoped RBAC and least privilege.** Kubernetes RBAC grants verbs on resources to an identity, through a Role (one namespace) or a ClusterRole (the whole cluster). Least privilege (Saltzer and Schroeder, 1975) gives each part only the rights its job needs. D-15 limits the operator to a Role in `openfief`, so it can't see or change anything elsewhere. Installing CRDs is a cluster-wide act, so a cluster admin does it once. Plugins get even less: a ServiceAccount with no Kubernetes API token at all.

**Server-side apply.** The operator writes objects with server-side apply. It sends only the fields it wants, and the API server records which *field manager* owns each field. A later apply by the same manager with fewer fields removes the ones it dropped. The operator uses this to own its objects outright, and to give each plugin's entry in the shared allow-list ConfigMap its own manager.

**Signed images and OCI referrers.** The [Plugins](plugins.md) page explains Sigstore, keyless signing and the three kinds of trusted signer (D-45). The operator's part is finding and checking signatures. `cosign sign` (v3) writes a Sigstore *bundle*: the signature, the signer's certificate and the transparency-log proof in one document. It stores the bundle in the registry as an OCI *referrer*, an artifact whose `subject` is the image digest, listed by the Referrers API of the OCI distribution spec 1.1 or, where a registry lacks it, under a `sha256-<hex>` tag. A bundle carries its own proofs, so checking it needs the trust root but no call to Sigstore's servers (D-56).

## How it works

### Packages and binaries

| Path | What it holds |
|---|---|
| `internal/plugins/operator/v1alpha1` | The API: `Plugin`, `PluginGrant`, `Manifest`, `Permissions`, `PluginStatus` and the condition names ([types.go](../../internal/plugins/operator/v1alpha1/types.go), [groupversion.go](../../internal/plugins/operator/v1alpha1/groupversion.go)); generated deepcopy code |
| `internal/plugins/operator` | The adapter: `PluginReconciler` ([reconciler.go](../../internal/plugins/operator/reconciler.go)), watches and `GrantReconciler` ([setup.go](../../internal/plugins/operator/setup.go)), status ([status.go](../../internal/plugins/operator/status.go)), resource-to-domain conversion ([convert.go](../../internal/plugins/operator/convert.go)), `BuildObjects` ([objects.go](../../internal/plugins/operator/objects.go)), `ClusterFromServices` ([cluster.go](../../internal/plugins/operator/cluster.go)), `ValidateConfig` ([config.go](../../internal/plugins/operator/config.go)), `ImageVerifier` and `PolicyVerifier` ([verify.go](../../internal/plugins/operator/verify.go)), `SigningKeysPublisher` and `IssuerDiscovery` ([signingkeys.go](../../internal/plugins/operator/signingkeys.go)) |
| `internal/plugins` | The pure domain it calls: `PlanRollout`, `Effective`, `NeedsApproval`, `Compile`, `ParseImageRef`, `EncodeAllowLists`, `TrustPolicy`, `ParseManifestLabel`; the embedded catalog is `internal/plugins/catalog`, empty today |
| `internal/plugins/cosign` | The signature adapter: `Verifier.Inspect` finds an image's bundles and verifies them with sigstore-go, then reads its manifest label ([verifier.go](../../internal/plugins/cosign/verifier.go), [registry.go](../../internal/plugins/cosign/registry.go)); `TrustedRoot` embeds Sigstore's public-good trust root ([trustroot.go](../../internal/plugins/cosign/trustroot.go)). The `openfief` CLI uses the same `Verifier`. `cosign/cosigntest` is test support for code outside the package: an in-memory registry and images signed with a fresh key as `cosign sign --key` stores them ([cosigntest.go](../../internal/plugins/cosign/cosigntest/cosigntest.go)) |
| `internal/plugins/kube` | The plugins context's Kubernetes adapter outside the operator. For the egress proxy: `AllowLists` follows the ConfigMap the operator writes ([allowlists.go](../../internal/plugins/kube/allowlists.go)); `TokenReviewer` checks plugin tokens for one audience, `openfief-egress` in the proxy ([tokenreview.go](../../internal/plugins/kube/tokenreview.go)). For the `openfief` CLI ([instances.go](../../internal/plugins/kube/instances.go)): `Instances` creates, lists and deletes the `Plugin`, `PluginGrant` and Secrets of an instance with the owner's kubeconfig; `OwnerSigners` reads `openfief-trusted-signers` as the operator does; `Resources` and `ResourcesYAML` build the resources an install applies or prints. See [Installing with the CLI](plugins.md#installing-with-the-cli) |
| `cmd/operator` | Wiring: flags, a controller-runtime manager whose cache sees only `openfief`, leader election through a Lease, the trust policy and verifier (`imageVerifier`), the signing keys publisher ([main.go](../../cmd/operator/main.go)); a distroless, non-root image ([Dockerfile](../../cmd/operator/Dockerfile)) |
| `cmd/sigstore-trusted-root` | `task trustroot:update`: refreshes `trusted_root.json` through TUF ([main.go](../../cmd/sigstore-trusted-root/main.go)) |
| `deploy/crds`, `deploy/operator` | Install manifests: the namespace and CRDs; the operator's ServiceAccount, Role, RoleBinding and Deployment |

One reconcile converts the resources into `plugins.Plugin` and `plugins.Grant`, asks `plugins.PlanRollout` which version runs, has `plugins.Compile` turn its effective permissions into a `plugins.Workload`, and only then builds Kubernetes objects with `BuildObjects`.

### The resources and what the operator owns

```mermaid
classDiagram
  direction LR
  class Plugin {
    metadata.name : the instance name
    spec.image : repository@sha256 digest
    spec.manifest : capabilities, egress, secrets, dataTiers, configSchema
    spec.config : JSON object
    status.conditions : Ready, ApprovalRequired, SignatureVerified
    status.deployed : the PluginSpec that runs now
    status.effective : Permissions
    status.pendingApproval : Permissions
  }
  class PluginGrant {
    metadata.name : same as its Plugin
    spec : capabilities, egress, secrets, dataTiers
    status.conditions : Valid
  }
  class ServiceAccount {
    name : plugin-NAME
    automountServiceAccountToken : false
  }
  class PluginConfigMap {
    name : plugin-NAME-config
    config.json : spec.config
  }
  class Service {
    name : plugin-NAME
    type : ClusterIP, port connect
  }
  class NetworkPolicy {
    name : plugin-NAME
    default deny, ingress and egress
  }
  class Deployment {
    name : plugin-NAME
    replicas 1, strategy Recreate
    token openfief-core : always
    token openfief-egress : only with granted egress
  }
  class AllowListConfigMap {
    name : openfief-egress-allowlist
    one key per running plugin
  }
  class Secret {
    created by the owner
  }
  PluginGrant ..> Plugin : approves, paired by name
  Plugin *-- ServiceAccount : owns
  Plugin *-- PluginConfigMap : owns
  Plugin *-- Service : owns
  Plugin *-- NetworkPolicy : owns
  Plugin *-- Deployment : owns
  Plugin ..> AllowListConfigMap : its key, own field manager
  Deployment ..> ServiceAccount : runs as
  Deployment ..> PluginConfigMap : OPENFIEF_CONFIG
  Deployment ..> Secret : mounts read-only if granted
```

The CRD schema rejects bad input before the operator sees it. Names must be lower-case DNS labels of at most 56 characters starting with a letter, so `plugin-<name>` fits a 63-character Service name. `spec.image` must match `^[^@]+@sha256:[0-9a-f]{64}$`. Capabilities are `connector` or `view`; egress lists hold at most 64 [egress hosts](../glossary.md). The operator then validates everything again in the domain, where rules such as "no wildcard hosts" live, because a CRD schema can't express them all.

The operator never creates a Secret (D-29) and its Role can't even read one: it only names the owner's Secrets in the pod spec. The allow-list ConfigMap has no owner reference; each plugin's key is applied by its own field manager, `openfief-operator/allowlist/<name>`.

What `BuildObjects` puts in the pod ([objects.go](../../internal/plugins/operator/objects.go)):

- **Identity.** The ServiceAccount mounts no API token. Every pod gets a projected token with audience `openfief-core` at `/var/run/secrets/openfief/token` (D-27), named by `OPENFIEF_TOKEN_FILE`. Only when an egress host is granted, it gets a second one with audience `openfief-egress` at `/var/run/secrets/openfief-egress/token` (D-63), named by `OPENFIEF_EGRESS_TOKEN_FILE`. Both last 3600 seconds and the kubelet renews them. With separate audiences, the proxy never sees a token the core would accept. That ServiceAccount is the [plugin identity](../glossary.md).
- **Addresses.** `OPENFIEF_CORE_ADDRESS` is the core Service's ClusterIP and port, because plugins have no DNS (D-32). `OPENFIEF_LISTEN_PORT` is the plugin port (8080 by default). Only when an egress host is granted, `HTTPS_PROXY` points at the egress proxy's ClusterIP and `NO_PROXY` at the core.
- **Hardening.** Non-root, read-only root filesystem, every Linux capability dropped, no privilege escalation, seccomp `RuntimeDefault`, no service links. These settings meet the `restricted` Pod Security Standard the namespace enforces.
- **Rollouts.** One replica with the `Recreate` strategy, so an old pod never runs beside a new one with other permissions. A SHA-256 of the config in a pod annotation makes a config change replace the pod.

The NetworkPolicy denies both directions by default. It admits ingress only from the core's pods on the plugin port, allows egress to the core's pods, and allows egress to the proxy's pods only when an egress host is granted. It never allows DNS. The peers are selected by the core's and the proxy's Service selectors. `ClusterFromServices` refuses a Service with no selector, a selector that would also match plugin pods (`ErrSelectorOverlap`, which would let a plugin pass for the core), or a named `targetPort`, because NetworkPolicy matches pod port numbers.

### One reconcile pass

```mermaid
flowchart TD
  start(["Reconcile openfief/NAME"]) --> exists{"Plugin exists?"}
  exists -- "no, deleted" --> drop["Remove NAME's allow-list key.<br/>Garbage collection removes the rest."]
  exists -- yes --> check["pluginFromSpec and ValidateConfig check the current version.<br/>verify: ParseImageRef, then signature and manifest label"]
  check --> grant{"PluginGrant NAME?"}
  grant -- "none" --> nogrant["NoGrant: everything requested is pending"]
  grant -- "invalid" --> badgrant["InvalidGrant: approves nothing"]
  grant -- "valid" --> plan{"plugins.PlanRollout"}
  plan -- "current deployable,<br/>nothing pending" --> eff{"Effective: any capability granted?"}
  plan -- "current waits or is undeployable,<br/>status.deployed has a version" --> eff
  plan -- "no version to run:<br/>AwaitingApproval or NotDeployable" --> teardown
  eff -- "no: NothingGranted" --> teardown
  eff -- "yes, previous version" --> prev{"previous image still verifies?"}
  prev -- no --> teardown
  prev -- yes --> build
  eff -- "yes, current version" --> build["Read core and proxy Services,<br/>Compile, BuildObjects"]
  build -- "unusable: ClusterUnavailable" --> teardown
  build --> apply["Server-side apply, owned by the Plugin:<br/>ServiceAccount, ConfigMap, NetworkPolicy, Service, then Deployment"]
  apply --> publish["Set NAME's allow-list key to the effective egress hosts"]
  nogrant --> teardown
  badgrant --> teardown
  teardown["Teardown: NetworkPolicy becomes deny-all, then delete<br/>Deployment, Service, ConfigMap, ServiceAccount"] --> unpublish["Remove NAME's allow-list key"]
  publish --> status["Write status if it changed.<br/>Requeue in 1 minute if a check was Unknown"]
  unpublish --> status
```

"Deployable" means the spec is valid, its config passes the manifest's JSON Schema, and the image check below returned `True`. `ValidateConfig` treats a missing config as `{}` and gives the schema a loader that refuses every `$ref` outside the schema itself, so a schema can't make the operator read a file or reach the network.

Three cases are worth tracing:

- **No grant.** `PlanRollout` returns `RolloutNoGrant` and puts everything the manifest requests into [pending approval](../glossary.md). The operator tears down, which leaves one owned object: a deny-all NetworkPolicy. Status: `Ready=False/NoGrant`, `ApprovalRequired=True/PermissionsPending`. Deleting a grant from a running plugin takes this same path, so revocation is immediate.
- **An upgrade that requests more.** The new spec asks for an egress host the grant lacks. `NeedsApproval` returns exactly that host. Because `status.deployed` still holds the running version, the [rollout](../glossary.md) keeps it (`RolloutPrevious`), limited to what the *current* grant approves, so a narrowed grant takes effect at once. The old image must still verify. Status: `pendingApproval` lists the host and `ApprovalRequired=True`; `Ready` becomes `True/RunningPreviousVersion` once the Deployment is available. With no earlier version, nothing runs (`AwaitingApproval`).
- **A tag-referenced image.** `repo:1.0` alone fails the CRD pattern, so the API server refuses the write and the operator never sees it. `repo:1.0@sha256:…` passes the pattern, but `plugins.ParseImageRef` refuses a tag next to a digest. The current version is then invalid: `SignatureVerified=False/InvalidImage` and `ApprovalRequired=Unknown/InvalidSpec`. As with an upgrade, an earlier verified version keeps running if there is one; otherwise nothing runs and `Ready=False/InvalidSpec`.

`Ready`'s reason gives the first cause: `Running`, `RunningPreviousVersion`, `Progressing` (Deployment not yet available), `NoGrant`, `InvalidGrant`, `InvalidPreviousVersion`, `ClusterUnavailable`, `AwaitingApproval`, `NothingGranted`, `InvalidSpec`, `InvalidConfig` or `SignatureNotVerified` ([status.go](../../internal/plugins/operator/status.go)). Messages are capped at 1024 bytes. `GrantReconciler` separately sets each `PluginGrant`'s `Valid` condition.

### The signature and manifest check

```mermaid
flowchart TD
  call(["verify: image and declared manifest"]) --> parse{"ParseImageRef"}
  parse -- "fails" --> invalid["False: InvalidImage"]
  parse -- ok --> cached{"signed image cached by digest?"}
  cached -- yes --> decide
  cached -- no --> refs["cosign.Verifier.Inspect: list referrers,<br/>read up to 32 Sigstore bundles"]
  refs -- "registry error" --> unknown["Unknown: RegistryUnavailable,<br/>reconcile again in 1 minute"]
  refs --> each["sigstore-go checks each bundle: the image digest, the cosign sign predicate;<br/>keyless: Fulcio chain, SCT, Rekor entry, timestamp, signer identity;<br/>key: an owner's or publisher's public key"]
  each -- "none verifies" --> decide
  each -- "some verify" --> label["Read the io.openfief.plugin.manifest label;<br/>every platform image must carry the same one"]
  label -- "missing or invalid" --> badlabel["False: InvalidManifestLabel"]
  label --> keep["cache the signed image"] --> decide{"TrustPolicy.Decide"}
  decide -- "no signature" --> unsigned["False: Unsigned"]
  decide -- "no trusted signer" --> untrusted["False: UntrustedSigner"]
  decide -- "label is not spec.manifest" --> mismatch["False: ManifestMismatch"]
  decide -- "catalog pins the digest,<br/>its publisher signed" --> verified["True: Verified"]
  decide -- "another trusted signer" --> signed["True: SignedByTrustedSigner"]
```

`PolicyVerifier` ([verify.go](../../internal/plugins/operator/verify.go)) turns the outcome into the `SignatureVerified` condition, so the status tells a [verified plugin](../glossary.md) from one only signed by a trusted signer (D-56). It runs for the current spec with its manifest, and again for the previous version with that version's manifest when the previous version would keep running. The manifest label is compared as sets with `spec.manifest` (`Manifest.Equal`), and it is read only from an image a trusted signer signed (D-42). Its signature covers the image config, and so the label.

`cmd/operator` builds the trust policy once at start ([main.go](../../cmd/operator/main.go)). It holds OpenFief's CI (keyless, GitHub Actions, the `images.yml` workflow on `main`), each publisher in the embedded catalog, and the owner's signers from `signers.json` in the optional ConfigMap `openfief-trusted-signers`, mounted at `/etc/openfief/trust`. A missing file adds no signer; an invalid one stops the operator; a change needs a restart. No flag skips the check (D-57). The registry is read anonymously, so plugin images must be public, and the operator needs HTTPS egress to every registry plugins come from and the blob storage it redirects to.

A signed image is cached by digest, at most 256 and emptied when full, so a reconcile doesn't refetch it and a running plugin stays verified through a registry outage while the operator keeps running. An unsigned or unreadable image is never cached. `UnimplementedVerifier` (`Unknown: NotImplemented`) is still there for tests; the binary always wires `PolicyVerifier`.

### What triggers a reconcile

`SetupWithManager` ([setup.go](../../internal/plugins/operator/setup.go)) runs `PluginReconciler` on changes to a `Plugin` and to the five objects it owns, so a hand edit to one of them is followed by a re-apply of the operator's fields. A `PluginGrant` maps to the `Plugin` of the same name. A change to the core's or the proxy's Service, or to the allow-list ConfigMap, reconciles every `Plugin`. Requests outside `openfief` are ignored.

### The egress allow-list hand-off

The operator compiles each running plugin's effective egress hosts into the [egress allow-list](../glossary.md) and writes them to the `openfief-egress-allowlist` ConfigMap (D-32). A running plugin with no egress gets an empty value, meaning "none"; a plugin that doesn't run has no key. On the proxy's side, `kube.WatchAllowLists` follows that one ConfigMap with an informer and swaps the decoded lists in atomically. It fails closed: while the ConfigMap is missing or doesn't decode, no plugin has a list. `kube.TokenReviewer` asks the TokenReview API whether a token is valid for the audience it was built with and returns its username. `cmd/egress-proxy` builds it for `openfief-egress`, so a plugin's core token is refused there, and wraps it in `app.NewCachingTokenReviewer`, which caches accepted reviews for at most five minutes and never past the token's expiry. The proxy's own behavior is documented with the egress proxy.

### Publishing the signing keys for plugins

Only the core may call a plugin, and it proves that with a ServiceAccount token whose audience is the plugin instance (D-59). The SDK checks that token's signature itself, because a plugin can't reach the Kubernetes API for a TokenReview. It needs the cluster's public signing keys and its token issuer, and the operator publishes them.

`SigningKeysPublisher` ([signingkeys.go](../../internal/plugins/operator/signingkeys.go)) reads them with `IssuerDiscovery` from the API server's ServiceAccount issuer discovery endpoints: the issuer from `/.well-known/openid-configuration` and the keys from `/openid/v1/jwks`. It reads the keys from the API server itself, never from the `jwks_uri` the configuration names, which may be an address the operator can't reach. `plugins.ParseSigningKeys` accepts only RSA and EC public keys with key IDs, and the publisher server-side applies them as the ConfigMap `openfief-sa-jwks` (keys `jwks.json` and `issuer`), with its own field manager, `openfief-operator/signing-keys`. It runs once at start, every `--signing-keys-refresh` (ten minutes by default) so a rotated key reaches plugins, and on every change to that ConfigMap, so an edited or deleted one is put back. If the keys can't be read, the published ones stay.

Every plugin pod mounts that ConfigMap read-only at `/var/run/secrets/openfief-sa-jwks`, and its environment names the instance (`OPENFIEF_INSTANCE`), the core's username (`OPENFIEF_CORE_SUBJECT`, from `--core-service-account`, `system:serviceaccount:openfief:openfief-core` by default) and the mount (`OPENFIEF_SA_JWKS_DIR`). The mount isn't optional: until the ConfigMap exists, a plugin pod doesn't start.

The discovery endpoints are non-resource URLs, which no namespaced Role can grant. Kubernetes' default RBAC binds the ClusterRole `system:service-account-issuer-discovery`, which allows reading both, to the group `system:serviceaccounts`, so the operator reads them as any ServiceAccount does and keeps its Role-only footprint (D-15). A cluster that removed that binding has to bind the ClusterRole to `openfief/openfief-operator`; until then plugins get no keys.

### What runs in the openfief namespace

```mermaid
flowchart LR
  api[("Kubernetes API server")]
  ext(["External hosts"])
  reg(["Container registries"])
  subgraph ns["namespace openfief, Pod Security restricted"]
    op["openfief-operator<br/>Role in openfief only"]
    ts[/"ConfigMap openfief-trusted-signers, optional"/]
    al[/"ConfigMap openfief-egress-allowlist"/]
    jwks[/"ConfigMap openfief-sa-jwks"/]
    proxy["egress-proxy<br/>Service port 4750"]
    core["core, not built yet<br/>expected Service openfief-core, port connect"]
    subgraph one["per Plugin NAME"]
      np{{"NetworkPolicy plugin-NAME"}}
      pod["plugin pod<br/>ServiceAccount plugin-NAME, no API token"]
    end
  end
  op -- "watch Plugin, PluginGrant; apply objects" --> api
  op -- "one key per running plugin" --> al
  op -- "issuer and public keys from issuer discovery" --> jwks
  jwks -. "mounted read-only" .-> pod
  ts -. "mounted, read at start" .-> op
  op -- "referrers, bundles, image config; anonymous HTTPS" --> reg
  proxy -- "watch, Role limited to this ConfigMap" --> al
  proxy -- "TokenReview, ClusterRole create only" --> api
  core -- "Connect RPC on the plugin port" --> pod
  pod -- "Connect RPC to the core" --> core
  pod -- "HTTP CONNECT with the openfief-egress token,<br/>only with granted egress" --> proxy
  proxy -- "allow-listed hostnames only" --> ext
  np -. "selects" .-> pod
  pod -- "DNS, API server, other pods, direct internet" --x denied["denied"]
```

Only plugin pods get NetworkPolicies from the operator; the install manifests give none to the operator or the proxy. The core has no binary or manifest yet. The operator's flags name its Service (`--core-service=openfief-core`, `--core-port=connect`), and until that Service exists, any plugin that would run gets `ClusterUnavailable`.

Install order ([deploy/crds/kustomization.yaml](../../deploy/crds/kustomization.yaml), [deploy/operator/kustomization.yaml](../../deploy/operator/kustomization.yaml)): a cluster admin runs `kubectl apply -k deploy/crds`, which creates the `openfief` namespace with `pod-security.kubernetes.io/enforce: restricted` and both CRDs. Then `kubectl apply -k deploy/operator` installs the operator, after an overlay pins its image by digest. Its [role.yaml](../../deploy/operator/role.yaml) is generated from the `+kubebuilder:rbac` markers in [reconciler.go](../../internal/plugins/operator/reconciler.go): read `plugins` and `plugingrants`; update their status; update `plugins/finalizers`, which owner references with `blockOwnerDeletion` need; full control of ServiceAccounts, ConfigMaps, Services, Deployments and NetworkPolicies; create and patch events; manage Leases for leader election. All of it is limited to `openfief`, but within the namespace it isn't limited to the operator's own objects: Kubernetes RBAC can't restrict `create` by name.

## Decisions

- **D-10**: Kubernetes-oriented; the operator is a controller-runtime manager.
- **D-11**: manifest compiled into NetworkPolicy, Secret mounts, RBAC and allow-list. The config is also checked against the manifest's JSON Schema.
- **D-13**: an operator manages plugins through a `Plugin` resource.
- **D-15**: Role in `openfief` only; CRDs installed by a cluster admin.
- **D-16**: requested and approved permissions split; only both enforced; upgrades wait.
- **D-17**: images pinned by digest and verified with cosign bundles. Digest pinning is in the CRD pattern and `ParseImageRef`; signatures are checked with sigstore-go.
- **D-27**: projected token with audience `openfief-core` mounted in every plugin.
- **D-28**: no Ingress; plugins get ClusterIP Services only.
- **D-29**: Secrets are the owner's; the operator only mounts them.
- **D-32**: allow-list, NetworkPolicy to the core and proxy only, no DNS, `HTTPS_PROXY` by ClusterIP. **Difference:** D-32 says the operator compiles each grant's hostnames; the code publishes the effective hosts, the intersection of manifest and grant. That's narrower and matches D-16.
- **D-42**: manifest in a signed image label, matched against `spec.manifest`.
- **D-43, D-44**: embedded catalog; a pinned digest signed by its publisher is `Verified`. The catalog is empty today.
- **D-45**: three kinds of trusted signer; the owner's come from a ConfigMap.
- **D-48**: each `Plugin` name is an instance with its own ServiceAccount, config, grant and allow-list key. **Gap:** Secret names come from the manifest, and D-42 now makes `spec.manifest` equal the image's label. So two instances of one image always request the same Secrets, and can't have their own as D-48 says.
- **D-55**: `Plugin` and `PluginGrant` are the only custom resources; no money.
- **D-56**: the operator verifies every image itself against a bundled trust root, and the status says `Verified` or `SignedByTrustedSigner`. "Offline" means no Sigstore service; the image's registry must still be reachable, at least once per digest while the operator runs. No UI shows the status yet.
- **D-57**: nothing deploys without a `True` verification; no flag or setting overrides it.
- **D-58**: **partly.** Deleting a `Plugin` removes its workload (by garbage collection) and its allow-list key. The operator never deletes a `PluginGrant` or a Secret; `openfief plugin uninstall` deletes them ([Plugins](plugins.md#installing-with-the-cli)). Nothing marks the ledger data of a removed instance yet.
- **D-59**: the NetworkPolicy admits ingress only from the core, and the operator publishes the cluster's token-signing public keys in `openfief-sa-jwks`, with which the SDK verifies the core's per-instance TokenRequest tokens.
- **D-63**: separate `openfief-egress` token, checked by TokenReview and cached. The operator projects it only for plugins with granted egress; `kube.AllowLists` swaps allow-lists atomically and fails closed.

## Verification

| Behavior | Tests | kibi |
|---|---|---|
| No grant: nothing runs, everything pending, deny-all NetworkPolicy | `TestReconcile/plugin_without_a_grant_runs_nothing` in [reconciler_test.go](../../internal/plugins/operator/reconciler_test.go); `TestPlanRollout` in [rollout_test.go](../../internal/plugins/rollout_test.go) | REQ-operator-runs-only-granted, REQ-plugins-rollout-no-grant |
| Upgrade requesting more keeps the old version; approval rolls out the new one | `TestReconcile/granted_plugin_runs,_and_an_upgrade_waits_for_approval`; `TestPlanRollout` | REQ-operator-upgrade-status, REQ-plugins-rollout-upgrade-waits |
| Deleting the grant tears the plugin down | `TestReconcile/revoking_the_grant_removes_the_plugin` | REQ-operator-runs-only-granted |
| A signed image with a matching label runs as `SignedByTrustedSigner`; a different label is refused | `TestReconcile/signature_checks_run_through_the_trust_policy` | REQ-plugins-manifest-label-match, REQ-plugins-trusted-signers-only |
| An image not verified is never deployed, and an `Unknown` check is retried | `TestReconcile/unverified_image_is_never_deployed`, `TestReconcile/an_image_that_could_not_be_checked_is_checked_again` | REQ-operator-signature-required |
| Outcomes map to condition reasons; signed images are cached, unsigned ones never | `TestPolicyVerifier`, `TestPolicyVerifierCache`, `TestUnimplementedVerifier` in [verify_test.go](../../internal/plugins/operator/verify_test.go) | REQ-plugins-verified-status |
| Bundles verified: Referrers API and tag, image indexes, owner keys and patterns; refused: forks, wrong issuer, wrong digest, attestations, expired or foreign certificates, missing SCT, missing or tampered labels | `TestInspect` (20 cases), `TestProductionVerifierRequiresSCT`, `TestInspectUnreachableRegistry`, `TestTrustedRoot` in [verifier_test.go](../../internal/plugins/cosign/verifier_test.go) | REQ-plugins-cosign-bundle-verification, REQ-plugins-bundled-trust-root |
| Trust decisions and the operator's trust configuration | `TestTrustPolicyDecide` in [trust_test.go](../../internal/plugins/trust_test.go); `TestTrustPolicy`, `TestImageVerifier`, `TestParseFlagsTrustedSigners` in [main_test.go](../../cmd/operator/main_test.go) | REQ-plugins-trust-decision, REQ-operator-trust-config |
| Config checked against its schema, which loads nothing | `TestValidateConfig` in [config_test.go](../../internal/plugins/operator/config_test.go); `TestReconcile/config_its_schema_rejects_is_never_deployed` | REQ-operator-config-schema |
| Invalid grant marked and enforced as no grant | `TestReconcile/invalid_grant_is_marked_and_enforced_as_no_grant` | REQ-operator-grant-validity |
| Allow-list key set while running, removed after deletion | the upgrade subtest; `TestDeletedPluginLeavesTheAllowList`; `TestAllowListConfigMap` | REQ-operator-allow-list-publish |
| Exact objects for a workload | `TestBuildObjects`, `TestDenyAllNetworkPolicy` in [objects_test.go](../../internal/plugins/operator/objects_test.go) | REQ-operator-workload-objects |
| Peers only from usable Services | `TestClusterFromServices` in [cluster_test.go](../../internal/plugins/operator/cluster_test.go), `TestBuildObjectsRefuses` | REQ-operator-cluster-peers |
| Image references by digest only | `TestParseImageRef` in [image_test.go](../../internal/plugins/image_test.go) | REQ-plugins-image-digest-only |
| No ClusterRole for the operator; no Ingress in the install bases `crds`, `operator`, `egress-proxy`, `core` (`deploy/demo` is excluded) | `TestManifests` in [manifests_test.go](../../internal/plugins/operator/manifests_test.go) | TEST-operator-manifests (D-15, D-28) |
| Egress token only with granted egress; core token always | `TestBuildObjects`; `TestReconcile/granted_plugin_without_egress_gets_only_the_core_token` and the upgrade subtest, which read both tokens from the Deployment | REQ-plugins-egress-token-projected |
| Proxy side: `openfief-egress` only (a core-only token is refused), ConfigMap followed fail-closed | `TestTokenReviewerReviewToken`, `TestWatchAllowLists*` in [kube_test.go](../../internal/plugins/kube/kube_test.go) | REQ-plugins-egress-proxy-identifies-plugin, REQ-plugins-egress-allow-lists-fail-closed |
| Accepted reviews cached, bounded, refused ones never | `TestCachingTokenReviewer*` in [tokencache_test.go](../../internal/plugins/app/tokencache_test.go) | REQ-plugins-egress-token-review-cache |
| The operator's ServiceAccount reads issuer discovery with default RBAC; the keys are published, restored after deletion and kept when the source fails; the core's TokenRequest tokens verify with the SDK from the ConfigMap alone | `TestSigningKeys`, `TestSigningKeysPublisherRuns` in [signingkeys_test.go](../../internal/plugins/operator/signingkeys_test.go) | REQ-plugins-signing-keys-published, REQ-sdk-go-jwks-verifier |
| Every plugin pod mounts the signing keys and gets its instance name and the core's username | `TestBuildObjects` | REQ-plugins-workload-core-auth |
| Flags, API types, controller setup | `TestParseFlags` in [main_test.go](../../cmd/operator/main_test.go); `TestAddToScheme`, `TestDeepCopyIsIndependent`; `TestSetupWithManager` | TEST-operator-cmd, TEST-operator-api |

The reconciler tests use envtest: a real `kube-apiserver` and `etcd` with the CRDs from `deploy/crds`, and no other controllers. Run everything with `task test`, which installs envtest 1.36.2 into `.bin/envtest` and sets `KUBEBUILDER_ASSETS`. A plain `go test ./internal/plugins/operator/...` fails those tests fast without it. `task gen:check` fails if the CRDs, `role.yaml` or the deepcopy code differ from what `task gen` generates from the Go markers.

What the tests don't show:

- envtest has no Deployment controller and no garbage collector. `Ready` never reaches `Running` or `RunningPreviousVersion` in a test, and the cascade on `Plugin` deletion is untested.
- The CRD's image pattern is not exercised by a test; only `ParseImageRef` is.
- From reading the code: if the operator is down when a `Plugin` is deleted, nothing removes that plugin's allow-list key afterwards, because no event arrives for an object that no longer exists. A reinstall under the same name overwrites the key.
- From server-side apply's rules, not a test: a re-apply restores the fields the operator owns, but keeps fields another field manager added, such as an extra container, volume or environment variable added with `kubectl edit`. NetworkPolicy rule lists are atomic, so they are restored whole.
- The cosign tests run against an in-process registry and sigstore-go's virtual Sigstore. No test reads a real registry such as GHCR or a bundle from the public-good instance; `TestTrustedRoot` only checks that the bundled root loads.
- From reading the code: the signed-image cache lives in memory. If the operator restarts while a plugin's registry is down, the running version's check is `Unknown`, so the plugin is torn down and comes back on a retry after the registry does.
- There is no end-to-end suite (`test/e2e` doesn't exist), so no test checks that a real CNI enforces the NetworkPolicies.

## Not built yet

- **D-43**: the catalog lists no plugin, so the `openfief` CLI's preview of the signature check (`cosign.Verifier` and the same trust policy) never shows a verified plugin yet.
- **D-56**: the UI that shows the verification status.
- Private registries: the operator and the CLI read registries anonymously.
- **D-58**: marking the ledger data of a removed instance.
- **D-12**: end-to-end tests on kind or k3d.
- The core's Deployment and Service, which the operator expects as `openfief-core`.

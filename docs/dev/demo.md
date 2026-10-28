# Demo environment

**(P)** This page is a Claude Code proposal, written for the author's request for committed automation that makes it easy to create and recreate demo machines within about €10 a month. The rules it serves are in [sdlc.md](sdlc.md#ci-and-environments): a Hetzner machine running k3s, deployed from `main` and from beta releases, holding synthetic data only (GR-12). Commands live in `Taskfile.yml`.

A demo is one Hetzner Cloud server running k3s. cloud-init installs pinned k3s and Flux, and Flux pulls everything else from `deploy/demo` in this repository. Nothing holds state outside Hetzner and git: every resource carries the labels `app=openfief` and `env=<name>`, and the scripts find resources by those labels. So any session with the token can create, inspect or destroy a demo, and `up` never creates something twice.

## Before the first demo

1. **API token.** In the Hetzner Cloud console, open the project, then Security > API tokens, and create a Read & Write token. Export it only in the shell that runs the commands: `read -rs HCLOUD_TOKEN && export HCLOUD_TOKEN`. The scripts never print it or write it to disk, and they ignore any `hcloud` CLI context you have stored.
2. **SSH key.** Upload your public key with the label the scripts look for: `hcloud ssh-key create --name <you> --public-key-from-file ~/.ssh/id_ed25519.pub --label openfief-ssh=true`, or label an existing key with `hcloud ssh-key add-label <key> openfief-ssh=true`. Every key with that label is installed on new demos; without one, `up` refuses.
3. **Where you connect from.** Set `DEMO_SSH_CIDRS` to your address, such as `203.0.113.7/32`; several CIDRs are separated by commas. With no value, the firewall allows no inbound traffic at all, not even SSH.
4. **Hard spending cap (optional).** Hetzner Cloud has no spending-limit API, so the budget check below lives in the scripts. For a hard cap, ask Hetzner support to lower the project's server limit to 1: no script or token can then create a second server.

## Commands

| Command | What it does |
|---|---|
| `task demo:plan ENV=demo` | Lists what `up` would create and the monthly cost against the budget. Changes nothing. |
| `task demo:up ENV=demo` | Creates the firewall and the server if they're missing, and updates the firewall rules from `DEMO_SSH_CIDRS`. Refuses to go over the budget. |
| `task demo:status ENV=demo` | Shows the server, its IPv4, the firewall and the project's monthly cost. |
| `task demo:down ENV=demo CONFIRM=demo` | Deletes every resource labelled `app=openfief,env=demo`, server first. |
| `task demo:recreate ENV=demo CONFIRM=demo` | `down`, then `up`: a fresh machine from the current cloud-init and pins. |

`ENV` names the demo: `task demo:up ENV=try-beta` makes a second, separate one (`openfief-try-beta`), if the budget allows it. Settings come from the shell or the command line:

| Setting | Default | Meaning |
|---|---|---|
| `HCLOUD_TOKEN` | none (required) | The project's API token. |
| `DEMO_SSH_CIDRS` | empty: no inbound traffic | Sources allowed to reach SSH (22/tcp). |
| `DEMO_BUDGET_EUR` | `10` | Maximum monthly gross cost of everything labelled `app=openfief` in the project. |
| `DEMO_SERVER_TYPE` | `cax11` | Hetzner server type; `cax*` types are Arm64. |
| `DEMO_LOCATION` | `nbg1` | Hetzner location. |
| `DEMO_IMAGE` | `ubuntu-24.04` | Server image. |
| `DEMO_GIT_BRANCH` | `main` | Branch Flux follows; another branch suits a throwaway demo of a change. |
| `DEMO_GIT_URL` | `https://github.com/arkadiuss/openfief` | Public repository Flux pulls from, with no credentials. |

## Size and cost

Before it creates anything, `up` fetches the live price list, adds up the monthly gross price of every resource labelled `app=openfief` in the project (servers with their IPv4 and backups, unattached primary IPs, floating IPs, volumes, snapshots and load balancers), adds what it would create, and refuses if the total is over `DEMO_BUDGET_EUR`. It prints each line, so you can check the sum. Prices are rounded up, never down.

One demo is a CAX11: 2 Arm64 vCPUs, 4 GB of RAM and 40 GB of disk, with one primary IPv4. It needs an IPv4 because GitHub and GHCR, which Flux and k3s pull from, have no IPv6. There's no floating IP, volume or load balancer. Hetzner's list prices on 2026-10-28, excluding VAT:

| Item | Net per month | Gross at 19% VAT | Gross at 23% VAT |
|---|---|---|---|
| CAX11 server | €6.49 | €7.72 | €7.98 |
| Primary IPv4 (about €0.50) | €0.50 | €0.60 | €0.62 |
| **Total** | **€6.99** | **€8.32** | **€8.60** |

The API reports gross prices at the account's VAT rate, and the scripts use those, so the check matches the invoice. Servers are billed by the hour up to the monthly price, so a deleted demo stops costing money. A CX23 (x86, 4 GB) is €0.50 cheaper and works the same way (`DEMO_SERVER_TYPE=cx23`). A CAX21 (8 GB, €10.99 net) is already over the budget alone.

Why 4 GB is enough: Ubuntu plus k3s (API server, SQLite datastore, kubelet, flannel, CoreDNS, local-path, metrics-server) uses about 1 GB; Flux's controllers about 0.4 GB; the CloudNativePG operator and PostgreSQL are capped at 256 MB and 512 MB in `deploy/demo`; and the OpenFief operator and egress proxy are capped at 256 MB each. That is about 2.5 GB at the limits, which leaves room for the core and a few plugins at 64 to 256 MB each.

## What the machine runs

`deploy/demo/cloud-init.yaml` is a template. `up` fills in the versions and sha256 checksums pinned in `Taskfile.yml`, and refuses to send it while any value is missing or malformed. It contains no secrets. On first boot it:

- allows SSH with keys only: no passwords, a locked root password, and root login by key only;
- turns on unattended security upgrades, rebooting at 04:00 UTC when an upgrade needs it;
- installs the pinned k3s release binary and install script, each checked against its pinned sha256, with Traefik and servicelb disabled (D-28: no ingress), flannel and k3s's NetworkPolicy controller on (D-12), local-path storage, and Secrets encrypted at rest in the datastore (D-29);
- places Flux's pinned `install.yaml` (checked against its sha256), plus a `GitRepository` and `Kustomization` for `./deploy/demo`, in k3s's auto-deploy manifests directory, so even the bootstrap is declared rather than applied by hand.

Flux then applies `deploy/demo/stack.yaml` in dependency order: the namespace and CRDs (`deploy/crds`), the CloudNativePG operator from a pinned Helm chart, a one-instance PostgreSQL cluster whose password CloudNativePG generates in the cluster, and finally the OpenFief operator and egress proxy from the signed artifact CI publishes, with images pinned by digest (D-17; see [Deploys](#deploys)).

The Hetzner disk isn't encrypted by us, so a demo must only ever hold synthetic data: never real broker credentials, account numbers or portfolio data (GR-12).

## Reach the demo

There is no ingress (D-28), and the firewall exposes nothing but SSH from `DEMO_SSH_CIDRS`. The simplest way in is kubectl on the machine:

```bash
ip="$(hcloud server ip openfief-demo)"
ssh root@"$ip" k3s kubectl get kustomizations -n flux-system   # is Flux in sync?
ssh root@"$ip" k3s kubectl get pods -A
```

To use your own kubectl or a browser, tunnel the API server and port-forward a service:

```bash
ssh -N -L 6443:127.0.0.1:6443 root@"$ip" &
ssh root@"$ip" cat /etc/rancher/k3s/k3s.yaml >~/.kube/openfief-demo.yaml   # cluster-admin: keep it local, never commit it
KUBECONFIG=~/.kube/openfief-demo.yaml kubectl -n openfief port-forward svc/<service> 8080:<port>
```

The tunnelled kubeconfig points at `127.0.0.1`, which `task e2e` takes for a local cluster. Keep it in its own file and never run the end-to-end tests with it.

## Deploys

Deploys are automatic GitOps pulls: a merge to `main` reaches the demo with no manual step, and nothing outside the machine holds a credential for the cluster. Flux follows two sources:

- **git, for the platform.** The namespace and CRDs, the CloudNativePG operator and the database follow `main` directly (`deploy/crds`, `deploy/demo/cnpg`, `deploy/demo/database`).
- **A signed OCI artifact, for OpenFief itself.** After CI pushes and signs the images on `main`, or a beta release does, the `deploy demo` job runs `task deploy:demo`. It verifies each image in `deploy/demo/openfief/kustomization.yaml` with cosign, against the keyless identity of this repository's `images.yml` workflow on `main`, and writes the digest that verified. It then renders the directory, pushes it to `ghcr.io/<owner>/<repo>/demo-manifests` tagged with the commit or the version, signs it keylessly, and moves the `demo` tag to it. Flux checks the artifact's signature against `deploy-demo.yml` on `main` before it applies anything, and polls every two minutes.

The committed digests in `deploy/demo/openfief` stay placeholders (all zeros): CI fills them in on its own checkout every time, so no deploy commit is ever needed. Until the first artifact exists, `demo-openfief` waits for its source while the other layers run. `task demo:validate` (part of `task check`) builds every layer and fails if an image isn't pinned by digest or an OCI source doesn't require a cosign signature.

To enable the job, the owner:

1. creates the GitHub Environment `demo` (Settings > Environments) with deployment branches limited to `main` and a required reviewer (see [From the pipeline](#from-the-pipeline));
2. creates the machine, with `task demo:up ENV=demo` or the Demo infrastructure workflow;
3. after the first push to `main`, makes the GHCR packages `operator`, `egress-proxy` and `demo-manifests` public (the package's settings > Change visibility), so the cluster pulls them without credentials;
4. sets the repository variable `DEMO_ENABLED=true`.

Every push to `main` and every beta then deploys on its own. A required reviewer on `demo` also gates these deploy jobs; leave it off if deploys should never wait for approval, since the deploy job can't touch the machine itself.

## From the pipeline

`.github/workflows/demo-infra.yml` (Actions > Demo infrastructure > Run workflow) runs the same targets as a terminal: `task demo:<action> ENV=<env>`, with the same budget check. It's started by hand only, one run per demo name at a time, and it runs in the GitHub Environment `demo`:

- add the environment secret `HCLOUD_TOKEN` (the project's API token);
- add a required reviewer, so every run that can spend money or delete a demo waits for approval;
- optionally set the environment variables `DEMO_SSH_CIDRS`, `DEMO_BUDGET_EUR`, `DEMO_SERVER_TYPE` and `DEMO_LOCATION`.

For `down` and `recreate`, type the demo's name again in the `confirm` input; it's passed on as `CONFIRM`.

## Recreate and destroy

`task demo:recreate ENV=demo CONFIRM=demo` replaces the machine. Everything in the cluster comes back from git, but the database starts empty, because a demo holds only synthetic data. `up` doesn't change an existing server's cloud-init, so a new k3s or Flux pin takes effect only on recreate. `task demo:down ENV=demo CONFIRM=demo` removes the demo and stops its cost. Both refuse without `CONFIRM` set to the demo's name.

## Update the pinned versions

k3s and Flux versions sit in `Taskfile.yml` with their checksums. Renovate proposes the version bumps but can't compute the checksums, so a bump needs them too, from the release's own files:

- k3s: `K3S_SHA256_ARM64` and `K3S_SHA256_AMD64` from `sha256sum-arm64.txt` and `sha256sum-amd64.txt` of the release (the lines for `k3s-arm64` and `k3s`), and `K3S_INSTALL_SHA256` as the sha256 of `https://raw.githubusercontent.com/k3s-io/k3s/<version>/install.sh` (with `+` written as `%2B`);
- Flux: `FLUX_INSTALL_SHA256` as the sha256 of the release's `install.yaml`.

A wrong checksum stops the machine's bootstrap at that download, so it fails safe. The CloudNativePG chart version is pinned in `deploy/demo/cnpg/helmrelease.yaml`.

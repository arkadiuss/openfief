---
title: release-bundle.sh pins every image of the bootstrap install to a given digest, deterministically
status: active
sourceFile: scripts/tests/release-bundle.test.sh
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-01T13:50:20.472Z'
id: TEST-tooling-release-bundle
type: test
---
Runs scripts/release-bundle.sh against the real deploy/ and a throwaway one, with synthetic digests. The install file pins exactly the operator, egress proxy and core images to the given repository and digest, from arguments matched by name or from images.yml's digest files under a lower-cased prefix (ignoring plugin and demo-seed files), or to all-zeros placeholders for a dry run; the CRDs file holds only the namespace and the CRDs; the same images in another order give byte-identical files. It refuses a missing image or digest file, a digest that is not sha256, an image by tag, two images for one name, an image the install doesn't name, digest files without a prefix or together with arguments, placeholders with images or with VERIFY, VERIFY without REPO, a version that can't name a file and a base that names its image by tag, and a failed run writes nothing. Signature verification (VERIFY=1) needs cosign and keyless signatures, so it runs only in the release workflow.
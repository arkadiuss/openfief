---
title: The install use cases verify with the operator's trust policy, refuse what it refuses and uninstall without touching shared Secrets
status: active
sourceFile: internal/plugins/app/install_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T22:43:21.690Z'
id: TEST-plugins-app-install
type: test
---
Tests in package app_test drive Installer.Prepare with a fake ImageInspector that answers per synthetic repository (publisher-signed, CI-signed, stranger-signed, unsigned, unreachable) and a real TrustPolicy with one catalog publisher. They check the verified catalog plugin, an unverified CI-signed image named after its repository, an owner-named instance, and refusals for unsigned, untrusted, unreachable, schema-rejected, unnamable, tagged and unknown arguments. An in-memory InstanceStore then checks that installing over an existing name fails, that Instances lists by name, and that uninstalling two instances of one image keeps their shared Secret until the last one goes. A false pass would be an unsigned image prepared; its case asserts both ErrUnsigned and ErrNotTrusted.
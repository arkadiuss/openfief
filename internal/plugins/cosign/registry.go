package cosign

import (
	"fmt"
	"io"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/sigstore/sigstore-go/pkg/bundle"

	"github.com/arkadiuss/openfief/internal/plugins"
)

// Bounds on what one image may make the operator read.
const (
	// maxReferrers is how many referrers of an image are read for bundles.
	maxReferrers = 32
	// maxBundleSize is the largest bundle read, in bytes.
	maxBundleSize = 1 << 20
	// maxPlatforms is how many platform images of an index are read.
	maxPlatforms = 16
)

const (
	// bundleMediaTypePrefix starts every Sigstore bundle media type.
	bundleMediaTypePrefix = "application/vnd.dev.sigstore.bundle"
	// emptyConfigMediaType is the config of an OCI artifact manifest, which
	// some registries report as a referrer's artifact type.
	emptyConfigMediaType = "application/vnd.oci.empty.v1+json"
	// attestationReference marks BuildKit's attestation manifests (SBOM,
	// provenance) inside an image index; they are not runnable images.
	attestationReference = "attestation-manifest"
)

// readBundles returns the Sigstore bundles attached to an image as OCI
// referrers, through the Referrers API or the referrers tag. A referrer that
// is not a readable bundle is skipped; a registry error fails.
func readBundles(d name.Digest, opts []remote.Option) ([]*bundle.Bundle, error) {
	idx, err := remote.Referrers(d, opts...)
	if err != nil {
		return nil, fmt.Errorf("%w: referrers of %s: %w", errRegistry, d, err)
	}
	im, err := idx.IndexManifest()
	if err != nil {
		return nil, fmt.Errorf("%w: referrers of %s: %w", errRegistry, d, err)
	}
	var out []*bundle.Bundle
	for i := range im.Manifests {
		desc := &im.Manifests[i]
		if i == maxReferrers {
			break
		}
		if t := desc.ArtifactType; t != "" && !isBundle(t) && t != emptyConfigMediaType {
			continue
		}
		bs, err := readReferrer(d.Context().Digest(desc.Digest.String()), opts)
		if err != nil {
			return nil, err
		}
		out = append(out, bs...)
	}
	return out, nil
}

// readReferrer returns the bundles among one referrer manifest's layers.
func readReferrer(ref name.Digest, opts []remote.Option) ([]*bundle.Bundle, error) {
	img, err := remote.Image(ref, opts...)
	if err != nil {
		return nil, fmt.Errorf("%w: referrer %s: %w", errRegistry, ref, err)
	}
	m, err := img.Manifest()
	if err != nil {
		return nil, fmt.Errorf("%w: referrer %s: %w", errRegistry, ref, err)
	}
	var out []*bundle.Bundle
	for i := range m.Layers {
		l := &m.Layers[i]
		if !isBundle(string(l.MediaType)) || l.Size > maxBundleSize {
			continue
		}
		data, err := readLayer(img, l.Digest)
		if err != nil {
			return nil, fmt.Errorf("%w: bundle %s: %w", errRegistry, l.Digest, err)
		}
		var b bundle.Bundle
		if b.UnmarshalJSON(data) != nil {
			continue // not a bundle sigstore-go accepts: no signature
		}
		out = append(out, &b)
	}
	return out, nil
}

// readLayer reads one layer blob, at most maxBundleSize bytes. The registry
// client checks the blob against its digest.
func readLayer(img v1.Image, digest v1.Hash) ([]byte, error) {
	layer, err := img.LayerByDigest(digest)
	if err != nil {
		return nil, fmt.Errorf("layer: %w", err)
	}
	rc, err := layer.Compressed()
	if err != nil {
		return nil, fmt.Errorf("layer: %w", err)
	}
	data, err := io.ReadAll(io.LimitReader(rc, maxBundleSize))
	if closeErr := rc.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return nil, fmt.Errorf("layer: %w", err)
	}
	return data, nil
}

// isBundle reports whether a media or artifact type is a Sigstore bundle.
func isBundle(mediaType string) bool { return strings.HasPrefix(mediaType, bundleMediaTypePrefix) }

// readManifestLabel reads the plugin manifest from the image config's
// ManifestLabel (D-42). For an image index, every platform image must carry
// the same label; BuildKit's attestation manifests are skipped.
func readManifestLabel(d name.Digest, opts []remote.Option) (plugins.Manifest, error) {
	images, err := runnableImages(d, opts)
	if err != nil {
		return plugins.Manifest{}, err
	}
	label := ""
	for i, img := range images {
		value, labelErr := labelOf(img)
		if labelErr != nil {
			return plugins.Manifest{}, fmt.Errorf("%s: %w", d, labelErr)
		}
		if i > 0 && value != label {
			return plugins.Manifest{}, fmt.Errorf("%w: %s: platform images carry different manifests", plugins.ErrManifestLabel, d)
		}
		label = value
	}
	m, err := plugins.ParseManifestLabel(label)
	if err != nil {
		return plugins.Manifest{}, fmt.Errorf("%w: %s: %w", plugins.ErrManifestLabel, d, err)
	}
	return m, nil
}

// runnableImages returns the image d names, or the platform images of the
// index it names.
func runnableImages(d name.Digest, opts []remote.Option) ([]v1.Image, error) {
	desc, err := remote.Get(d, opts...)
	if err != nil {
		return nil, fmt.Errorf("%w: %s: %w", errRegistry, d, err)
	}
	switch {
	case desc.MediaType.IsImage():
		img, imgErr := desc.Image()
		if imgErr != nil {
			return nil, fmt.Errorf("%w: %s: %w", errRegistry, d, imgErr)
		}
		return []v1.Image{img}, nil
	case desc.MediaType.IsIndex():
		return platformImages(desc)
	default:
		return nil, fmt.Errorf("%w: %s has media type %s", plugins.ErrManifestLabel, d, desc.MediaType)
	}
}

// platformImages returns the runnable images of an index: at least one, at
// most maxPlatforms, and no nested index.
func platformImages(desc *remote.Descriptor) ([]v1.Image, error) {
	idx, err := desc.ImageIndex()
	if err != nil {
		return nil, fmt.Errorf("%w: %s: %w", errRegistry, desc.Digest, err)
	}
	im, err := idx.IndexManifest()
	if err != nil {
		return nil, fmt.Errorf("%w: %s: %w", errRegistry, desc.Digest, err)
	}
	var out []v1.Image
	for i := range im.Manifests {
		m := &im.Manifests[i]
		if m.Annotations["vnd.docker.reference.type"] == attestationReference {
			continue
		}
		if !m.MediaType.IsImage() || len(out) == maxPlatforms {
			return nil, fmt.Errorf("%w: %s: unsupported index entry %s", plugins.ErrManifestLabel, desc.Digest, m.Digest)
		}
		img, err := idx.Image(m.Digest)
		if err != nil {
			return nil, fmt.Errorf("%w: %s: %w", errRegistry, m.Digest, err)
		}
		out = append(out, img)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("%w: %s: index has no image", plugins.ErrManifestLabel, desc.Digest)
	}
	return out, nil
}

// labelOf returns an image's ManifestLabel value.
func labelOf(img v1.Image) (string, error) {
	cf, err := img.ConfigFile()
	if err != nil {
		return "", fmt.Errorf("%w: image config: %w", errRegistry, err)
	}
	value, ok := cf.Config.Labels[plugins.ManifestLabel]
	if !ok {
		return "", fmt.Errorf("%w: no %s label", plugins.ErrManifestLabel, plugins.ManifestLabel)
	}
	return value, nil
}

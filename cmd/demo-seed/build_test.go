package main

import "testing"

// TestLoaderBuild checks that a build is named by the SHA-256 of its
// executable: the same bytes always give the same name, other bytes
// another, and the running binary has one (D-161).
func TestLoaderBuild(t *testing.T) {
	t.Parallel()
	const want = "5ea2ed298e3e41687add27d529f52fcd3a13ca6324ab1fdac6262dd8c3226fa4"
	if got := buildDigest([]byte("demo-seed")); got != want || got == buildDigest([]byte("demo-seed v2")) {
		t.Errorf("buildDigest = %s, want %s, and another for other bytes", got, want)
	}
	first, err := loaderBuild()
	if err != nil {
		t.Fatalf("loaderBuild: %v", err)
	}
	if again, err := loaderBuild(); err != nil || again != first || len(first) != len(want) {
		t.Errorf("loaderBuild = %q then %q, %v; want the same digest twice", first, again, err)
	}
}

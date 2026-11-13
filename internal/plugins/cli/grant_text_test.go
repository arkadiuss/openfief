package cli_test

import (
	"strings"
	"testing"
)

// grantLine is what the summary of an installed view plugin says about
// grants (D-209): true whether or not the core runs with SpiceDB.
const grantLine = "If OpenFief runs with SpiceDB authorization, grant it accounts in Settings → Access.\n"

// TestInstallOfAViewPluginSaysToGrantItAccounts checks that the install of a
// view plugin ends by telling the owner where to grant it accounts, since with
// SpiceDB authorization it reads nothing until then, and that the install of a
// connector, which reads nothing of the owner's accounts, does not.
func TestInstallOfAViewPluginSaysToGrantItAccounts(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		image string
		want  bool
	}{
		{name: "a view plugin", image: sampleImage, want: true},
		{name: "a view plugin with a browser path", image: galleryImage, want: true},
		{name: "a connector", image: caImage, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out, err := runOn(installed(), "", "plugin", "install", tt.image, "--yes")
			if err != nil {
				t.Fatalf("install: %v\n%s", err, out)
			}
			if got := strings.Contains(out, grantLine); got != tt.want {
				t.Errorf("output has the grant line = %v, want %v:\n%s", got, tt.want, out)
			}
			if tt.want && !strings.HasSuffix(out, grantLine) {
				t.Errorf("the grant line is not the summary's last:\n%s", out)
			}
		})
	}
}

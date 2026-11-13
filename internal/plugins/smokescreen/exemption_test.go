package smokescreen_test

import (
	"bytes"
	"net/http"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/arkadiuss/openfief/internal/plugins/egress"
	"github.com/arkadiuss/openfief/internal/plugins/smokescreen"
)

// exemptName is the host the exemption tests exempt, listed for plugin
// "sample" and resolving to privateAddr in the fake resolver.
const exemptName = "fixture.example.com"

func mustExempt(t *testing.T, s string) egress.Exemption {
	t.Helper()
	e, err := egress.ParseExemption(s)
	if err != nil {
		t.Fatal(err)
	}
	return e
}

// TestProxyExemptsOneHost checks D-215 on the proxy: the exempt host, at
// exactly its exempt address, is dialed although the address is private, and
// nothing else changes. Every other listed host that resolves into the same
// range is refused with 502 and never dialed, as are the exempt host at
// another address and at a further one, and, with no exemption, the exempt
// host itself.
func TestProxyExemptsOneHost(t *testing.T) {
	t.Parallel()
	exempt := mustExempt(t, exemptName+"="+privateAddr)
	tests := []struct {
		name       string
		exempt     egress.Exemption
		host       string
		wantStatus int
		wantDialed bool
	}{
		{name: "the exempt host at its address", exempt: exempt, host: exemptName, wantStatus: http.StatusOK, wantDialed: true},
		{name: "another host at the exempt address", exempt: exempt, host: "twin.example.com", wantStatus: http.StatusBadGateway},
		{name: "another host in the same range", exempt: exempt, host: "neighbor.example.com", wantStatus: http.StatusBadGateway},
		{name: "another host in the same range and a public address", exempt: exempt, host: "also-mixed.example.com", wantStatus: http.StatusBadGateway},
		{name: "a private host that isn't listed", exempt: exempt, host: "other.example.com", wantStatus: http.StatusProxyAuthRequired},
		{name: "an earlier private host", exempt: exempt, host: "internal.example.com", wantStatus: http.StatusBadGateway},
		{
			name: "the exempt host at another address in the range", exempt: mustExempt(t, "moved.example.com="+privateAddr),
			host: "moved.example.com", wantStatus: http.StatusBadGateway,
		},
		{
			name: "the exempt host at its address and a further one", exempt: mustExempt(t, "also-private.example.com="+privateAddr),
			host: "also-private.example.com", wantStatus: http.StatusBadGateway,
		},
		{
			name: "the exempt host at both of its addresses", exempt: mustExempt(t, "also-private.example.com="+privateAddr+","+otherPrivateAddr),
			host: "also-private.example.com", wantStatus: http.StatusOK, wantDialed: true,
		},
		{name: "no exemption: the same host", host: exemptName, wantStatus: http.StatusBadGateway},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			f := startProxyOptions(t, false, func(o *smokescreen.Options) { o.Exempt = tt.exempt })
			o := f.open(t, http.MethodConnect, tt.host, basic("plugin", sampleToken), nil)
			closeOrFail(t, o.conn)
			if o.status != tt.wantStatus {
				t.Errorf("CONNECT %s = %d, want %d", tt.host, o.status, tt.wantStatus)
			}
			if dialed := len(f.dials.all()) > 0; dialed != tt.wantDialed {
				t.Errorf("dialed %q, want a dial: %t", f.dials.all(), tt.wantDialed)
			}
		})
	}
}

// TestProxyExemptionIsRefusedForAnIPLiteral checks that the exempt address
// without its hostname stays refused, even where a plugin could name it: an
// allow-list holds no IP literal, so the request fails the plugin's
// allow-list, and the exempt address as a host name is refused as before.
func TestProxyExemptionIsRefusedForAnIPLiteral(t *testing.T) {
	t.Parallel()
	f := startProxyOptions(t, false, func(o *smokescreen.Options) { o.Exempt = mustExempt(t, exemptName+"="+privateAddr) })
	o := f.open(t, http.MethodConnect, privateAddr, basic("plugin", sampleToken), nil)
	closeOrFail(t, o.conn)
	if o.status != http.StatusProxyAuthRequired {
		t.Errorf("CONNECT %s = %d, want %d", privateAddr, o.status, http.StatusProxyAuthRequired)
	}
	if dials := f.dials.all(); len(dials) != 0 {
		t.Errorf("dialed %q, want nothing", dials)
	}
}

// TestProxyWarnsOfAnExemption checks that the proxy logs a warning naming
// the exempt host when it starts with an exemption, and none without.
func TestProxyWarnsOfAnExemption(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name   string
		exempt egress.Exemption
		want   bool
	}{
		{name: "an exemption", exempt: mustExempt(t, exemptName+"="+privateAddr), want: true},
		{name: "none"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var out bytes.Buffer
			logger := logrus.New()
			logger.SetOutput(&out)
			logger.SetFormatter(&logrus.JSONFormatter{})
			startProxyOptions(t, false, func(o *smokescreen.Options) { o.Exempt, o.Log = tt.exempt, logger })
			parts := []string{smokescreen.ExemptionWarning, `"level":"warning"`, `"host":"` + exemptName + `"`, `"addresses":["` + privateAddr + `"]`}
			if !tt.want {
				parts = nil
				if bytes.Contains(out.Bytes(), []byte(smokescreen.ExemptionWarning)) {
					t.Errorf("start log %q holds the exemption warning, want none", out.String())
				}
			}
			for _, part := range parts {
				if !bytes.Contains(out.Bytes(), []byte(part)) {
					t.Errorf("start log %q lacks %s", out.String(), part)
				}
			}
		})
	}
}

// TestProxyLogsExemptionDecisions checks the decision lines the end-to-end
// suite reads (D-200 criterion 16, D-215): the exempt host's CONNECT is
// recorded allowed with no error, another private host's with an error that
// names the refused address.
func TestProxyLogsExemptionDecisions(t *testing.T) {
	t.Parallel()
	var out bytes.Buffer
	logger := logrus.New()
	logger.SetOutput(&out)
	logger.SetFormatter(&logrus.JSONFormatter{})
	f := startProxyOptions(t, false, func(o *smokescreen.Options) {
		o.Exempt, o.Log = mustExempt(t, exemptName+"="+privateAddr), logger
	})
	for _, host := range []string{exemptName, "twin.example.com"} {
		o := f.open(t, http.MethodConnect, host, basic("plugin", sampleToken), nil)
		closeOrFail(t, o.conn)
	}
	var allowed, refused bool
	for _, l := range smokescreen.DecisionLines(out.Bytes()) {
		switch l.Text("requested_host") {
		case exemptName + ":" + f.port:
			allowed = l["allow"] == true && l["error"] == nil && l.Text("role") == "sample"
		case "twin.example.com:" + f.port:
			refused = strings.Contains(l.Text("error"), "forbidden destination address")
		}
	}
	if !allowed || !refused {
		t.Errorf("allowed line for the exempt host: %t, refused line for the other host: %t; log:\n%s", allowed, refused, out.String())
	}
}

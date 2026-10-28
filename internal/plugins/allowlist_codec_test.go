package plugins_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/plugins"
)

func TestEncodeAllowLists(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		lists   []plugins.EgressAllowList
		want    map[string]string
		wantErr *fieldError
	}{
		{
			name: "hosts sorted, one per line",
			lists: []plugins.EgressAllowList{
				{Plugin: "sample", Hosts: []plugins.EgressHost{host("data.example.org", 8443), host("api.example.com", 443)}},
				{Plugin: "xtb"},
			},
			want: map[string]string{"sample": "api.example.com:443\ndata.example.org:8443\n", "xtb": ""},
		},
		{name: "no allow-lists", lists: nil, want: map[string]string{}},
		{
			name:    "invalid plugin name",
			lists:   []plugins.EgressAllowList{{Plugin: "Sample"}},
			wantErr: &fieldError{Field: "allowLists[0].plugin", Err: plugins.ErrInvalidName},
		},
		{
			name:    "plugin listed twice",
			lists:   []plugins.EgressAllowList{{Plugin: "sample"}, {Plugin: "sample"}},
			wantErr: &fieldError{Field: "allowLists[1].plugin", Err: plugins.ErrDuplicate},
		},
		{
			name:    "invalid host",
			lists:   []plugins.EgressAllowList{{Plugin: "sample", Hosts: []plugins.EgressHost{host("*.example.com", 443)}}},
			wantErr: &fieldError{Field: "allowLists[0].hosts", Err: plugins.ErrWildcardHost},
		},
		{
			name:    "host listed twice",
			lists:   []plugins.EgressAllowList{{Plugin: "sample", Hosts: []plugins.EgressHost{host("api.example.com", 443), host("api.example.com", 443)}}},
			wantErr: &fieldError{Field: "allowLists[0].hosts", Err: plugins.ErrDuplicate},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := plugins.EncodeAllowLists(tt.lists)
			if diff := cmp.Diff(tt.wantErr, codecError(err), cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("EncodeAllowLists() error (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("EncodeAllowLists() (-want +got):\n%s", diff)
			}
		})
	}
}

func TestDecodeAllowLists(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		data    map[string]string
		want    []plugins.EgressAllowList
		wantErr *fieldError
	}{
		{
			name: "sorted by plugin, hosts sorted",
			data: map[string]string{"xtb": "", "sample": "data.example.org:8443\napi.example.com:443\n"},
			want: []plugins.EgressAllowList{
				{Plugin: "sample", Hosts: []plugins.EgressHost{host("api.example.com", 443), host("data.example.org", 8443)}},
				{Plugin: "xtb"},
			},
		},
		{name: "empty data", data: map[string]string{}, want: []plugins.EgressAllowList{}},
		{
			name:    "key is not a plugin name",
			data:    map[string]string{"not_a_plugin": ""},
			wantErr: &fieldError{Field: `data["not_a_plugin"]`, Err: plugins.ErrInvalidName},
		},
		{
			name:    "port left out",
			data:    map[string]string{"sample": "api.example.com\n"},
			wantErr: &fieldError{Field: `data["sample"][0]`, Err: plugins.ErrInvalidHostname},
		},
		{
			name:    "upper case host",
			data:    map[string]string{"sample": "API.example.com:443\n"},
			wantErr: &fieldError{Field: `data["sample"][0]`, Err: plugins.ErrInvalidHostname},
		},
		{
			name:    "IP literal",
			data:    map[string]string{"sample": "192.0.2.10:443\n"},
			wantErr: &fieldError{Field: `data["sample"][0]`, Err: plugins.ErrIPLiteral},
		},
		{
			name:    "blank line",
			data:    map[string]string{"sample": "api.example.com:443\n\n"},
			wantErr: &fieldError{Field: `data["sample"][1]`, Err: plugins.ErrInvalidHostname},
		},
		{
			name:    "missing final newline",
			data:    map[string]string{"sample": "api.example.com:443"},
			wantErr: &fieldError{Field: `data["sample"]`, Err: plugins.ErrInvalidHostname},
		},
		{
			name:    "host listed twice",
			data:    map[string]string{"sample": strings.Repeat("api.example.com:443\n", 2)},
			wantErr: &fieldError{Field: `data["sample"]`, Err: plugins.ErrDuplicate},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := plugins.DecodeAllowLists(tt.data)
			if diff := cmp.Diff(tt.wantErr, codecError(err), cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("DecodeAllowLists() error (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("DecodeAllowLists() (-want +got):\n%s", diff)
			}
		})
	}
}

func TestAllowListsRoundTrip(t *testing.T) {
	t.Parallel()
	lists := []plugins.EgressAllowList{
		{Plugin: "ibkr", Hosts: []plugins.EgressHost{host("ndcdyn.example.com", 443)}},
		{Plugin: "sample", Hosts: []plugins.EgressHost{host("api.example.com", 443), host("data.example.org", 8443)}},
		{Plugin: "xtb"},
	}
	data, err := plugins.EncodeAllowLists(lists)
	if err != nil {
		t.Fatalf("EncodeAllowLists() error: %v", err)
	}
	got, err := plugins.DecodeAllowLists(data)
	if err != nil {
		t.Fatalf("DecodeAllowLists() error: %v", err)
	}
	if diff := cmp.Diff(lists, got); diff != "" {
		t.Errorf("round trip (-want +got):\n%s", diff)
	}
}

// codecError is the *FieldError an encode or decode failure carries, or nil.
func codecError(err error) *fieldError {
	if err == nil {
		return nil
	}
	fe, ok := errors.AsType[*plugins.FieldError](err)
	if !ok {
		return &fieldError{Err: err}
	}
	return &fieldError{Field: fe.Field, Err: fe.Err}
}

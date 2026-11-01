package main

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/types/descriptorpb"
)

// baseSet is a synthetic base contract: test.v1.Things with ListGadgets(ListGadgetsRequest) returns
// ListGadgetsResponse, which holds Gadget, and GetThing, whose response holds Thing.
func baseSet() *descriptorpb.FileDescriptorSet {
	msg := func(name string, holds ...string) *descriptorpb.DescriptorProto {
		m := &descriptorpb.DescriptorProto{Name: &name}
		for _, h := range holds {
			typeName, kind := ".test.v1."+h, descriptorpb.FieldDescriptorProto_TYPE_MESSAGE
			m.Field = append(m.Field, &descriptorpb.FieldDescriptorProto{TypeName: &typeName, Type: &kind})
		}
		return m
	}
	method := func(name, in, out string) *descriptorpb.MethodDescriptorProto {
		in, out = ".test.v1."+in, ".test.v1."+out
		return &descriptorpb.MethodDescriptorProto{Name: &name, InputType: &in, OutputType: &out}
	}
	file, pkg, svc := "test/v1/things.proto", "test.v1", "Things"
	return &descriptorpb.FileDescriptorSet{File: []*descriptorpb.FileDescriptorProto{{
		Name: &file, Package: &pkg,
		MessageType: []*descriptorpb.DescriptorProto{
			msg("ListGadgetsRequest"), msg("ListGadgetsResponse", "Gadget"), msg("Gadget"),
			msg("GetThingRequest"), msg("GetThingResponse", "Thing"), msg("Thing"),
		},
		Service: []*descriptorpb.ServiceDescriptorProto{{Name: &svc, Method: []*descriptorpb.MethodDescriptorProto{
			method("ListGadgets", "ListGadgetsRequest", "ListGadgetsResponse"),
			method("GetThing", "GetThingRequest", "GetThingResponse"),
		}}},
	}}}
}

// The findings buf reports when ListGadgets and its messages are removed, and when a field changes type.
const (
	rpcGone      = `{"path":"proto/test/v1/things.proto","type":"RPC_NO_DELETE","message":"Previously present RPC \"ListGadgets\" on service \"Things\" was deleted."}`
	requestGone  = `{"path":"proto/test/v1/things.proto","type":"MESSAGE_NO_DELETE","message":"Previously present message \"ListGadgetsRequest\" was deleted from file."}`
	gadgetGone   = `{"path":"proto/test/v1/things.proto","type":"MESSAGE_NO_DELETE","message":"Previously present message \"Gadget\" was deleted from file."}`
	thingGone    = `{"path":"proto/test/v1/things.proto","type":"MESSAGE_NO_DELETE","message":"Previously present message \"Thing\" was deleted from file."}`
	fieldChanged = `{"path":"proto/test/v1/things.proto","type":"FIELD_SAME_TYPE","message":"Field \"1\" with name \"name\" on message \"Thing\" changed type from \"string\" to \"int64\"."}`
)

// The exceptions entries for removing ListGadgets and its messages under D-9.
const (
	rpcEntry     = "D-9 RPC_NO_DELETE test.v1.Things.ListGadgets"
	requestEntry = "D-9 MESSAGE_NO_DELETE test.v1.ListGadgetsRequest"
	gadgetEntry  = "D-9 MESSAGE_NO_DELETE test.v1.Gadget"
)

// design is a synthetic docs/DESIGN.md: D-9 names ListGadgets, D-10 names nothing removed.
const design = "## Decisions\n\n- **[D-9]** The `ListGadgets` method and its messages are removed.\n" +
	"- **[D-10]** Things are listed by name.\n"

func TestJudge(t *testing.T) {
	t.Parallel()
	removal := []string{rpcGone, requestGone, gadgetGone}
	entries := []string{rpcEntry, requestEntry, gadgetEntry}
	tests := []struct {
		name           string
		findings       []string
		base, now      []string
		wantExcused    []string
		wantProblemsIn []string
	}{
		{name: "no change, no entries"},
		{name: "a removal its decision names, entries added with it", findings: removal, now: entries, wantExcused: entries},
		{name: "a removal without entries fails", findings: removal, wantProblemsIn: []string{"not excused", "ListGadgets"}},
		{name: "another breaking change still fails", findings: append([]string{fieldChanged}, removal...), now: entries, wantProblemsIn: []string{"not excused", "changed type"}},
		{name: "entries already on the base excuse nothing again", findings: removal, base: entries, now: entries, wantProblemsIn: []string{"not excused"}},
		{name: "entries already on the base, nothing breaks", base: entries, now: entries},
		{name: "a new entry that excuses nothing fails", now: []string{rpcEntry}, wantProblemsIn: []string{"excuses nothing"}},
		{name: "a message the removed method never held fails", findings: append([]string{thingGone}, removal...), now: append([]string{"D-9 MESSAGE_NO_DELETE test.v1.Thing"}, entries...), wantProblemsIn: []string{"no method D-9 removes holds this message"}},
		{name: "a message entry without its method's entry fails", findings: []string{gadgetGone}, now: []string{gadgetEntry}, wantProblemsIn: []string{"no method D-9 removes holds"}},
		{name: "a decision that doesn't name the method fails", findings: []string{rpcGone}, now: []string{"D-10 RPC_NO_DELETE test.v1.Things.ListGadgets"}, wantProblemsIn: []string{"D-10 doesn't name `ListGadgets`"}},
		{name: "a decision that doesn't exist fails", findings: []string{rpcGone}, now: []string{"D-99 RPC_NO_DELETE test.v1.Things.ListGadgets"}, wantProblemsIn: []string{"no author decision D-99"}},
		{name: "an old entry citing a missing decision fails", base: []string{"D-99 RPC_NO_DELETE test.v1.Things.Gone"}, now: []string{"D-99 RPC_NO_DELETE test.v1.Things.Gone"}, wantProblemsIn: []string{"no author decision D-99"}},
		{name: "a method the base lacks fails", findings: []string{rpcGone}, now: []string{rpcEntry, "D-9 RPC_NO_DELETE test.v1.Things.ListGadgetsAgain"}, wantProblemsIn: []string{"the base has no such method"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			in := judgeInput{Decisions: parseDecisions(design), Base: indexContracts(baseSet())}
			var err error
			if in.Findings, err = parseFindings(strings.Join(tt.findings, "\n")); err != nil {
				t.Fatalf("parseFindings: %v", err)
			}
			if in.BaseExceptions, err = parseExceptions(strings.Join(tt.base, "\n")); err != nil {
				t.Fatalf("parseExceptions(base): %v", err)
			}
			if in.Exceptions, err = parseExceptions(strings.Join(tt.now, "\n")); err != nil {
				t.Fatalf("parseExceptions(now): %v", err)
			}
			excused, err := judge(&in)
			checkVerdict(t, excused, err, tt.wantExcused, tt.wantProblemsIn)
		})
	}
}

// checkVerdict compares judge's result with the entries it should excuse, or the words its error should
// hold.
func checkVerdict(t *testing.T, excused []Entry, err error, wantExcused, wantProblemsIn []string) {
	t.Helper()
	if len(wantProblemsIn) > 0 {
		if err == nil {
			t.Fatalf("judge excused %v, want a failure naming %q", excused, wantProblemsIn)
		}
		for _, w := range wantProblemsIn {
			if !strings.Contains(err.Error(), w) {
				t.Errorf("judge's error %q doesn't say %q", err, w)
			}
		}
		return
	}
	if err != nil {
		t.Fatalf("judge: %v", err)
	}
	got := make([]string, 0, len(excused))
	for _, e := range excused {
		got = append(got, e.String())
	}
	if diff := cmp.Diff(wantExcused, got, cmpopts.EquateEmpty()); diff != "" {
		t.Errorf("excused (-want +got):\n%s", diff)
	}
}

func TestParseExceptions(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, text, wantErr string
		want                []Entry
	}{
		{name: "comments and blank lines", text: "# header\n\n" + rpcEntry + "\n", want: []Entry{{Decision: "D-9", Rule: ruleRPCDelete, Element: "test.v1.Things.ListGadgets"}}},
		{name: "a field missing", text: "D-9 RPC_NO_DELETE", wantErr: "want DECISION RULE ELEMENT"},
		{name: "not a decision", text: "P-9 RPC_NO_DELETE test.v1.Things.ListGadgets", wantErr: "no author decision ID"},
		{name: "a rule other than a removal", text: "D-9 FIELD_SAME_TYPE test.v1.Thing.name", wantErr: "only RPC_NO_DELETE and MESSAGE_NO_DELETE"},
		{name: "a name that isn't qualified", text: "D-9 MESSAGE_NO_DELETE Gadget", wantErr: "no fully qualified protobuf name"},
		{name: "an entry twice", text: rpcEntry + "\n" + rpcEntry, wantErr: "listed twice"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := parseExceptions(tt.text)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("parseExceptions = %v, %v; want an error saying %q", got, err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseExceptions: %v", err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("entries (-want +got):\n%s", diff)
			}
		})
	}
}

func TestParseFindingsRefusesWhatItCantRead(t *testing.T) {
	t.Parallel()
	for _, text := range []string{"not json", `{"path":"a.proto","type":"RPC_NO_DELETE"}`} {
		if _, err := parseFindings(text); err == nil {
			t.Errorf("parseFindings(%q) succeeded, want an error", text)
		}
	}
}

package openaichat

import "testing"

func TestCompleteReadsTheCachedTokensTheHostReports(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		usage string
		want  int64
	}{
		{name: "in the prompt details", usage: `{"prompt_tokens":900,"prompt_tokens_details":{"cached_tokens":640}}`, want: 640},
		{name: "at the top of the usage", usage: `{"prompt_tokens":900,"cached_tokens":12}`, want: 12},
		{name: "a null detail", usage: `{"prompt_tokens":900,"prompt_tokens_details":null}`},
		{name: "none", usage: `{"prompt_tokens":900}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, c := newFake(t, `{"choices":[{"message":{"content":"a"}}],"usage":`+tt.usage+`}`)
			reply, err := c.Complete(t.Context(), &Request{Model: "m", Messages: []Message{{Role: "user", Content: "hi"}}})
			if err != nil {
				t.Fatal(err)
			}
			if reply.CachedTokens != tt.want {
				t.Errorf("cached tokens %d, want %d", reply.CachedTokens, tt.want)
			}
		})
	}
}

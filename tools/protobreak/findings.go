package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// errFindings reports output of buf breaking this tool can't read.
var errFindings = errors.New("unreadable buf breaking finding")

// Finding is one breaking change buf breaking reported against the base: the file, buf's rule ID and its
// message.
type Finding struct {
	Path    string `json:"path"`
	Rule    string `json:"type"`
	Message string `json:"message"`
}

// String returns the finding as buf prints it, without the position.
func (f Finding) String() string { return f.Path + ": " + f.Message }

// The messages of the removals an entry can excuse, which name what was removed.
var (
	rpcDeleted     = regexp.MustCompile(`^Previously present RPC "([A-Za-z0-9_]+)" on service "([A-Za-z0-9_]+)" was deleted\.$`)
	messageDeleted = regexp.MustCompile(`^Previously present message "([A-Za-z0-9_]+)" was deleted from file\.$`)
)

// parseFindings reads buf breaking's --error-format json output: one JSON object a line.
func parseFindings(text string) ([]Finding, error) {
	var out []Finding
	for line := range strings.SplitSeq(text, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var f Finding
		if err := json.Unmarshal([]byte(line), &f); err != nil {
			return nil, fmt.Errorf("%w: %q: %w", errFindings, line, err)
		}
		if f.Path == "" || f.Rule == "" || f.Message == "" {
			return nil, fmt.Errorf("%w: %q: no path, type or message", errFindings, line)
		}
		out = append(out, f)
	}
	return out, nil
}

// removed returns the fully qualified name of the method or top-level message f reports removed, with
// the package base holds for f's file; ok is false for any other finding, which no entry can excuse.
func removed(f *Finding, base *contracts) (string, bool) {
	pkg, ok := base.packageOf(f.Path)
	if !ok {
		return "", false
	}
	switch f.Rule {
	case ruleRPCDelete:
		if m := rpcDeleted.FindStringSubmatch(f.Message); m != nil {
			return pkg + "." + m[2] + "." + m[1], true
		}
	case ruleMessageDelete:
		if m := messageDeleted.FindStringSubmatch(f.Message); m != nil {
			return pkg + "." + m[1], true
		}
	}
	return "", false
}

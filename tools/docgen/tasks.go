package main

import (
	"fmt"
	"strings"
)

// taskList is the part of `task --list-all --json` that docgen reads.
type taskList struct {
	Tasks []task `json:"tasks"`
}

type task struct {
	Name string `json:"name"`
	Desc string `json:"desc"`
}

// tasksPage renders every task target, grouped by the part of its name before the first colon.
func tasksPage(list taskList) string {
	var b strings.Builder
	b.WriteString("---\ntitle: Tasks\ndescription: Every task target, the single entry point for local work, git hooks and CI.\nweight: 30\n---\n\n")
	b.WriteString("Generated from [Taskfile.yml](../../../../Taskfile.yml) on every build. The [delivery pipeline](../../../../docs/dev/pipeline.md) page explains how they fit together.\n")
	group := ""
	for _, t := range list.Tasks {
		g, _, _ := strings.Cut(t.Name, ":")
		if g != group {
			group = g
			fmt.Fprintf(&b, "\n## %s\n\n| Task | What it does |\n|---|---|\n", group)
		}
		fmt.Fprintf(&b, "| `task %s` | %s |\n", t.Name, tableCell(t.Desc))
	}
	return b.String()
}

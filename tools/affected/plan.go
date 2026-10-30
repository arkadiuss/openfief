package main

import (
	"fmt"
	"maps"
	"path"
	"slices"
	"strings"
)

// change is one path a commit or a range adds, modifies or deletes.
type change struct {
	path    string
	deleted bool
}

// step is one task target the hook runs, with its variables and why it runs.
type step struct {
	task string
	vars []string
	why  string
	// details list each package of a Go step with why it is in it.
	details []string
}

// plan is what the hook runs for a change: either the full check, or the listed steps.
type plan struct {
	full bool
	// reasons say why the plan is full, or, for a partial plan, what each changed path selected.
	reasons []string
	steps   []step
	// skipped names the steps left out, with why.
	skipped []step
}

// selection is what a change set touches, before it becomes steps.
type selection struct {
	full        []string
	notes       []string
	areas       map[area]bool
	deleted     bool
	goFiles     map[string][]string // module -> changed, existing .go files
	changedPkgs map[string]string   // import path -> why
	dataPkgs    map[string]string   // import path -> why (tests only)
	scripts     map[string]string   // script test -> why
	paths       []string
}

// scriptTest is a test in scripts/tests and the repository paths it reads ($root/<path>).
type scriptTest struct {
	file  string
	reads []string
}

// operatorDirs hold the operator's API types and the files controller-gen writes from them (task gen:check).
func operatorDirs() []string {
	return []string{"internal/plugins/operator", "deploy/crds", "deploy/operator"}
}

// labelInput is an input of the plugin label check (task plugins:labels:check): a first-party plugin's
// manifest.json or Dockerfile, or a demo Plugin under deploy/demo.
func labelInput(file string) bool {
	parts := strings.Split(file, "/")
	const pluginFile = 3 // plugins/<name>/<file>
	if len(parts) == pluginFile && parts[0] == "plugins" {
		return parts[2] == "manifest.json" || parts[2] == "Dockerfile"
	}
	return within(file, "deploy/demo")
}

// selectChanges resolves every changed path against the rules, the package graph and the script tests.
func selectChanges(changes []change, g *graph, scripts []scriptTest) selection {
	s := selection{
		areas: map[area]bool{}, goFiles: map[string][]string{},
		changedPkgs: map[string]string{}, dataPkgs: map[string]string{}, scripts: map[string]string{},
	}
	for _, c := range changes {
		s.paths = append(s.paths, c.path)
		s.deleted = s.deleted || c.deleted
		r := classify(c.path)
		a, why := r.area, r.why
		s.areas[a] = true
		for _, t := range scripts {
			if slices.ContainsFunc(t.reads, func(r string) bool { return within(c.path, r) }) {
				s.scripts[t.file] = "reads " + c.path
			}
		}
		switch a {
		case areaFull:
			s.full = append(s.full, c.path+": "+why)
		case areaUnknown:
			s.full = append(s.full, c.path+": "+why+"; when in doubt, everything runs")
		case areaCode:
			s.selectCode(c, g)
		case areaDocs, areaKibi, areaWorkflows, areaDeploy, areaQuiet:
			s.notes = append(s.notes, c.path+": "+why)
			s.selectReaders(c, g)
		}
	}
	return s
}

// selectCode maps a changed Go file to its package, an embedded file to the packages embedding it, and any
// other file to the package around it (test data) and the packages whose tests name it.
func (s *selection) selectCode(c change, g *graph) {
	if path.Ext(c.path) == ".go" && !strings.Contains(c.path, "/testdata/") {
		ip, ok := g.byDir[path.Dir(c.path)]
		if !ok {
			s.full = append(s.full, c.path+": Go file outside any package (a removed or new package?); when in doubt, everything runs")
			return
		}
		s.changedPkgs[ip] = "changed " + c.path
		if !c.deleted {
			mod := g.pkgs[ip].module
			s.goFiles[mod] = append(s.goFiles[mod], c.path)
		}
		s.notes = append(s.notes, c.path+": Go source of "+shortPath(ip))
		return
	}
	if owners := g.embedOwners[c.path]; len(owners) > 0 {
		for _, ip := range owners {
			s.changedPkgs[ip] = "embeds " + c.path
		}
		s.notes = append(s.notes, c.path+": embedded in "+strings.Join(owners, ", "))
		return
	}
	found := s.selectReaders(c, g)
	if ip, ok := g.owningPackage(c.path); ok {
		s.dataPkgs[ip] = "test data " + c.path
		found = true
	}
	if !found {
		s.full = append(s.full, c.path+": no package holds or reads it; when in doubt, everything runs")
		return
	}
	s.notes = append(s.notes, c.path+": data read by tests")
}

// selectReaders adds the packages whose code or tests name a path containing the changed file.
func (s *selection) selectReaders(c change, g *graph) bool {
	readers := g.readers(c.path)
	for _, ip := range readers {
		s.dataPkgs[ip] = "reads " + c.path
	}
	return len(readers) > 0
}

// affectedSets are the packages a selection reaches: compiled ones (the changed packages and everything
// that imports them), the ones whose tests must run, and the ones lint must check.
type affectedSets struct {
	compiled map[string]string
	tested   map[string]string
	linted   map[string]string
}

// reach follows reverse imports from the changed packages, then adds the packages whose tests import a
// compiled one, and the packages whose test data changed.
func reach(s *selection, g *graph) affectedSets {
	compiled := importersOf(s.changedPkgs, g)
	linted := maps.Clone(compiled)
	for _, ip := range slices.Sorted(maps.Keys(compiled)) {
		for _, imp := range g.testImporters[ip] {
			if _, ok := linted[imp]; !ok {
				linted[imp] = "tests import " + shortPath(ip)
			}
		}
	}
	tested := maps.Clone(linted)
	for ip, why := range s.dataPkgs {
		if _, ok := tested[ip]; !ok {
			tested[ip] = why
		}
	}
	maps.DeleteFunc(tested, func(ip, _ string) bool { p, ok := g.pkgs[ip]; return !ok || !p.testable })
	return affectedSets{compiled: compiled, tested: tested, linted: linted}
}

// importersOf is the changed packages and every package importing one of them, directly or not, each with
// why it is there.
func importersOf(changed map[string]string, g *graph) map[string]string {
	out := maps.Clone(changed)
	queue := slices.Sorted(maps.Keys(out))
	for len(queue) > 0 {
		ip := queue[0]
		queue = queue[1:]
		for _, imp := range g.importers[ip] {
			if _, ok := out[imp]; !ok {
				out[imp] = "imports " + shortPath(ip)
				queue = append(queue, imp)
			}
		}
	}
	return out
}

// hookPlan turns a staged change into the steps of `task check:commit`, in the order of `task check`.
func hookPlan(changes []change, g *graph, scripts []scriptTest, modules []string) plan {
	s := selectChanges(changes, g, scripts)
	if len(s.full) > 0 {
		return plan{full: true, reasons: s.full}
	}
	sets := reach(&s, g)
	b := &planBuilder{plan: plan{reasons: s.notes}}
	b.add(touches(s.paths, operatorDirs()) || anyUnder(sets.compiled, g, "internal/plugins/operator"),
		&step{task: "gen:check", why: "the operator's API types or the manifests generated from them may change"},
		"no operator code or generated manifest changed")
	b.add(slices.ContainsFunc(s.paths, labelInput),
		&step{task: "plugins:labels:check", why: "a plugin's manifest, Dockerfile or a demo Plugin changed"},
		"no plugin manifest, Dockerfile or demo Plugin changed")
	for _, m := range modules {
		b.add(len(s.goFiles[m]) > 0,
			&step{task: "fmt:check:files", vars: []string{"MODULE=" + m, "FILES=" + inModule(m, s.goFiles[m])}, why: "changed Go files"},
			"no Go file changed in module "+m)
	}
	b.goSteps(sets, g, modules)
	b.add(s.areas[areaWorkflows], &step{task: "lint:ci", why: "workflows changed"}, "no workflow changed")
	b.skip("proto", "no contract changed (proto/ and buf files select the full check)")
	b.scriptSteps(s.scripts)
	b.add(true, &step{task: "docs:check", why: "always: it reads only the design documents and takes a second"}, "")
	b.add(s.areas[areaDocs] || s.areas[areaKibi] || s.deleted,
		&step{task: "docs:build", why: "docs, requirements or a deleted file (a page may link to it) changed"},
		"no docs, requirement or deleted file: a code change can't break a page's links")
	b.add(s.areas[areaDeploy], &step{task: "demo:validate", why: "manifests changed"}, "no manifest changed")
	b.add(true, &step{task: "kibi:staged", why: "always: traceability of the staged change"}, "")
	return b.plan
}

// planBuilder collects the steps that run and the ones that don't.
type planBuilder struct{ plan plan }

func (b *planBuilder) add(run bool, s *step, whyNot string) {
	if run {
		b.plan.steps = append(b.plan.steps, *s)
		return
	}
	b.skip(s.task, whyNot)
}

func (b *planBuilder) skip(task, why string) {
	b.plan.skipped = append(b.plan.skipped, step{task: task, why: why})
}

// goSteps adds lint, vulnerability checks and tests for the affected packages of each module.
func (b *planBuilder) goSteps(sets affectedSets, g *graph, modules []string) {
	b.add(len(sets.linted) > 0, &step{task: "lint:config", why: "Go packages are linted"}, "no Go package affected")
	for _, m := range modules {
		linted, tested, compiled := byModule(sets.linted, g, m), byModule(sets.tested, g, m), byModule(sets.compiled, g, m)
		b.add(len(linted) > 0,
			goStep("lint:go", m, linted, sets.linted, g),
			"no package affected in module "+m)
		b.add(len(compiled) > 0,
			goStep("vuln:go", m, compiled, sets.compiled, g),
			"no compiled package affected in module "+m)
		b.add(len(tested) > 0,
			goStep("test:go", m, tested, sets.tested, g),
			"no package to test in module "+m)
	}
}

// scriptSteps adds each script test that reads a changed path.
func (b *planBuilder) scriptSteps(scripts map[string]string) {
	if len(scripts) == 0 {
		b.skip("test:script", "no script test reads a changed path")
	}
	for _, file := range slices.Sorted(maps.Keys(scripts)) {
		b.add(true, &step{task: "test:script", vars: []string{"SCRIPT=" + file}, why: scripts[file]}, "")
	}
}

// touches reports whether any path lies in one of dirs.
func touches(paths, dirs []string) bool {
	return slices.ContainsFunc(paths, func(p string) bool {
		return slices.ContainsFunc(dirs, func(d string) bool { return within(p, d) })
	})
}

// anyUnder reports whether a set holds a package in dir or below it.
func anyUnder(set map[string]string, g *graph, dir string) bool {
	for ip := range set {
		if p, ok := g.pkgs[ip]; ok && within(p.dir, dir) {
			return true
		}
	}
	return false
}

// byModule is the sorted import paths of a set that belong to a module.
func byModule(set map[string]string, g *graph, module string) []string {
	var out []string
	for ip := range set {
		if p, ok := g.pkgs[ip]; ok && p.module == module {
			out = append(out, ip)
		}
	}
	slices.Sort(out)
	return out
}

// goStep runs a task on packages of one module, passed as ./dir arguments relative to the module.
func goStep(task, module string, ips []string, why map[string]string, g *graph) *step {
	dirs := make([]string, 0, len(ips))
	details := make([]string, 0, len(ips))
	for _, ip := range ips {
		dir := g.pkgs[ip].dir
		if module != "." {
			dir = strings.TrimPrefix(dir, module+"/")
		}
		dirs = append(dirs, "./"+dir)
		details = append(details, fmt.Sprintf("%s: %s", shortPath(ip), why[ip]))
	}
	return &step{
		task: task, vars: []string{"MODULE=" + module, "PKGS=" + strings.Join(dirs, " ")},
		why: fmt.Sprintf("%d package(s)", len(ips)), details: details,
	}
}

// inModule is the files relative to their module, space-separated.
func inModule(module string, files []string) string {
	out := make([]string, 0, len(files))
	for _, f := range files {
		if module != "." {
			f = strings.TrimPrefix(f, module+"/")
		}
		out = append(out, f)
	}
	return strings.Join(out, " ")
}

// shortPath drops the module path prefix from an import path for the report.
func shortPath(ip string) string {
	const modulePath = "github.com/arkadiuss/openfief/"
	return strings.TrimPrefix(ip, modulePath)
}

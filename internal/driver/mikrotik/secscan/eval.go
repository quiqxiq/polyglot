package secscan

import (
	"sort"
)

var weights = map[Severity]int{
	Critical: 25,
	High:     12,
	Medium:   6,
	Low:      2,
	Info:     0,
}

// CriticalCap is the max score a router with a failed critical check can receive.
const CriticalCap = 59

// Checks is the complete security catalogue.
var Checks []Check

func init() {
	Checks = make([]Check, 0, len(checksMgmt)+len(checksFW)+len(checksServices)+len(checksAcctSys))
	Checks = append(Checks, checksMgmt...)
	Checks = append(Checks, checksFW...)
	Checks = append(Checks, checksServices...)
	Checks = append(Checks, checksAcctSys...)
}

type catTotals struct {
	answered, failed    int
	pass, fail, unknown int
	criticalFailed      bool
}

func (t *catTotals) add(s Severity, st Status) {
	switch st {
	case Pass:
		t.pass++
		t.answered += weights[s]
	case Fail:
		t.fail++
		t.answered += weights[s]
		t.failed += weights[s]
		if s == Critical {
			t.criticalFailed = true
		}
	default:
		t.unknown++
	}
}

func (t *catTotals) score() int {
	if t.answered == 0 {
		return 100
	}
	s := (100*(t.answered-t.failed) + t.answered/2) / t.answered
	if t.criticalFailed && s > CriticalCap {
		s = CriticalCap
	}
	return s
}

// Run evaluates all checks in the catalogue against the provided inputs and calculates the score.
func Run(in Inputs) Report {
	rep := Report{
		Findings:   []Finding{},
		Categories: []CategoryScore{},
		Failed:     map[string]int{},
	}
	for _, s := range Severities {
		rep.Failed[string(s)] = 0
	}

	cats := map[string]*catTotals{}
	var order []string
	all := &catTotals{}

	for _, c := range Checks {
		st, detail := c.Eval(in)
		if detail == nil {
			detail = []string{}
		}
		rep.Findings = append(rep.Findings, Finding{
			ID:       c.ID,
			Category: c.Category,
			Severity: c.Severity,
			Title:    c.Title,
			Status:   st,
			Detail:   detail,
			Why:      c.Why,
			Fix:      c.Fix,
			Link:     c.Link,
		})

		ct := cats[c.Category]
		if ct == nil {
			ct = &catTotals{}
			cats[c.Category] = ct
			order = append(order, c.Category)
		}
		ct.add(c.Severity, st)
		all.add(c.Severity, st)

		switch st {
		case Fail:
			rep.Failed[string(c.Severity)]++
		case Pass:
			rep.Passed++
		default:
			rep.Unknown++
		}
	}

	rep.Score = all.score()
	for _, name := range order {
		ct := cats[name]
		rep.Categories = append(rep.Categories, CategoryScore{
			Name:    name,
			Score:   ct.score(),
			Pass:    ct.pass,
			Fail:    ct.fail,
			Unknown: ct.unknown,
		})
	}

	rank := map[Status]int{Fail: 0, Pass: 1, Unknown: 2}
	sevRank := map[Severity]int{}
	for i, s := range Severities {
		sevRank[s] = i
	}
	sort.SliceStable(rep.Findings, func(i, j int) bool {
		a, b := rep.Findings[i], rep.Findings[j]
		if rank[a.Status] != rank[b.Status] {
			return rank[a.Status] < rank[b.Status]
		}
		return sevRank[a.Severity] < sevRank[b.Severity]
	})

	rep.Facts = buildFacts(in)
	return rep
}

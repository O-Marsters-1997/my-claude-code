package digest

import (
	"cmp"
	"slices"
)

type Instance struct {
	Agent string
	Line  int
}

type Cluster struct {
	Mechanism string
	Instances []Instance
}

func (c Cluster) Agents() []string {
	var out []string
	for _, i := range c.Instances {
		if !slices.Contains(out, i.Agent) {
			out = append(out, i.Agent)
		}
	}
	return out
}

// Clusters groups failures by the mechanism that raised them, keeps those
// hitting two or more agents, and marks their signals as clustered in place.
func Clusters(ds []Digest) []Cluster {
	byMech := map[string]*Cluster{}
	var order []string
	for _, d := range ds {
		seen := map[int]bool{}
		for _, s := range d.Signals {
			if s.Mechanism == "" || seen[s.Line] {
				continue
			}
			seen[s.Line] = true
			c := byMech[s.Mechanism]
			if c == nil {
				c = &Cluster{Mechanism: s.Mechanism}
				byMech[s.Mechanism] = c
				order = append(order, s.Mechanism)
			}
			c.Instances = append(c.Instances, Instance{d.Agent.ID, s.Line})
		}
	}
	var out []Cluster
	for _, m := range order {
		c := byMech[m]
		if len(c.Agents()) < 2 {
			continue
		}
		out = append(out, *c)
	}
	slices.SortStableFunc(out, func(a, b Cluster) int { return cmp.Compare(len(b.Instances), len(a.Instances)) })
	for _, c := range out {
		for _, d := range ds {
			for i := range d.Signals {
				if d.Signals[i].Mechanism == c.Mechanism {
					d.Signals[i].Cluster = c.Mechanism
				}
			}
		}
	}
	return out
}

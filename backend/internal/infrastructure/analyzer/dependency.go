package analyzer

import "github.com/uber/unm-platform/internal/domain/entity"

// DependencyCycle holds the ordered list of node names that form a cycle.
// The first and last element are the same to clearly show the loop,
// e.g. ["svc-a", "svc-b", "svc-c", "svc-a"].
type DependencyCycle struct {
	Path []string
}

// DependencyReport holds the result of a dependency analysis pass over a UNMModel.
type DependencyReport struct {
	// ServiceCycles contains detected cycles in the service dependency graph.
	ServiceCycles []DependencyCycle
	// CapabilityCycles contains detected cycles in the capability dependency graph.
	CapabilityCycles []DependencyCycle
	// MaxServiceDepth is the longest chain depth in the service dependency graph.
	MaxServiceDepth int
	// MaxCapabilityDepth is the longest chain depth in the capability dependency graph.
	MaxCapabilityDepth int
	// CriticalServicePath is the service names forming the longest dependency chain.
	CriticalServicePath []string
}

// DependencyAnalyzer detects cycles and computes depth metrics in dependency graphs.
type DependencyAnalyzer struct{}

// NewDependencyAnalyzer constructs a DependencyAnalyzer.
func NewDependencyAnalyzer() *DependencyAnalyzer {
	return &DependencyAnalyzer{}
}

// Analyze inspects m and returns a DependencyReport.
func (a *DependencyAnalyzer) Analyze(m *entity.UNMModel) DependencyReport {
	var report DependencyReport

	// Build adjacency lists for services (by name).
	svcAdj := make(map[string][]string, len(m.Services))
	for name, svc := range m.Services {
		for _, rel := range svc.DependsOn {
			target := rel.TargetID.String()
			if target == name {
				continue // skip self-loops
			}
			svcAdj[name] = append(svcAdj[name], target)
		}
	}

	// Detect service cycles.
	report.ServiceCycles = detectDependencyCycles(svcAdj)

	// Build adjacency list for capabilities (DependsOn).
	capAdj := make(map[string][]string, len(m.Capabilities))
	for name, cap := range m.Capabilities {
		for _, rel := range cap.DependsOn {
			target := rel.TargetID.String()
			if target == name {
				continue // skip self-loops
			}
			capAdj[name] = append(capAdj[name], target)
		}
	}

	// Detect capability cycles.
	report.CapabilityCycles = detectDependencyCycles(capAdj)

	// Compute critical path for services (only meaningful when no cycles).
	if len(report.ServiceCycles) == 0 && len(m.Services) > 0 {
		report.CriticalServicePath, report.MaxServiceDepth = criticalPath(svcAdj)
	} else {
		// Still compute depth even with cycles (use visited guard to avoid infinite loop).
		report.MaxServiceDepth = maxDepthWithGuard(svcAdj)
	}

	// Compute capability depth.
	if len(report.CapabilityCycles) == 0 && len(m.Capabilities) > 0 {
		_, report.MaxCapabilityDepth = criticalPath(capAdj)
	} else {
		report.MaxCapabilityDepth = maxDepthWithGuard(capAdj)
	}

	return report
}

// detectDependencyCycles runs DFS over the adjacency list and returns each distinct cycle once.
func detectDependencyCycles(adj map[string][]string) []DependencyCycle {
	const (
		white = 0 // unvisited
		gray  = 1 // in current DFS stack
		black = 2 // fully processed
	)
	color := make(map[string]int)
	parent := make(map[string]string)
	var cycles []DependencyCycle
	cycleSet := make(map[string]bool) // deduplicate by canonical key

	var dfs func(node string)
	dfs = func(node string) {
		color[node] = gray
		for _, neighbor := range adj[node] {
			if color[neighbor] == gray {
				// Found a cycle — reconstruct it.
				cycle := []string{neighbor}
				cur := node
				for cur != neighbor {
					cycle = append([]string{cur}, cycle...)
					p, ok := parent[cur]
					if !ok {
						break
					}
					cur = p
				}
				// Close the cycle back to its start node.
				cycle = append(cycle, cycle[0])
				key := canonicalizeCycle(cycle)
				if !cycleSet[key] {
					cycleSet[key] = true
					cycles = append(cycles, DependencyCycle{Path: cycle})
				}
			} else if color[neighbor] == white {
				parent[neighbor] = node
				dfs(neighbor)
			}
		}
		color[node] = black
	}

	for node := range adj {
		if color[node] == white {
			dfs(node)
		}
	}

	return cycles
}

// canonicalizeCycle returns a string key that is invariant to the rotation of the cycle.
func canonicalizeCycle(cycle []string) string {
	if len(cycle) == 0 {
		return ""
	}
	// Find smallest element.
	minIdx := 0
	for i, v := range cycle {
		if v < cycle[minIdx] {
			minIdx = i
		}
	}
	key := ""
	n := len(cycle)
	for i := 0; i < n; i++ {
		key += cycle[(minIdx+i)%n] + "|"
	}
	return key
}

// criticalPath returns the longest path (by node count) and its length in an acyclic graph.
func criticalPath(adj map[string][]string) ([]string, int) {
	// Compute depth from each node using memoized DFS.
	memo := make(map[string][]string)
	var longestFrom func(node string) []string
	longestFrom = func(node string) []string {
		if p, ok := memo[node]; ok {
			return p
		}
		best := []string{node}
		for _, neighbor := range adj[node] {
			sub := longestFrom(neighbor)
			if 1+len(sub) > len(best) {
				combined := make([]string, 1+len(sub))
				combined[0] = node
				copy(combined[1:], sub)
				best = combined
			}
		}
		memo[node] = best
		return best
	}

	var critical []string
	for node := range adj {
		p := longestFrom(node)
		if len(p) > len(critical) {
			critical = p
		}
	}
	// Also consider nodes that may only be targets (leaf nodes with no outgoing edges).
	// They are not keys in adj but are referenced as values.
	seen := make(map[string]bool)
	for n := range adj {
		seen[n] = true
	}
	for _, neighbors := range adj {
		for _, neighbor := range neighbors {
			if !seen[neighbor] {
				seen[neighbor] = true
				p := longestFrom(neighbor)
				if len(p) > len(critical) {
					critical = p
				}
			}
		}
	}

	return critical, len(critical)
}

// maxDepthWithGuard computes the maximum depth using a visited guard (for graphs with cycles).
func maxDepthWithGuard(adj map[string][]string) int {
	visited := make(map[string]bool)
	var dfs func(node string) int
	dfs = func(node string) int {
		if visited[node] {
			return 0
		}
		visited[node] = true
		best := 1
		for _, neighbor := range adj[node] {
			d := 1 + dfs(neighbor)
			if d > best {
				best = d
			}
		}
		visited[node] = false
		return best
	}

	max := 0
	for node := range adj {
		d := dfs(node)
		if d > max {
			max = d
		}
	}
	return max
}

package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/krzachariassen/unm-platform/internal/domain/entity"
	"github.com/krzachariassen/unm-platform/internal/domain/service"
	"github.com/krzachariassen/unm-platform/internal/infrastructure/analyzer"
	infraconfig "github.com/krzachariassen/unm-platform/internal/infrastructure/config"
	"github.com/krzachariassen/unm-platform/internal/infrastructure/parser"
	"github.com/krzachariassen/unm-platform/internal/usecase"
)

func main() {
	// Extract --env flag before dispatching commands.
	env := os.Getenv("UNM_ENV")
	args := os.Args[1:]
	filtered := args[:0]
	for _, a := range args {
		if strings.HasPrefix(a, "--env=") {
			env = strings.TrimPrefix(a, "--env=")
		} else {
			filtered = append(filtered, a)
		}
	}
	args = filtered

	cfg, err := infraconfig.LoadConfig(env)
	if err != nil {
		fmt.Fprintf(os.Stderr, "config load error: %v\n", err)
		cfg = &entity.Config{} // fall back to zero — will fail validate but let commands error explicitly
		*cfg = entity.DefaultConfig()
	}

	if len(args) < 1 {
		printUsage(os.Stderr)
		os.Exit(1)
	}
	switch args[0] {
	case "parse":
		os.Exit(runParseCommand(args[1:], os.Stdout))
	case "validate":
		os.Exit(runValidateCommand(args[1:], os.Stdout))
	case "analyze":
		os.Exit(runAnalyzeCommand(args[1:], os.Stdout, *cfg))
	default:
		printUsage(os.Stderr)
		os.Exit(1)
	}
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage: unm-cli <command> [options] <file>")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Commands:")
	fmt.Fprintln(w, "  parse <file>                      Parse and validate a .unm.yaml or .unm file, print full summary")
	fmt.Fprintln(w, "  validate <file>                   Validate a .unm.yaml or .unm file, print validation result only")
	fmt.Fprintln(w, "  analyze <type> <file>             Run analysis on a .unm.yaml or .unm file")
	fmt.Fprintln(w, "    Types: fragmentation, cognitive-load, dependencies, gaps, bottleneck, coupling, complexity, interactions, all")
}

// runParseCommand parses and validates the given file, printing a full summary to w.
// Returns: 0 = valid (even with warnings), 1 = parse error, 2 = validation errors.
func runParseCommand(args []string, w io.Writer) int {
	if len(args) < 1 {
		fmt.Fprintln(w, "Error: parse requires a file argument")
		return 1
	}
	path := args[0]

	f, err := os.Open(path)
	if err != nil {
		fmt.Fprintf(w, "Parse error: %v\n", err)
		return 1
	}
	defer f.Close()

	uc := usecase.NewParseAndValidate(parser.NewParserForPath(path), service.NewValidationEngine())
	model, result, err := uc.Execute(f)
	if err != nil {
		fmt.Fprintf(w, "Parse error: %v\n", err)
		return 1
	}
	summary := model.Summary()

	fmt.Fprintf(w, "=== UNM Model: %s ===\n", summary.SystemName)
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Entities:")
	fmt.Fprintf(w, "  Actors:       %d\n", summary.ActorCount)
	fmt.Fprintf(w, "  Needs:        %d\n", summary.NeedCount)
	fmt.Fprintf(w, "  Capabilities: %d\n", summary.CapabilityCount)
	fmt.Fprintf(w, "  Services:     %d\n", summary.ServiceCount)
	fmt.Fprintf(w, "  Teams:        %d\n", summary.TeamCount)

	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Analysis:")
	fmt.Fprintf(w, "  Orphan services:         %d\n", summary.OrphanServiceCount)
	fmt.Fprintf(w, "  Fragmented capabilities: %d\n", summary.FragmentedCapCount)
	fmt.Fprintf(w, "  Overloaded teams:        %d\n", summary.OverloadedTeamCount)
	fmt.Fprintln(w, "")

	printValidationResult(w, &result)

	if len(model.Warnings) > 0 {
		fmt.Fprintf(w, "Warnings (%d):\n", len(model.Warnings))
		for _, warn := range model.Warnings {
			fmt.Fprintf(w, "  \u26a0  %s\n", warn)
		}
		fmt.Fprintln(w)
	}

	if !result.IsValid() {
		return 2
	}
	return 0
}

// runValidateCommand validates the given file, printing only the validation result to w.
// Returns: 0 = valid (even with warnings), 1 = parse error, 2 = validation errors.
// With --strict flag, also returns 2 if model.Warnings is non-empty.
func runValidateCommand(args []string, w io.Writer) int {
	fs := flag.NewFlagSet("validate", flag.ContinueOnError)
	fs.SetOutput(w)
	strict := fs.Bool("strict", false, "treat unresolved reference warnings as errors")
	if err := fs.Parse(args); err != nil {
		return 1
	}
	remaining := fs.Args()
	if len(remaining) < 1 {
		fmt.Fprintln(w, "Error: validate requires a file argument")
		return 1
	}
	path := remaining[0]

	f, err := os.Open(path)
	if err != nil {
		fmt.Fprintf(w, "Parse error: %v\n", err)
		return 1
	}
	defer f.Close()

	uc := usecase.NewParseAndValidate(parser.NewParserForPath(path), service.NewValidationEngine())
	model, result, err := uc.Execute(f)
	if err != nil {
		fmt.Fprintf(w, "Parse error: %v\n", err)
		return 1
	}

	printValidationResult(w, &result)

	if !result.IsValid() {
		return 2
	}

	// --strict only promotes unresolved reference warnings, not deprecation notices.
	var refWarnings []string
	for _, warn := range model.Warnings {
		if strings.Contains(warn, "unresolved reference") {
			refWarnings = append(refWarnings, warn)
		}
	}
	if *strict && len(refWarnings) > 0 {
		fmt.Fprintf(w, "Strict mode: %d unresolved reference warning(s) treated as errors:\n", len(refWarnings))
		for _, warn := range refWarnings {
			fmt.Fprintf(w, "  \u26a0  %s\n", warn)
		}
		return 2
	}

	return 0
}

// printValidationResult writes the validation summary section to w.
func printValidationResult(w io.Writer, result *service.ValidationResult) {
	if result.IsValid() {
		fmt.Fprint(w, "Validation: PASSED")
		if result.HasWarnings() {
			fmt.Fprintf(w, "  (with %d warning(s))", len(result.Warnings))
		}
		fmt.Fprintln(w)
	} else {
		fmt.Fprintf(w, "Validation: FAILED with %d error(s)\n", len(result.Errors))
	}

	if result.HasWarnings() {
		fmt.Fprintf(w, "  Warnings: %d\n", len(result.Warnings))
		for _, warn := range result.Warnings {
			fmt.Fprintf(w, "  \u26a0  %s: %s (%s)\n", warn.Code, warn.Message, warn.Entity)
		}
	}

	if !result.IsValid() {
		fmt.Fprintf(w, "  Errors:\n")
		for _, e := range result.Errors {
			fmt.Fprintf(w, "  \u2717  %s: %s (%s)\n", e.Code, e.Message, e.Entity)
		}
	}
}

// runAnalyzeCommand runs one or all analysis types on the given model file.
// Usage: analyze <type> <file>
// Types: fragmentation, cognitive-load, dependencies, gaps, all
// Returns: 0 = success, 1 = usage/parse error.
func runAnalyzeCommand(args []string, w io.Writer, cfg entity.Config) int {
	if len(args) < 2 {
		fmt.Fprintln(w, "Error: analyze requires <type> and <file> arguments")
		fmt.Fprintln(w, "  Types: fragmentation, cognitive-load, dependencies, gaps, all")
		return 1
	}
	analyzeType := args[0]
	path := args[1]

	switch analyzeType {
	case "fragmentation", "cognitive-load", "dependencies", "gaps", "bottleneck", "coupling", "complexity", "interactions", "all":
		// valid
	default:
		fmt.Fprintf(w, "Error: unknown analysis type %q\n", analyzeType)
		fmt.Fprintln(w, "  Types: fragmentation, cognitive-load, dependencies, gaps, bottleneck, coupling, complexity, interactions, all")
		return 1
	}

	f, err := os.Open(path)
	if err != nil {
		fmt.Fprintf(w, "Error: %v\n", err)
		return 1
	}
	defer f.Close()

	uc := usecase.NewParseAndValidate(parser.NewParserForPath(path), service.NewValidationEngine())
	model, _, err := uc.Execute(f)
	if err != nil {
		fmt.Fprintf(w, "Parse error: %v\n", err)
		return 1
	}

	switch analyzeType {
	case "fragmentation":
		printFragmentation(w, analyzer.NewFragmentationAnalyzer().Analyze(model))
	case "cognitive-load":
		printCognitiveLoad(w, analyzer.NewCognitiveLoadAnalyzer(cfg.Analysis.CognitiveLoad, cfg.Analysis.InteractionWeights).Analyze(model))
	case "dependencies":
		printDependencies(w, analyzer.NewDependencyAnalyzer().Analyze(model))
	case "gaps":
		printGaps(w, analyzer.NewGapAnalyzer().Analyze(model))
	case "bottleneck":
		printBottleneck(w, analyzer.NewBottleneckAnalyzer(cfg.Analysis.Bottleneck).Analyze(model))
	case "coupling":
		printCoupling(w, analyzer.NewCouplingAnalyzer().Analyze(model))
	case "complexity":
		printComplexity(w, analyzer.NewComplexityAnalyzer().Analyze(model))
	case "interactions":
		printInteractionDiversity(w, analyzer.NewInteractionDiversityAnalyzer(cfg.Analysis.Signals).Analyze(model))
	case "all":
		fragReport := analyzer.NewFragmentationAnalyzer().Analyze(model)
		cogReport := analyzer.NewCognitiveLoadAnalyzer(cfg.Analysis.CognitiveLoad, cfg.Analysis.InteractionWeights).Analyze(model)
		depReport := analyzer.NewDependencyAnalyzer().Analyze(model)
		gapReport := analyzer.NewGapAnalyzer().Analyze(model)
		bottleReport := analyzer.NewBottleneckAnalyzer(cfg.Analysis.Bottleneck).Analyze(model)
		unlinkedReport := analyzer.NewUnlinkedCapabilityAnalyzer().Analyze(model)
		interactionReport := analyzer.NewInteractionDiversityAnalyzer(cfg.Analysis.Signals).Analyze(model)

		printFragmentation(w, fragReport)
		printCognitiveLoad(w, cogReport)
		printDependencies(w, depReport)
		printGaps(w, gapReport)
		printBottleneck(w, bottleReport)
		printCoupling(w, analyzer.NewCouplingAnalyzer().Analyze(model))
		printComplexity(w, analyzer.NewComplexityAnalyzer().Analyze(model))
		printUnlinkedCapabilities(w, unlinkedReport)
		printInteractionDiversity(w, interactionReport)
		suggestReport := analyzer.NewSignalSuggestionGenerator(cfg.Analysis.Signals).Generate(bottleReport, cogReport, fragReport, depReport, unlinkedReport, model)
		printSignalSuggestions(w, suggestReport)
	}
	return 0
}

func printFragmentation(w io.Writer, report analyzer.FragmentationReport) {
	fmt.Fprintln(w, "=== Fragmentation Analysis ===")
	if len(report.FragmentedCapabilities) == 0 {
		fmt.Fprintln(w, "  No fragmented capabilities detected.")
	} else {
		fmt.Fprintf(w, "  Fragmented capabilities: %d\n", len(report.FragmentedCapabilities))
		for _, fc := range report.FragmentedCapabilities {
			teamNames := make([]string, 0, len(fc.Teams))
			for _, t := range fc.Teams {
				teamNames = append(teamNames, t.Name)
			}
			sort.Strings(teamNames)
			fmt.Fprintf(w, "  \u26a0  %s → owned by %d teams: %v\n", fc.Capability.Name, len(fc.Teams), teamNames)
		}
	}
	fmt.Fprintln(w)
}

func printCognitiveLoad(w io.Writer, report analyzer.CognitiveLoadReport) {
	fmt.Fprintln(w, "=== Structural Load Assessment (Team Topologies) ===")
	if len(report.TeamLoads) == 0 {
		fmt.Fprintln(w, "  No teams found.")
	} else {
		fmt.Fprintf(w, "  %-30s %8s %8s %8s %8s %8s\n", "Team", "Domain", "Service", "Interact", "DepFan", "Overall")
		fmt.Fprintf(w, "  %-30s %8s %8s %8s %8s %8s\n", "----", "------", "-------", "--------", "------", "-------")
		for _, tl := range report.TeamLoads {
			sizeNote := ""
			if !tl.SizeIsExplicit {
				sizeNote = " (size?)"
			}
		fmt.Fprintf(w, "  %-30s %4g/%-3s %4g/%-3s %4g/%-3s %4g/%-3s %7s%s\n",
			tl.Team.Name,
			tl.DomainSpread.Value, tl.DomainSpread.Level,
			tl.ServiceLoad.Value, tl.ServiceLoad.Level,
			tl.InteractionLoad.Value, tl.InteractionLoad.Level,
			tl.DependencyLoad.Value, tl.DependencyLoad.Level,
			tl.OverallLevel, sizeNote)
		}
	}
	fmt.Fprintln(w)
}

func printDependencies(w io.Writer, report analyzer.DependencyReport) {
	fmt.Fprintln(w, "=== Dependencies Analysis ===")
	fmt.Fprintf(w, "  Max service dependency depth:     %d\n", report.MaxServiceDepth)
	fmt.Fprintf(w, "  Max capability dependency depth:  %d\n", report.MaxCapabilityDepth)
	if len(report.CriticalServicePath) > 0 {
		fmt.Fprintf(w, "  Critical service path:  %v\n", report.CriticalServicePath)
	}
	if len(report.ServiceCycles) > 0 {
		fmt.Fprintf(w, "  \u26a0 Service cycles detected: %d\n", len(report.ServiceCycles))
		for _, c := range report.ServiceCycles {
			fmt.Fprintf(w, "    cycle: %v\n", c.Path)
		}
	} else {
		fmt.Fprintln(w, "  No service dependency cycles.")
	}
	if len(report.CapabilityCycles) > 0 {
		fmt.Fprintf(w, "  \u26a0 Capability cycles detected: %d\n", len(report.CapabilityCycles))
		for _, c := range report.CapabilityCycles {
			fmt.Fprintf(w, "    cycle: %v\n", c.Path)
		}
	} else {
		fmt.Fprintln(w, "  No capability dependency cycles.")
	}
	fmt.Fprintln(w)
}

func printGaps(w io.Writer, report analyzer.GapReport) {
	fmt.Fprintln(w, "=== Gaps Analysis ===")
	if len(report.UnmappedNeeds) == 0 && len(report.UnrealizedCapabilities) == 0 &&
		len(report.UnownedServices) == 0 && len(report.UnneededCapabilities) == 0 {
		fmt.Fprintln(w, "  No gaps detected.")
	} else {
		if len(report.UnmappedNeeds) > 0 {
			fmt.Fprintf(w, "  Unmapped needs (%d):\n", len(report.UnmappedNeeds))
			for _, n := range report.UnmappedNeeds {
				fmt.Fprintf(w, "    \u26a0  %s\n", n.Name)
			}
		}
		if len(report.UnrealizedCapabilities) > 0 {
			fmt.Fprintf(w, "  Unrealized capabilities (%d):\n", len(report.UnrealizedCapabilities))
			for _, c := range report.UnrealizedCapabilities {
				fmt.Fprintf(w, "    \u26a0  %s\n", c.Name)
			}
		}
		if len(report.UnownedServices) > 0 {
			fmt.Fprintf(w, "  Unowned services (%d):\n", len(report.UnownedServices))
			for _, s := range report.UnownedServices {
				fmt.Fprintf(w, "    \u26a0  %s\n", s.Name)
			}
		}
		if len(report.UnneededCapabilities) > 0 {
			fmt.Fprintf(w, "  Unneeded capabilities (%d):\n", len(report.UnneededCapabilities))
			for _, c := range report.UnneededCapabilities {
				fmt.Fprintf(w, "    \u26a0  %s\n", c.Name)
			}
		}
		if len(report.OrphanServices) > 0 {
			fmt.Fprintf(w, "  Orphan services (%d):\n", len(report.OrphanServices))
			for _, s := range report.OrphanServices {
				fmt.Fprintf(w, "    ⚠  %s\n", s.Name)
			}
		}
	}
	fmt.Fprintln(w)
}

func printBottleneck(w io.Writer, report analyzer.BottleneckReport) {
	fmt.Fprintln(w, "=== Bottleneck Analysis ===")
	if len(report.ServiceBottlenecks) == 0 {
		fmt.Fprintln(w, "  No services found.")
	} else {
		fmt.Fprintf(w, "  %-35s %6s %6s %s\n", "Service", "FanIn", "FanOut", "Status")
		fmt.Fprintf(w, "  %-35s %6s %6s %s\n", "-------", "-----", "------", "------")
		for _, sb := range report.ServiceBottlenecks {
			status := ""
			if sb.IsCritical {
				status = " \u2757 critical bottleneck"
			} else if sb.IsWarning {
				status = " \u26a0 bottleneck warning"
			}
			fmt.Fprintf(w, "  %-35s %6d %6d%s\n", sb.Service.Name, sb.FanIn, sb.FanOut, status)
		}
	}
	fmt.Fprintln(w)
}

func printCoupling(w io.Writer, report analyzer.CouplingReport) {
	fmt.Fprintln(w, "=== Coupling Analysis ===")
	if len(report.DataAssetCouplings) == 0 {
		fmt.Fprintln(w, "  No data-asset-mediated coupling detected.")
	} else {
		fmt.Fprintf(w, "  Coupled data assets: %d\n", len(report.DataAssetCouplings))
		for _, dac := range report.DataAssetCouplings {
			crossTeam := ""
			if dac.IsCrossteam {
				crossTeam = " \u26a0 cross-team"
			}
			fmt.Fprintf(w, "  %s (%s)%s\n", dac.DataAsset.Name, dac.DataAsset.Type, crossTeam)
			for _, svc := range dac.Services {
				fmt.Fprintf(w, "    \u2192 %s\n", svc)
			}
		}
	}
	fmt.Fprintln(w)
}

func printComplexity(w io.Writer, report analyzer.ComplexityReport) {
	fmt.Fprintln(w, "=== Complexity Analysis ===")
	if len(report.Services) == 0 {
		fmt.Fprintln(w, "  No services found.")
	} else {
		fmt.Fprintf(w, "  %-35s %5s %5s %5s %s\n", "Service", "Deps", "Caps", "Data", "Total")
		fmt.Fprintf(w, "  %-35s %5s %5s %5s %s\n", "-------", "----", "----", "----", "-----")
		for _, sc := range report.Services {
			fmt.Fprintf(w, "  %-35s %5d %5d %5d %5d\n",
				sc.Service.Name, sc.DependencyScore, sc.CapabilityScore,
				sc.DataAssetScore, sc.TotalComplexity)
		}
	}
	fmt.Fprintln(w)
}

func printUnlinkedCapabilities(w io.Writer, report analyzer.UnlinkedCapabilityReport) {
	fmt.Fprintln(w, "=== Unlinked Capability Analysis ===")
	fmt.Fprintf(w, "  Need coverage: %d/%d leaf capabilities linked (%.0f%%)\n",
		report.LinkedCount, report.TotalLeafCapabilityCount, report.LinkedPercentage)
	if len(report.UnlinkedLeafCapabilities) == 0 {
		fmt.Fprintln(w, "  All leaf capabilities are referenced by at least one need.")
	} else {
		suspicious := 0
		for _, uc := range report.UnlinkedLeafCapabilities {
			if !uc.IsExpected {
				suspicious++
			}
		}
		if suspicious > 0 {
			fmt.Fprintf(w, "  \u26a0 %d suspicious unlinked (domain/foundational):\n", suspicious)
			for _, uc := range report.UnlinkedLeafCapabilities {
				if !uc.IsExpected {
					fmt.Fprintf(w, "    \u26a0  %s [%s]\n", uc.Capability.Name, uc.Visibility)
				}
			}
		}
		expected := len(report.UnlinkedLeafCapabilities) - suspicious
		if expected > 0 {
			fmt.Fprintf(w, "  \u2139 %d expected unlinked (infrastructure):\n", expected)
			for _, uc := range report.UnlinkedLeafCapabilities {
				if uc.IsExpected {
					fmt.Fprintf(w, "    \u2013  %s\n", uc.Capability.Name)
				}
			}
		}
	}
	fmt.Fprintln(w)
}

func printInteractionDiversity(w io.Writer, report analyzer.InteractionDiversityReport) {
	fmt.Fprintln(w, "=== Interaction Diversity Analysis ===")
	if len(report.ModeDistribution) == 0 {
		fmt.Fprintln(w, "  No interactions declared.")
	} else {
		fmt.Fprintln(w, "  Mode distribution:")
		for mode, count := range report.ModeDistribution {
			fmt.Fprintf(w, "    %-22s %d\n", string(mode), count)
		}
		if report.AllModesSame {
			fmt.Fprintln(w, "  \u26a0 All interactions use the same mode — Team Topologies model may be incomplete.")
		}
	}
	if len(report.IsolatedTeams) > 0 {
		fmt.Fprintf(w, "  \u26a0 Isolated teams (%d — no interactions):\n", len(report.IsolatedTeams))
		for _, name := range report.IsolatedTeams {
			fmt.Fprintf(w, "    \u2013  %s\n", name)
		}
	}
	if len(report.OverReliantTeams) > 0 {
		fmt.Fprintln(w, "  \u26a0 Over-reliant teams (4+ interactions of same mode):")
		for _, ot := range report.OverReliantTeams {
			fmt.Fprintf(w, "    \u2013  %-30s %s ×%d\n", ot.TeamName, string(ot.Mode), ot.Count)
		}
	}
	fmt.Fprintln(w)
}

func printSignalSuggestions(w io.Writer, report analyzer.SignalSuggestionsReport) {
	fmt.Fprintln(w, "=== Suggested Signals ===")
	if len(report.Suggestions) == 0 {
		fmt.Fprintln(w, "  No signal suggestions generated.")
	} else {
		fmt.Fprintf(w, "  %d candidate signal(s) to consider adding to your model:\n", len(report.Suggestions))
		for i, s := range report.Suggestions {
			fmt.Fprintf(w, "\n  [%d] %s / %s (%s)\n", i+1, s.Category, s.OnEntityName, string(s.Severity))
			fmt.Fprintf(w, "      %s\n", s.Description)
			fmt.Fprintf(w, "      Evidence: %s\n", s.Evidence)
			fmt.Fprintf(w, "      Source: %s\n", s.Source)
		}
	}
	fmt.Fprintln(w)
}

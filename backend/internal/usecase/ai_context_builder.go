package usecase

import (
	"sort"
	"strings"

	"github.com/krzachariassen/unm-platform/internal/domain/entity"
	"github.com/krzachariassen/unm-platform/internal/infrastructure/analyzer"
)

// AIContextBuilder assembles all model data and analysis results into a map
// suitable for rendering AI prompt templates.
type AIContextBuilder struct {
	cognitiveLoad CognitiveLoader
	valueChain    ValueChainAnalyzer
	fragmentation Fragmenter
	dependency    DependencyAnalyzer
	gap           GapAnalyzer
	bottleneck    BottleneckAnalyzer
	coupling      CouplingAnalyzer
	complexity    ComplexityAnalyzer
	interactions  InteractionDiversityAnalyzer
	unlinked      UnlinkedCapabilityAnalyzer
	valueStream   ValueStreamAnalyzer
}

// NewAIContextBuilder constructs an AIContextBuilder.
func NewAIContextBuilder(
	cl CognitiveLoader,
	vc ValueChainAnalyzer,
	frag Fragmenter,
	dep DependencyAnalyzer,
	g GapAnalyzer,
	bn BottleneckAnalyzer,
	cp CouplingAnalyzer,
	cx ComplexityAnalyzer,
	intr InteractionDiversityAnalyzer,
	unl UnlinkedCapabilityAnalyzer,
	vs ValueStreamAnalyzer,
) *AIContextBuilder {
	return &AIContextBuilder{
		cognitiveLoad: cl,
		valueChain:    vc,
		fragmentation: frag,
		dependency:    dep,
		gap:           g,
		bottleneck:    bn,
		coupling:      cp,
		complexity:    cx,
		interactions:  intr,
		unlinked:      unl,
		valueStream:   vs,
	}
}

// BuildPromptData assembles all model data and ALL analysis results into a map
// suitable for rendering AI prompt templates. Every analyzer is run so the AI
// has the complete picture.
func (b *AIContextBuilder) BuildPromptData(m *entity.UNMModel, userQuestion string) (map[string]any, error) {
	// Run ALL analyzers
	clReport := b.cognitiveLoad.Analyze(m)
	vcReport := b.valueChain.Analyze(m)
	fragReport := b.fragmentation.Analyze(m)
	depReport := b.dependency.Analyze(m)
	gapReport := b.gap.Analyze(m)
	bnReport := b.bottleneck.Analyze(m)
	cpReport := b.coupling.Analyze(m)
	cxReport := b.complexity.Analyze(m)
	ixDivReport := b.interactions.Analyze(m)
	unlReport := b.unlinked.Analyze(m)
	vsReport := b.valueStream.Analyze(m)

	// Cognitive load by team for quick lookup
	loadByTeam := make(map[string]string)
	for _, tl := range clReport.TeamLoads {
		loadByTeam[tl.Team.Name] = string(tl.OverallLevel)
	}

	// Value chain by need for quick lookup
	vcByNeed := make(map[string]analyzer.NeedDeliveryRisk)
	for _, ndr := range vcReport.NeedRisks {
		vcByNeed[ndr.NeedName] = ndr
	}

	// Service-to-capabilities reverse index
	svcToCaps := make(map[string][]string)
	for _, cap := range m.Capabilities {
		for _, rel := range cap.RealizedBy {
			svcName := rel.TargetID.String()
			svcToCaps[svcName] = append(svcToCaps[svcName], cap.Name)
		}
	}

	// Teams
	type aiTeamSummary struct {
		Name            string
		TeamType        string
		Size            int
		CognitiveLoad   string
		ServiceCount    int
		CapabilityCount int
		Services        []string
		Capabilities    []string
	}
	teams := make([]aiTeamSummary, 0, len(m.Teams))
	for _, t := range m.Teams {
		var svcNames []string
		for _, svc := range m.Services {
			if svc.OwnerTeamName == t.Name {
				svcNames = append(svcNames, svc.Name)
			}
		}
		capNames := make([]string, 0, len(t.Owns))
		for _, rel := range t.Owns {
			capNames = append(capNames, rel.TargetID.String())
		}
		teams = append(teams, aiTeamSummary{
			Name:            t.Name,
			TeamType:        string(t.TeamType),
			Size:            t.EffectiveSize(),
			CognitiveLoad:   loadByTeam[t.Name],
			ServiceCount:    len(svcNames),
			CapabilityCount: len(capNames),
			Services:        svcNames,
			Capabilities:    capNames,
		})
	}
	sort.Slice(teams, func(i, j int) bool { return teams[i].Name < teams[j].Name })

	// Services
	type aiServiceSummary struct {
		Name            string
		OwnerTeam       string
		DependencyCount int
		DependsOn       []string
		Capabilities    []string
	}
	services := make([]aiServiceSummary, 0, len(m.Services))
	for _, svc := range m.Services {
		depTargets := make([]string, 0, len(svc.DependsOn))
		for _, rel := range svc.DependsOn {
			depTargets = append(depTargets, rel.TargetID.String())
		}
		services = append(services, aiServiceSummary{
			Name:            svc.Name,
			OwnerTeam:       svc.OwnerTeamName,
			DependencyCount: len(svc.DependsOn),
			DependsOn:       depTargets,
			Capabilities:    svcToCaps[svc.Name],
		})
	}
	sort.Slice(services, func(i, j int) bool { return services[i].Name < services[j].Name })

	// Capabilities
	type aiCapSummary struct {
		Name              string
		Visibility        string
		OwnerTeams        []string
		RealizingServices []string
	}
	caps := make([]aiCapSummary, 0, len(m.Capabilities))
	for _, cap := range m.Capabilities {
		ownerTeams := make([]string, 0)
		for _, t := range m.Teams {
			for _, rel := range t.Owns {
				if rel.TargetID.String() == cap.Name {
					ownerTeams = append(ownerTeams, t.Name)
				}
			}
		}
		realizingServices := make([]string, 0, len(cap.RealizedBy))
		for _, rel := range cap.RealizedBy {
			realizingServices = append(realizingServices, rel.TargetID.String())
		}
		caps = append(caps, aiCapSummary{
			Name:              cap.Name,
			Visibility:        cap.Visibility,
			OwnerTeams:        ownerTeams,
			RealizingServices: realizingServices,
		})
	}
	sort.Slice(caps, func(i, j int) bool { return caps[i].Name < caps[j].Name })

	// Needs
	type aiNeedSummary struct {
		Name                    string
		Actor                   string
		SupportedByCapabilities []string
		TeamSpan                int
		AtRisk                  bool
	}
	needs := make([]aiNeedSummary, 0, len(m.Needs))
	for _, n := range m.Needs {
		suppBy := make([]string, 0, len(n.SupportedBy))
		for _, rel := range n.SupportedBy {
			suppBy = append(suppBy, rel.TargetID.String())
		}
		ndr := vcByNeed[n.Name]
		needs = append(needs, aiNeedSummary{
			Name:                    n.Name,
			Actor:                   strings.Join(n.ActorNames, ", "),
			SupportedByCapabilities: suppBy,
			TeamSpan:                ndr.TeamSpan,
			AtRisk:                  ndr.AtRisk,
		})
	}
	sort.Slice(needs, func(i, j int) bool { return needs[i].Name < needs[j].Name })

	// Interactions
	type aiInteractionSummary struct {
		From        string
		To          string
		Mode        string
		Via         string
		Description string
	}
	interactions := make([]aiInteractionSummary, 0, len(m.Interactions))
	for _, ix := range m.Interactions {
		interactions = append(interactions, aiInteractionSummary{
			From:        ix.FromTeamName,
			To:          ix.ToTeamName,
			Mode:        string(ix.Mode),
			Via:         ix.Via,
			Description: ix.Description,
		})
	}

	// Signals
	type aiSignalSummary struct {
		Category         string
		Severity         string
		Description      string
		AffectedEntities string
		Evidence         string
	}
	signals := make([]aiSignalSummary, 0, len(m.Signals))
	for _, s := range m.Signals {
		affectedStr := strings.Join(s.AffectedEntities, ", ")
		signals = append(signals, aiSignalSummary{
			Category:         string(s.Category),
			Severity:         string(s.Severity),
			Description:      s.Description,
			AffectedEntities: affectedStr,
			Evidence:         s.Evidence,
		})
	}

	// Value chains
	type valueChainEntry struct {
		Actor        string
		Need         string
		Capabilities []string
		Services     []string
		Teams        []string
	}
	var valueChains []valueChainEntry
	for _, n := range m.Needs {
		vc := valueChainEntry{Actor: strings.Join(n.ActorNames, ", "), Need: n.Name}
		teamSet := make(map[string]bool)
		for _, rel := range n.SupportedBy {
			capName := rel.TargetID.String()
			vc.Capabilities = append(vc.Capabilities, capName)
			for _, cap := range m.Capabilities {
				if cap.Name == capName {
					for _, rRel := range cap.RealizedBy {
						svcName := rRel.TargetID.String()
						vc.Services = append(vc.Services, svcName)
						for _, svc := range m.Services {
							if svc.Name == svcName {
								teamSet[svc.OwnerTeamName] = true
							}
						}
					}
				}
			}
		}
		for t := range teamSet {
			vc.Teams = append(vc.Teams, t)
		}
		valueChains = append(valueChains, vc)
	}

	// Fragmented capabilities
	type aiFragmentedCap struct {
		Name       string
		Teams      []string
		Visibility string
	}
	var fragmentedCaps []aiFragmentedCap
	for _, fc := range fragReport.FragmentedCapabilities {
		teamNames := make([]string, 0, len(fc.Teams))
		for _, t := range fc.Teams {
			teamNames = append(teamNames, t.Name)
		}
		fragmentedCaps = append(fragmentedCaps, aiFragmentedCap{
			Name:       fc.Capability.Name,
			Teams:      teamNames,
			Visibility: fc.Capability.Visibility,
		})
	}

	// Cognitive load detail per team
	type cogLoadDetail struct {
		Team               string
		OverallLevel       string
		DomainSpread       string
		DomainSpreadVal    float64
		ServiceLoad        string
		ServiceLoadVal     float64
		InteractionLoad    string
		InteractionLoadVal float64
		DependencyLoad     string
		DependencyLoadVal  float64
		ServiceCount       int
		CapabilityCount    int
		TeamSize           int
	}
	var cogLoadDetails []cogLoadDetail
	for _, tl := range clReport.TeamLoads {
		cogLoadDetails = append(cogLoadDetails, cogLoadDetail{
			Team:               tl.Team.Name,
			OverallLevel:       string(tl.OverallLevel),
			DomainSpread:       string(tl.DomainSpread.Level),
			DomainSpreadVal:    tl.DomainSpread.Value,
			ServiceLoad:        string(tl.ServiceLoad.Level),
			ServiceLoadVal:     tl.ServiceLoad.Value,
			InteractionLoad:    string(tl.InteractionLoad.Level),
			InteractionLoadVal: tl.InteractionLoad.Value,
			DependencyLoad:     string(tl.DependencyLoad.Level),
			DependencyLoadVal:  tl.DependencyLoad.Value,
			ServiceCount:       tl.ServiceCount,
			CapabilityCount:    tl.CapabilityCount,
			TeamSize:           tl.TeamSize,
		})
	}

	// Bottleneck analysis
	type aiBottleneckSummary struct {
		Service    string
		FanIn      int
		FanOut     int
		IsCritical bool
	}
	bottlenecks := make([]aiBottleneckSummary, 0)
	for _, b := range bnReport.ServiceBottlenecks {
		if b.IsCritical || b.IsWarning {
			bottlenecks = append(bottlenecks, aiBottleneckSummary{
				Service:    b.Service.Name,
				FanIn:      b.FanIn,
				FanOut:     b.FanOut,
				IsCritical: b.IsCritical,
			})
		}
	}

	// Coupling analysis
	type aiCouplingSummary struct {
		DataAsset   string
		Type        string
		Services    []string
		IsCrossteam bool
	}
	couplings := make([]aiCouplingSummary, 0, len(cpReport.DataAssetCouplings))
	for _, c := range cpReport.DataAssetCouplings {
		assetType := ""
		if c.DataAsset != nil {
			assetType = c.DataAsset.Type
		}
		couplings = append(couplings, aiCouplingSummary{
			DataAsset:   c.DataAsset.Name,
			Type:        assetType,
			Services:    c.Services,
			IsCrossteam: c.IsCrossteam,
		})
	}

	// Gap analysis
	type aiGapSummary struct {
		UnmappedNeeds          []string
		UnrealizedCapabilities []string
		UnownedServices        []string
		UnneededCapabilities   []string
	}
	gaps := aiGapSummary{}
	for _, n := range gapReport.UnmappedNeeds {
		gaps.UnmappedNeeds = append(gaps.UnmappedNeeds, n.Name+" (actor: "+strings.Join(n.ActorNames, ", ")+")")
	}
	for _, c := range gapReport.UnrealizedCapabilities {
		gaps.UnrealizedCapabilities = append(gaps.UnrealizedCapabilities, c.Name)
	}
	for _, s := range gapReport.UnownedServices {
		gaps.UnownedServices = append(gaps.UnownedServices, s.Name)
	}
	for _, c := range gapReport.UnneededCapabilities {
		gaps.UnneededCapabilities = append(gaps.UnneededCapabilities, c.Name)
	}

	// Complexity analysis
	type aiComplexitySummary struct {
		Service         string
		DependencyScore int
		CapabilityScore int
		DataAssetScore  int
		TotalComplexity int
	}
	complexities := make([]aiComplexitySummary, 0, len(cxReport.Services))
	for _, s := range cxReport.Services {
		complexities = append(complexities, aiComplexitySummary{
			Service:         s.Service.Name,
			DependencyScore: s.DependencyScore,
			CapabilityScore: s.CapabilityScore,
			DataAssetScore:  s.DataAssetScore,
			TotalComplexity: s.TotalComplexity,
		})
	}

	// Dependency analysis
	svcCycles := make([][]string, 0)
	for _, c := range depReport.ServiceCycles {
		svcCycles = append(svcCycles, c.Path)
	}
	capCycles := make([][]string, 0)
	for _, c := range depReport.CapabilityCycles {
		capCycles = append(capCycles, c.Path)
	}
	critPath := depReport.CriticalServicePath
	if critPath == nil {
		critPath = []string{}
	}

	// Value stream coherence
	type aiValueStreamSummary struct {
		TeamName       string
		NeedsServed    []string
		CoherenceScore float64
		LowCoherence   bool
	}
	valueStreams := make([]aiValueStreamSummary, 0, len(vsReport.TeamCoherences))
	for _, tc := range vsReport.TeamCoherences {
		valueStreams = append(valueStreams, aiValueStreamSummary{
			TeamName:       tc.TeamName,
			NeedsServed:    tc.NeedsServed,
			CoherenceScore: tc.CoherenceScore,
			LowCoherence:   tc.LowCoherence,
		})
	}

	// Interaction diversity
	modeDist := make(map[string]int, len(ixDivReport.ModeDistribution))
	for mode, count := range ixDivReport.ModeDistribution {
		modeDist[string(mode)] = count
	}
	type aiOverReliantTeam struct {
		TeamName string
		Mode     string
		Count    int
	}
	overReliant := make([]aiOverReliantTeam, 0, len(ixDivReport.OverReliantTeams))
	for _, or_ := range ixDivReport.OverReliantTeams {
		overReliant = append(overReliant, aiOverReliantTeam{
			TeamName: or_.TeamName,
			Mode:     string(or_.Mode),
			Count:    or_.Count,
		})
	}
	isolatedTeams := ixDivReport.IsolatedTeams
	if isolatedTeams == nil {
		isolatedTeams = []string{}
	}

	// Unlinked capabilities
	type aiUnlinkedCapSummary struct {
		Name       string
		Visibility string
	}
	unlinkedCaps := make([]aiUnlinkedCapSummary, 0)
	for _, uc := range unlReport.UnlinkedLeafCapabilities {
		if !uc.IsExpected {
			unlinkedCaps = append(unlinkedCaps, aiUnlinkedCapSummary{
				Name:       uc.Capability.Name,
				Visibility: uc.Visibility,
			})
		}
	}

	// External dependencies
	type aiExternalDepSummary struct {
		Name   string
		UsedBy []string
	}
	externalDeps := make([]aiExternalDepSummary, 0, len(m.ExternalDependencies))
	for _, ed := range m.ExternalDependencies {
		usedBy := make([]string, 0, len(ed.UsedBy))
		for _, u := range ed.UsedBy {
			usedBy = append(usedBy, u.ServiceName)
		}
		externalDeps = append(externalDeps, aiExternalDepSummary{
			Name:   ed.Name,
			UsedBy: usedBy,
		})
	}

	// Value chain aggregate counts
	type vcCounts struct {
		CrossTeamNeeds int
		AtRiskNeeds    int
		UnbackedNeeds  int
	}
	vcAgg := vcCounts{
		CrossTeamNeeds: vcReport.CrossTeamNeedCount,
		AtRiskNeeds:    vcReport.AtRiskNeedCount,
		UnbackedNeeds:  vcReport.UnbackedNeedCount,
	}

	return map[string]any{
		"SystemName":             m.System.Name,
		"ModelDescription":       m.System.Description,
		"Teams":                  teams,
		"Services":               services,
		"Capabilities":           caps,
		"Needs":                  needs,
		"Interactions":           interactions,
		"Signals":                signals,
		"ValueChains":            valueChains,
		"FragmentedCapabilities": fragmentedCaps,
		"CognitiveLoadDetails":   cogLoadDetails,
		"UserQuestion":           userQuestion,
		"Question":               userQuestion,
		"Bottlenecks":            bottlenecks,
		"Couplings":              couplings,
		"Gaps":                   gaps,
		"Complexities":           complexities,
		"ServiceCycles":          svcCycles,
		"CapabilityCycles":       capCycles,
		"CriticalServicePath":    critPath,
		"MaxServiceDepth":        depReport.MaxServiceDepth,
		"MaxCapabilityDepth":     depReport.MaxCapabilityDepth,
		"ValueStreamCoherence":   valueStreams,
		"LowCoherenceCount":      vsReport.LowCoherenceCount,
		"ModeDistribution":       modeDist,
		"OverReliantTeams":       overReliant,
		"IsolatedTeams":          isolatedTeams,
		"AllModesSame":           ixDivReport.AllModesSame,
		"UnlinkedCapabilities":   unlinkedCaps,
		"ExternalDependencies":   externalDeps,
		"ValueChainCounts":       vcAgg,
	}, nil
}

package service

import (
	"fmt"

	"github.com/krzachariassen/unm-platform/internal/domain/entity"
)

// ── Severity levels ───────────────────────────────────────────────────────────

// Severity classifies a validation finding by urgency.
type Severity string

const (
	SeverityError   Severity = "error"   // Blocks model validity
	SeverityWarning Severity = "warning" // Non-blocking advisory
	SeverityInfo    Severity = "info"    // Informational diagnostic (orphan entities, etc.)
)

// ── Error codes ───────────────────────────────────────────────────────────────

// ValidationErrorCode identifies a specific validation error rule.
type ValidationErrorCode string

const (
	ErrNeedNoCapability   ValidationErrorCode = "need-no-capability"
	ErrLeafCapNoService   ValidationErrorCode = "leaf-cap-no-service"
	ErrServiceNoOwner     ValidationErrorCode = "service-no-owner"
	ErrInvalidInteraction ValidationErrorCode = "invalid-interaction"
	ErrInvalidConfidence  ValidationErrorCode = "invalid-confidence"
)

// ValidationWarningCode identifies a specific validation warning rule.
type ValidationWarningCode string

const (
	WarnFragmentation             ValidationWarningCode = "fragmentation"
	WarnCognitiveLoad             ValidationWarningCode = "cognitive-load"
	WarnOrphanService             ValidationWarningCode = "orphan-service"
	WarnCircularDep               ValidationWarningCode = "circular-dependency"
	WarnLowConfidence             ValidationWarningCode = "low-confidence"
	WarnParentCapHasServices      ValidationWarningCode = "parent-cap-has-services"
	WarnNonPlatformTeamInPlatform ValidationWarningCode = "non-platform-team-in-platform"
	WarnSelfDependency            ValidationWarningCode = "self-dependency"
	WarnTeamSizeUnset             ValidationWarningCode = "team-size-unset"

	// Info-level diagnostics for orphaned entities.
	InfoOrphanActor ValidationWarningCode = "info-orphan-actor"
	InfoOrphanTeam  ValidationWarningCode = "info-orphan-team"
)

// ── Result types ──────────────────────────────────────────────────────────────

// ValidationError describes a rule violation that blocks model validity.
type ValidationError struct {
	Code     ValidationErrorCode
	Severity Severity
	Message  string
	Entity   string // name of the entity causing the error
}

// ValidationWarning describes a non-blocking advisory about the model.
type ValidationWarning struct {
	Code     ValidationWarningCode
	Severity Severity
	Message  string
	Entity   string
}

// ValidationResult aggregates all errors and warnings produced by validation.
type ValidationResult struct {
	Errors   []ValidationError
	Warnings []ValidationWarning
}

// IsValid returns true when the model has no validation errors.
func (r *ValidationResult) IsValid() bool { return len(r.Errors) == 0 }

// HasWarnings returns true when the model has at least one warning.
func (r *ValidationResult) HasWarnings() bool { return len(r.Warnings) > 0 }

// ── Engine ────────────────────────────────────────────────────────────────────

// ValidationEngine runs the full set of UNM validation rules against a model.
type ValidationEngine struct{}

// NewValidationEngine constructs a ValidationEngine.
func NewValidationEngine() *ValidationEngine {
	return &ValidationEngine{}
}

// Validate executes all validation rules and returns a ValidationResult.
func (v *ValidationEngine) Validate(model *entity.UNMModel) ValidationResult {
	var result ValidationResult

	v.checkNeeds(model, &result)
	v.checkCapabilities(model, &result)
	v.checkServices(model, &result)
	v.checkInteractions(model, &result)
	v.checkInferredMappings(model, &result)
	v.checkFragmentation(model, &result)
	v.checkCognitiveLoad(model, &result)
	v.checkTeamSizes(model, &result)
	v.checkSelfDependencies(model, &result)
	v.checkCircularDeps(model, &result)
	v.checkPlatforms(model, &result)
	v.checkOrphanActors(model, &result)
	v.checkOrphanTeams(model, &result)

	return result
}

// addError appends an error-severity validation error.
func addError(result *ValidationResult, code ValidationErrorCode, entity, message string) {
	result.Errors = append(result.Errors, ValidationError{
		Code:     code,
		Severity: SeverityError,
		Message:  message,
		Entity:   entity,
	})
}

// addWarning appends a warning-severity validation warning.
func addWarning(result *ValidationResult, code ValidationWarningCode, entity, message string) {
	result.Warnings = append(result.Warnings, ValidationWarning{
		Code:     code,
		Severity: SeverityWarning,
		Message:  message,
		Entity:   entity,
	})
}

// addInfo appends an info-severity diagnostic as a ValidationWarning.
func addInfo(result *ValidationResult, code ValidationWarningCode, entity, message string) {
	result.Warnings = append(result.Warnings, ValidationWarning{
		Code:     code,
		Severity: SeverityInfo,
		Message:  message,
		Entity:   entity,
	})
}

// ── Error rules ───────────────────────────────────────────────────────────────

// checkNeeds enforces ErrNeedNoCapability.
func (v *ValidationEngine) checkNeeds(model *entity.UNMModel, result *ValidationResult) {
	for _, need := range model.Needs {
		if len(need.SupportedBy) == 0 {
			addError(result, ErrNeedNoCapability,
				need.Name,
				fmt.Sprintf("need %q has no supporting capabilities", need.Name))
		}
	}
}

// checkCapabilities enforces ErrLeafCapNoService and WarnParentCapHasServices.
func (v *ValidationEngine) checkCapabilities(model *entity.UNMModel, result *ValidationResult) {
	// Build set of realized capability names from service.Realizes (canonical source).
	realizedCaps := make(map[string]bool)
	parentCapsWithServices := make(map[string]bool)
	for _, svc := range model.Services {
		for _, rel := range svc.Realizes {
			capName := rel.TargetID.String()
			realizedCaps[capName] = true
			if cap, ok := model.Capabilities[capName]; ok && !cap.IsLeaf() {
				parentCapsWithServices[capName] = true
			}
		}
	}
	for _, cap := range model.Capabilities {
		if cap.IsLeaf() && !realizedCaps[cap.Name] {
			addError(result, ErrLeafCapNoService,
				cap.Name,
				fmt.Sprintf("leaf capability %q is not realized by any service", cap.Name))
		}
		// Warn if a non-leaf capability is realized by services (only leaf caps should realize services).
		if !cap.IsLeaf() && parentCapsWithServices[cap.Name] {
			addWarning(result, WarnParentCapHasServices,
				cap.Name,
				fmt.Sprintf("non-leaf capability %q has services realizing it; only leaf capabilities should realize services", cap.Name))
		}
	}
}

// checkServices enforces ErrServiceNoOwner and generates WarnOrphanService.
func (v *ValidationEngine) checkServices(model *entity.UNMModel, result *ValidationResult) {
	for _, svc := range model.Services {
		if svc.OwnerTeamName == "" {
			addError(result, ErrServiceNoOwner,
				svc.Name,
				fmt.Sprintf("service %q has no owner team", svc.Name))
		}
		// Orphan check: service not realizing any capability.
		if len(model.GetCapabilitiesForService(svc.Name)) == 0 {
			addWarning(result, WarnOrphanService,
				svc.Name,
				fmt.Sprintf("service %q supports no capabilities", svc.Name))
		}
	}
}

// checkInteractions enforces ErrInvalidInteraction.
// An interaction is invalid if:
//   - FromTeamName or ToTeamName does not reference a known team, OR
//   - Mode is empty.
func (v *ValidationEngine) checkInteractions(model *entity.UNMModel, result *ValidationResult) {
	for _, interaction := range model.Interactions {
		entityLabel := fmt.Sprintf("%s→%s", interaction.FromTeamName, interaction.ToTeamName)

		_, fromExists := model.Teams[interaction.FromTeamName]
		_, toExists := model.Teams[interaction.ToTeamName]

		if !fromExists || !toExists || interaction.Mode == "" {
			addError(result, ErrInvalidInteraction,
				entityLabel,
				fmt.Sprintf("interaction %q is invalid (unknown team or empty mode)", entityLabel))
		}
	}
}

// checkInferredMappings enforces ErrInvalidConfidence and generates WarnLowConfidence.
func (v *ValidationEngine) checkInferredMappings(model *entity.UNMModel, result *ValidationResult) {
	for _, im := range model.InferredMappings {
		entityLabel := fmt.Sprintf("%s→%s", im.ServiceName, im.CapabilityName)

		score := im.Confidence.Score
		if score < 0.0 || score > 1.0 {
			addError(result, ErrInvalidConfidence,
				entityLabel,
				fmt.Sprintf("inferred mapping %q has confidence score %v outside [0.0, 1.0]", entityLabel, score))
		} else if im.IsLowConfidence() {
			addWarning(result, WarnLowConfidence,
				entityLabel,
				fmt.Sprintf("inferred mapping %q has low confidence score %v", entityLabel, score))
		}
	}
}

// ── Warning rules ─────────────────────────────────────────────────────────────

// checkFragmentation generates WarnFragmentation for capabilities owned by more than 2 teams.
func (v *ValidationEngine) checkFragmentation(model *entity.UNMModel, result *ValidationResult) {
	for capName, cap := range model.Capabilities {
		teams := model.GetTeamsForCapability(capName)
		if len(teams) > 2 {
			addWarning(result, WarnFragmentation,
				cap.Name,
				fmt.Sprintf("capability %q is owned by %d teams (fragmentation)", cap.Name, len(teams)))
		}
	}
}

// checkCognitiveLoad generates WarnCognitiveLoad for teams owning more than 6 capabilities.
func (v *ValidationEngine) checkCognitiveLoad(model *entity.UNMModel, result *ValidationResult) {
	for _, team := range model.Teams {
		caps := model.GetCapabilitiesForTeam(team.Name)
		if len(caps) > 6 {
			addWarning(result, WarnCognitiveLoad,
				team.Name,
				fmt.Sprintf("team %q owns %d capabilities (cognitive load)", team.Name, len(caps)))
		}
	}
}

// checkTeamSizes generates WarnTeamSizeUnset for any team whose Size is not set (== 0).
// Without an explicit size, cognitive load percentages fall back to a default of 5 people
// which may be inaccurate. This warning prompts the user to add size: N to each team.
func (v *ValidationEngine) checkTeamSizes(model *entity.UNMModel, result *ValidationResult) {
	for _, team := range model.Teams {
		if !team.SizeExplicit {
			addWarning(result, WarnTeamSizeUnset,
				team.Name,
				fmt.Sprintf("team %q has no size set — structural load uses default of 5 people; add 'size: N' to the team definition", team.Name))
		}
	}
}

// checkSelfDependencies detects services or capabilities that list themselves
// in their own DependsOn, and emits WarnSelfDependency for each occurrence.
func (v *ValidationEngine) checkSelfDependencies(model *entity.UNMModel, result *ValidationResult) {
	for _, svc := range model.Services {
		for _, rel := range svc.DependsOn {
			if rel.TargetID.String() == svc.Name {
				addWarning(result, WarnSelfDependency,
					svc.Name,
					fmt.Sprintf("service %q depends on itself", svc.Name))
				break // only emit once per service
			}
		}
	}
	for capName, cap := range model.Capabilities {
		for _, rel := range cap.DependsOn {
			if rel.TargetID.String() == capName {
				addWarning(result, WarnSelfDependency,
					cap.Name,
					fmt.Sprintf("capability %q depends on itself", cap.Name))
				break // only emit once per capability
			}
		}
	}
}

// checkCircularDeps detects cycles in the capability dependsOn graph using iterative DFS.
// It reports WarnCircularDep for any capability that is part of a cycle.
// Self-loops are excluded here — they are handled by checkSelfDependencies.
func (v *ValidationEngine) checkCircularDeps(model *entity.UNMModel, result *ValidationResult) {
	// Build adjacency list: capName → []depCapName (skip self-loops)
	adj := make(map[string][]string, len(model.Capabilities))
	for capName, cap := range model.Capabilities {
		var deps []string
		for _, rel := range cap.DependsOn {
			targetName := rel.TargetID.String()
			if targetName == capName {
				// self-loop — skip; handled by checkSelfDependencies
				continue
			}
			if _, exists := model.Capabilities[targetName]; exists {
				deps = append(deps, targetName)
			}
		}
		adj[capName] = deps
	}

	// DFS state: 0 = unvisited, 1 = in-stack, 2 = done
	const (
		unvisited = 0
		inStack   = 1
		done      = 2
	)
	state := make(map[string]int, len(model.Capabilities))
	reported := make(map[string]bool)

	var dfs func(node string) bool
	dfs = func(node string) bool {
		state[node] = inStack
		for _, neighbor := range adj[node] {
			switch state[neighbor] {
			case inStack:
				// back edge → cycle detected
				if !reported[neighbor] {
					reported[neighbor] = true
					addWarning(result, WarnCircularDep,
						neighbor,
						fmt.Sprintf("capability %q is part of a circular dependency", neighbor))
				}
				return true
			case unvisited:
				if dfs(neighbor) {
					// also flag the current node if not already reported
					if !reported[node] {
						reported[node] = true
						addWarning(result, WarnCircularDep,
							node,
							fmt.Sprintf("capability %q is part of a circular dependency", node))
					}
				}
			}
		}
		state[node] = done
		return false
	}

	for capName := range model.Capabilities {
		if state[capName] == unvisited {
			dfs(capName)
		}
	}
}

// checkPlatforms generates WarnNonPlatformTeamInPlatform for platforms that contain
// teams that are not of type "platform".
func (v *ValidationEngine) checkPlatforms(model *entity.UNMModel, result *ValidationResult) {
	for _, platform := range model.Platforms {
		for _, teamName := range platform.TeamNames {
			team, exists := model.Teams[teamName]
			if !exists {
				continue
			}
			if team.TeamType.String() != "platform" {
				addWarning(result, WarnNonPlatformTeamInPlatform,
					platform.Name,
					fmt.Sprintf("platform %q contains team %q which is not of type platform (got %q)", platform.Name, teamName, team.TeamType.String()))
			}
		}
	}
}

// ── Info diagnostics ──────────────────────────────────────────────────────────

// checkOrphanActors emits an info-level diagnostic for actors that have no needs.
func (v *ValidationEngine) checkOrphanActors(model *entity.UNMModel, result *ValidationResult) {
	// Build set of actor names referenced by needs
	referenced := make(map[string]bool)
	for _, need := range model.Needs {
		for _, actorName := range need.ActorNames {
			referenced[actorName] = true
		}
	}
	for _, actor := range model.Actors {
		if !referenced[actor.Name] {
			addInfo(result, InfoOrphanActor,
				actor.Name,
				fmt.Sprintf("actor %q has no needs — consider adding needs or removing the actor", actor.Name))
		}
	}
}

// checkOrphanTeams emits an info-level diagnostic for teams that own no services
// and no capabilities.
func (v *ValidationEngine) checkOrphanTeams(model *entity.UNMModel, result *ValidationResult) {
	for _, team := range model.Teams {
		caps := model.GetCapabilitiesForTeam(team.Name)

		// Count services owned by this team
		svcCount := 0
		for _, svc := range model.Services {
			if svc.OwnerTeamName == team.Name {
				svcCount++
			}
		}

		if len(caps) == 0 && svcCount == 0 {
			addInfo(result, InfoOrphanTeam,
				team.Name,
				fmt.Sprintf("team %q owns no services or capabilities — it may be incomplete", team.Name))
		}
	}
}

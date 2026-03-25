package service

import (
	"fmt"

	"github.com/krzachariassen/unm-platform/internal/domain/entity"
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
)

// ── Result types ──────────────────────────────────────────────────────────────

// ValidationError describes a rule violation that blocks model validity.
type ValidationError struct {
	Code    ValidationErrorCode
	Message string
	Entity  string // name of the entity causing the error
}

// ValidationWarning describes a non-blocking advisory about the model.
type ValidationWarning struct {
	Code    ValidationWarningCode
	Message string
	Entity  string
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

	return result
}

// ── Error rules ───────────────────────────────────────────────────────────────

// checkNeeds enforces ErrNeedNoCapability.
func (v *ValidationEngine) checkNeeds(model *entity.UNMModel, result *ValidationResult) {
	for _, need := range model.Needs {
		if len(need.SupportedBy) == 0 {
			result.Errors = append(result.Errors, ValidationError{
				Code:    ErrNeedNoCapability,
				Message: fmt.Sprintf("need %q has no supporting capabilities", need.Name),
				Entity:  need.Name,
			})
		}
	}
}

// checkCapabilities enforces ErrLeafCapNoService and WarnParentCapHasServices.
func (v *ValidationEngine) checkCapabilities(model *entity.UNMModel, result *ValidationResult) {
	for _, cap := range model.Capabilities {
		if cap.IsLeaf() && len(cap.RealizedBy) == 0 {
			result.Errors = append(result.Errors, ValidationError{
				Code:    ErrLeafCapNoService,
				Message: fmt.Sprintf("leaf capability %q is not realized by any service", cap.Name),
				Entity:  cap.Name,
			})
		}
		// Warn if a non-leaf capability also has RealizedBy entries (only leaf caps should have services).
		if !cap.IsLeaf() && len(cap.RealizedBy) > 0 {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Code:    WarnParentCapHasServices,
				Message: fmt.Sprintf("non-leaf capability %q has RealizedBy entries; only leaf capabilities should realize services", cap.Name),
				Entity:  cap.Name,
			})
		}
	}
}

// checkServices enforces ErrServiceNoOwner and generates WarnOrphanService.
func (v *ValidationEngine) checkServices(model *entity.UNMModel, result *ValidationResult) {
	for _, svc := range model.Services {
		if svc.OwnerTeamName == "" {
			result.Errors = append(result.Errors, ValidationError{
				Code:    ErrServiceNoOwner,
				Message: fmt.Sprintf("service %q has no owner team", svc.Name),
				Entity:  svc.Name,
			})
		}
		// Orphan check: service not referenced in any capability's RealizedBy.
		if len(model.GetCapabilitiesForService(svc.Name)) == 0 {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Code:    WarnOrphanService,
				Message: fmt.Sprintf("service %q supports no capabilities", svc.Name),
				Entity:  svc.Name,
			})
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
			result.Errors = append(result.Errors, ValidationError{
				Code:    ErrInvalidInteraction,
				Message: fmt.Sprintf("interaction %q is invalid (unknown team or empty mode)", entityLabel),
				Entity:  entityLabel,
			})
		}
	}
}

// checkInferredMappings enforces ErrInvalidConfidence and generates WarnLowConfidence.
func (v *ValidationEngine) checkInferredMappings(model *entity.UNMModel, result *ValidationResult) {
	for _, im := range model.InferredMappings {
		entityLabel := fmt.Sprintf("%s→%s", im.ServiceName, im.CapabilityName)

		score := im.Confidence.Score
		if score < 0.0 || score > 1.0 {
			result.Errors = append(result.Errors, ValidationError{
				Code:    ErrInvalidConfidence,
				Message: fmt.Sprintf("inferred mapping %q has confidence score %v outside [0.0, 1.0]", entityLabel, score),
				Entity:  entityLabel,
			})
		} else if im.IsLowConfidence() {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Code:    WarnLowConfidence,
				Message: fmt.Sprintf("inferred mapping %q has low confidence score %v", entityLabel, score),
				Entity:  entityLabel,
			})
		}
	}
}

// ── Warning rules ─────────────────────────────────────────────────────────────

// checkFragmentation generates WarnFragmentation for capabilities owned by more than 2 teams.
func (v *ValidationEngine) checkFragmentation(model *entity.UNMModel, result *ValidationResult) {
	for capName, cap := range model.Capabilities {
		teams := model.GetTeamsForCapability(capName)
		if len(teams) > 2 {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Code:    WarnFragmentation,
				Message: fmt.Sprintf("capability %q is owned by %d teams (fragmentation)", cap.Name, len(teams)),
				Entity:  cap.Name,
			})
		}
	}
}

// checkCognitiveLoad generates WarnCognitiveLoad for teams owning more than 6 capabilities.
func (v *ValidationEngine) checkCognitiveLoad(model *entity.UNMModel, result *ValidationResult) {
	for _, team := range model.Teams {
		caps := model.GetCapabilitiesForTeam(team.Name)
		if len(caps) > 6 {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Code:    WarnCognitiveLoad,
				Message: fmt.Sprintf("team %q owns %d capabilities (cognitive load)", team.Name, len(caps)),
				Entity:  team.Name,
			})
		}
	}
}

// checkTeamSizes generates WarnTeamSizeUnset for any team whose Size is not set (== 0).
// Without an explicit size, cognitive load percentages fall back to a default of 5 people
// which may be inaccurate. This warning prompts the user to add size: N to each team.
func (v *ValidationEngine) checkTeamSizes(model *entity.UNMModel, result *ValidationResult) {
	for _, team := range model.Teams {
		if !team.SizeExplicit {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Code:    WarnTeamSizeUnset,
				Message: fmt.Sprintf("team %q has no size set — structural load uses default of 5 people; add 'size: N' to the team definition", team.Name),
				Entity:  team.Name,
			})
		}
	}
}

// checkSelfDependencies detects services or capabilities that list themselves
// in their own DependsOn, and emits WarnSelfDependency for each occurrence.
func (v *ValidationEngine) checkSelfDependencies(model *entity.UNMModel, result *ValidationResult) {
	for _, svc := range model.Services {
		for _, rel := range svc.DependsOn {
			if rel.TargetID.String() == svc.Name {
				result.Warnings = append(result.Warnings, ValidationWarning{
					Code:    WarnSelfDependency,
					Message: fmt.Sprintf("service %q depends on itself", svc.Name),
					Entity:  svc.Name,
				})
				break // only emit once per service
			}
		}
	}
	for capName, cap := range model.Capabilities {
		for _, rel := range cap.DependsOn {
			if rel.TargetID.String() == capName {
				result.Warnings = append(result.Warnings, ValidationWarning{
					Code:    WarnSelfDependency,
					Message: fmt.Sprintf("capability %q depends on itself", cap.Name),
					Entity:  cap.Name,
				})
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
					result.Warnings = append(result.Warnings, ValidationWarning{
						Code:    WarnCircularDep,
						Message: fmt.Sprintf("capability %q is part of a circular dependency", neighbor),
						Entity:  neighbor,
					})
				}
				return true
			case unvisited:
				if dfs(neighbor) {
					// also flag the current node if not already reported
					if !reported[node] {
						reported[node] = true
						result.Warnings = append(result.Warnings, ValidationWarning{
							Code:    WarnCircularDep,
							Message: fmt.Sprintf("capability %q is part of a circular dependency", node),
							Entity:  node,
						})
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
				result.Warnings = append(result.Warnings, ValidationWarning{
					Code:    WarnNonPlatformTeamInPlatform,
					Message: fmt.Sprintf("platform %q contains team %q which is not of type platform (got %q)", platform.Name, teamName, team.TeamType.String()),
					Entity:  platform.Name,
				})
			}
		}
	}
}

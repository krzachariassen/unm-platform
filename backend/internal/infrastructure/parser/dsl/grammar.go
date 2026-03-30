package dsl

import (
	"fmt"
	"strings"
	"unicode"
)

// Parse parses UNM DSL source text and returns the AST File.
// Returns *ParseError for any parse failure (implements error).
func Parse(src string) (*File, error) {
	p := &parser{src: src, pos: 0, line: 1}
	f, err := p.parseFile()
	if err != nil {
		// Propagate *ParseError directly; wrap plain errors.
		if pe, ok := err.(*ParseError); ok {
			return nil, pe
		}
		return nil, &ParseError{Line: p.line, Message: err.Error()}
	}
	return f, nil
}

type parser struct {
	src  string
	pos  int
	line int
}

// errorf creates a *ParseError at the current line.
func (p *parser) errorf(format string, args ...any) *ParseError {
	return &ParseError{Line: p.line, Message: fmt.Sprintf(format, args...)}
}

// ---------------------------------------------------------------------------
// Top-level
// ---------------------------------------------------------------------------

func (p *parser) parseFile() (*File, error) {
	f := &File{}
	for {
		tok := p.peekToken()
		if tok == "" {
			break
		}
		switch tok {
		case "system":
			p.readToken()
			node, err := p.parseSystem()
			if err != nil {
				return nil, err
			}
			f.System = node
		case "actor":
			p.readToken()
			node, err := p.parseActor()
			if err != nil {
				return nil, err
			}
			f.Actors = append(f.Actors, node)
		case "need":
			p.readToken()
			node, err := p.parseNeed()
			if err != nil {
				return nil, err
			}
			f.Needs = append(f.Needs, node)
		case "capability":
			p.readToken()
			node, err := p.parseCapability()
			if err != nil {
				return nil, err
			}
			f.Capabilities = append(f.Capabilities, node)
		case "service":
			p.readToken()
			node, err := p.parseService()
			if err != nil {
				return nil, err
			}
			f.Services = append(f.Services, node)
		case "team":
			p.readToken()
			node, err := p.parseTeam()
			if err != nil {
				return nil, err
			}
			f.Teams = append(f.Teams, node)
		case "platform":
			p.readToken()
			node, err := p.parsePlatform()
			if err != nil {
				return nil, err
			}
			f.Platforms = append(f.Platforms, node)
		case "interaction":
			p.readToken()
			node, err := p.parseInteraction()
			if err != nil {
				return nil, err
			}
			f.Interactions = append(f.Interactions, node)
		case "data_asset", "data":
			p.readToken()
			node, err := p.parseDataAsset()
			if err != nil {
				return nil, err
			}
			f.DataAssets = append(f.DataAssets, node)
		case "external_dependency", "external":
			p.readToken()
			node, err := p.parseExternalDependency()
			if err != nil {
				return nil, err
			}
			f.ExternalDependencies = append(f.ExternalDependencies, node)
		case "signal":
			p.readToken()
			node, err := p.parseSignal()
			if err != nil {
				return nil, err
			}
			f.Signals = append(f.Signals, node)
		case "import":
			p.readToken()
			node, err := p.parseImport()
			if err != nil {
				return nil, err
			}
			f.Imports = append(f.Imports, node)
		case "inferred":
			p.readToken()
			node, err := p.parseInferredMapping()
			if err != nil {
				return nil, err
			}
			f.InferredMappings = append(f.InferredMappings, node)
		case "transition":
			p.readToken()
			node, err := p.parseTransition()
			if err != nil {
				return nil, err
			}
			f.Transitions = append(f.Transitions, node)
		default:
			p.readToken() // consume to advance p.line to the line of the unknown token
			return nil, p.errorf("unknown top-level keyword %q", tok)
		}
	}
	return f, nil
}

// ---------------------------------------------------------------------------
// Block parsers
// ---------------------------------------------------------------------------

func (p *parser) parseSystem() (*SystemNode, error) {
	name, err := p.readString()
	if err != nil {
		return nil, p.errorf("system: %s", err.Error())
	}
	node := &SystemNode{Name: name}
	if err := p.expect("{"); err != nil {
		return nil, p.errorf("system %q: %s", name, err.Error())
	}
	for {
		tok := p.peekToken()
		if tok == "}" || tok == "" {
			break
		}
		switch tok {
		case "description":
			p.readToken()
			v, err := p.readString()
			if err != nil {
				return nil, p.errorf("system description: %s", err.Error())
			}
			node.Description = v
		default:
			p.readToken() // consume to advance p.line to the token's line
			return nil, p.errorf("system: unexpected field %q", tok)
		}
	}
	if err := p.expect("}"); err != nil {
		return nil, p.errorf("system %q: %s", name, err.Error())
	}
	return node, nil
}

func (p *parser) parseActor() (*ActorNode, error) {
	name, err := p.readString()
	if err != nil {
		return nil, p.errorf("actor: %s", err.Error())
	}
	node := &ActorNode{Name: name}
	if err := p.expect("{"); err != nil {
		return nil, p.errorf("actor %q: %s", name, err.Error())
	}
	for {
		tok := p.peekToken()
		if tok == "}" || tok == "" {
			break
		}
		switch tok {
		case "description":
			p.readToken()
			v, err := p.readString()
			if err != nil {
				return nil, p.errorf("actor description: %s", err.Error())
			}
			node.Description = v
		default:
			p.readToken() // consume to advance p.line to the token's line
			return nil, p.errorf("actor: unexpected field %q", tok)
		}
	}
	if err := p.expect("}"); err != nil {
		return nil, p.errorf("actor %q: %s", name, err.Error())
	}
	return node, nil
}

func (p *parser) parseNeed() (*NeedNode, error) {
	name, err := p.readString()
	if err != nil {
		return nil, p.errorf("need: %s", err.Error())
	}
	node := &NeedNode{Name: name}
	if err := p.expect("{"); err != nil {
		return nil, p.errorf("need %q: %s", name, err.Error())
	}
	for {
		tok := p.peekToken()
		if tok == "}" || tok == "" {
			break
		}
		switch tok {
		case "description":
			p.readToken()
			v, err := p.readString()
			if err != nil {
				return nil, p.errorf("need description: %s", err.Error())
			}
			node.Description = v
		case "outcome":
			p.readToken()
			v, err := p.readString()
			if err != nil {
				return nil, p.errorf("need outcome: %s", err.Error())
			}
			node.Outcome = v
		case "actor":
			p.readToken()
			v, err := p.readString()
			if err != nil {
				return nil, p.errorf("need actor: %s", err.Error())
			}
			actors := []string{v}
			// Support comma-separated actors: actor "A", "B"
			for p.peekToken() == "," {
				p.readToken() // consume ","
				next, err := p.readString()
				if err != nil {
					return nil, p.errorf("need actor: %s", err.Error())
				}
				actors = append(actors, next)
			}
			node.Actors = actors
		case "supportedBy":
			p.readToken()
			rel, err := p.parseRelationship()
			if err != nil {
				return nil, p.errorf("need supportedBy: %s", err.Error())
			}
			node.SupportedBy = append(node.SupportedBy, rel)
		default:
			p.readToken()
			return nil, p.errorf("need: unexpected field %q", tok)
		}
	}
	if err := p.expect("}"); err != nil {
		return nil, p.errorf("need %q: %s", name, err.Error())
	}
	return node, nil
}

func (p *parser) parseCapability() (*CapabilityNode, error) {
	name, err := p.readString()
	if err != nil {
		return nil, p.errorf("capability: %s", err.Error())
	}
	node := &CapabilityNode{Name: name}
	if err := p.expect("{"); err != nil {
		return nil, p.errorf("capability %q: %s", name, err.Error())
	}
	for {
		tok := p.peekToken()
		if tok == "}" || tok == "" {
			break
		}
		switch tok {
		case "description":
			p.readToken()
			v, err := p.readString()
			if err != nil {
				return nil, p.errorf("capability description: %s", err.Error())
			}
			node.Description = v
		case "visibility":
			p.readToken()
			v, err := p.readString()
			if err != nil {
				return nil, p.errorf("capability visibility: %s", err.Error())
			}
			node.Visibility = v
		case "parent":
			p.readToken()
			v, err := p.readString()
			if err != nil {
				return nil, p.errorf("capability parent: %s", err.Error())
			}
			node.Parent = v
		case "realizedBy":
			p.readToken()
			rel, err := p.parseRelationship()
			if err != nil {
				return nil, p.errorf("capability realizedBy: %s", err.Error())
			}
			node.RealizedBy = append(node.RealizedBy, rel)
		case "dependsOn":
			p.readToken()
			rel, err := p.parseRelationship()
			if err != nil {
				return nil, p.errorf("capability dependsOn: %s", err.Error())
			}
			node.DependsOn = append(node.DependsOn, rel)
		case "capability":
			p.readToken()
			child, err := p.parseCapability()
			if err != nil {
				return nil, err
			}
			node.Children = append(node.Children, child)
		default:
			p.readToken()
			return nil, p.errorf("capability: unexpected field %q", tok)
		}
	}
	if err := p.expect("}"); err != nil {
		return nil, p.errorf("capability %q: %s", name, err.Error())
	}
	return node, nil
}

func (p *parser) parseService() (*ServiceNode, error) {
	name, err := p.readString()
	if err != nil {
		return nil, p.errorf("service: %s", err.Error())
	}
	node := &ServiceNode{Name: name}
	if err := p.expect("{"); err != nil {
		return nil, p.errorf("service %q: %s", name, err.Error())
	}
	for {
		tok := p.peekToken()
		if tok == "}" || tok == "" {
			break
		}
		switch tok {
		case "description":
			p.readToken()
			v, err := p.readString()
			if err != nil {
				return nil, p.errorf("service description: %s", err.Error())
			}
			node.Description = v
		case "ownedBy":
			p.readToken()
			v, err := p.readString()
			if err != nil {
				return nil, p.errorf("service ownedBy: %s", err.Error())
			}
			node.OwnedBy = v
		case "dependsOn":
			p.readToken()
			rel, err := p.parseRelationship()
			if err != nil {
				return nil, p.errorf("service dependsOn: %s", err.Error())
			}
			node.DependsOn = append(node.DependsOn, rel)
		case "realizes":
			p.readToken()
			target, err := p.readString()
			if err != nil {
				return nil, p.errorf("service realizes: %s", err.Error())
			}
			r := ServiceRealizesNode{Target: target}
			if p.peekToken() == "role" {
				p.readToken() // consume "role"
				role, err := p.readString()
				if err != nil {
					return nil, p.errorf("service realizes role: %s", err.Error())
				}
				r.Role = role
			}
			node.Realizes = append(node.Realizes, r)
		case "externalDeps":
			p.readToken()
			v, err := p.readString()
			if err != nil {
				return nil, p.errorf("service externalDeps: %s", err.Error())
			}
			node.ExternalDeps = append(node.ExternalDeps, v)
		default:
			p.readToken()
			return nil, p.errorf("service: unexpected field %q", tok)
		}
	}
	if err := p.expect("}"); err != nil {
		return nil, p.errorf("service %q: %s", name, err.Error())
	}
	return node, nil
}

func (p *parser) parseTeam() (*TeamNode, error) {
	name, err := p.readString()
	if err != nil {
		return nil, p.errorf("team: %s", err.Error())
	}
	node := &TeamNode{Name: name}
	if err := p.expect("{"); err != nil {
		return nil, p.errorf("team %q: %s", name, err.Error())
	}
	for {
		tok := p.peekToken()
		if tok == "}" || tok == "" {
			break
		}
		switch tok {
		case "description":
			p.readToken()
			v, err := p.readString()
			if err != nil {
				return nil, p.errorf("team description: %s", err.Error())
			}
			node.Description = v
		case "type":
			p.readToken()
			v, err := p.readString()
			if err != nil {
				return nil, p.errorf("team type: %s", err.Error())
			}
			node.Type = v
		case "size":
			p.readToken()
			v, err := p.readInt()
			if err != nil {
				return nil, p.errorf("team size: %s", err.Error())
			}
			node.Size = v
		case "owns":
			p.readToken()
			v, err := p.readString()
			if err != nil {
				return nil, p.errorf("team owns: %s", err.Error())
			}
			node.Owns = append(node.Owns, v)
		case "interacts":
			p.readToken()
			inter, err := p.parseTeamInteraction()
			if err != nil {
				return nil, p.errorf("team interacts: %s", err.Error())
			}
			node.Interacts = append(node.Interacts, inter)
		default:
			p.readToken()
			return nil, p.errorf("team: unexpected field %q", tok)
		}
	}
	if err := p.expect("}"); err != nil {
		return nil, p.errorf("team %q: %s", name, err.Error())
	}
	return node, nil
}

func (p *parser) parsePlatform() (*PlatformNode, error) {
	name, err := p.readString()
	if err != nil {
		return nil, p.errorf("platform: %s", err.Error())
	}
	node := &PlatformNode{Name: name}
	if err := p.expect("{"); err != nil {
		return nil, p.errorf("platform %q: %s", name, err.Error())
	}
	for {
		tok := p.peekToken()
		if tok == "}" || tok == "" {
			break
		}
		switch tok {
		case "description":
			p.readToken()
			v, err := p.readString()
			if err != nil {
				return nil, p.errorf("platform description: %s", err.Error())
			}
			node.Description = v
		case "teams":
			p.readToken()
			teams, err := p.parseStringList()
			if err != nil {
				return nil, p.errorf("platform teams: %s", err.Error())
			}
			node.Teams = append(node.Teams, teams...)
		default:
			p.readToken()
			return nil, p.errorf("platform: unexpected field %q", tok)
		}
	}
	if err := p.expect("}"); err != nil {
		return nil, p.errorf("platform %q: %s", name, err.Error())
	}
	return node, nil
}

// parseTeamInteraction parses an inline interaction inside a team block:
//
//	interacts "team-b" mode x-as-a-service via "API" description "..."
func (p *parser) parseTeamInteraction() (TeamInteractionNode, error) {
	with, err := p.readString()
	if err != nil {
		return TeamInteractionNode{}, fmt.Errorf("interacts target: %s", err.Error())
	}
	inter := TeamInteractionNode{With: with}
	// Optional inline modifiers: mode <val> via <val> description <val>
	for {
		tok := p.peekToken()
		switch tok {
		case "mode":
			p.readToken()
			v, err := p.readString()
			if err != nil {
				return TeamInteractionNode{}, fmt.Errorf("interacts mode: %s", err.Error())
			}
			inter.Mode = v
		case "via":
			p.readToken()
			v, err := p.readString()
			if err != nil {
				return TeamInteractionNode{}, fmt.Errorf("interacts via: %s", err.Error())
			}
			inter.Via = v
		case "description":
			p.readToken()
			v, err := p.readString()
			if err != nil {
				return TeamInteractionNode{}, fmt.Errorf("interacts description: %s", err.Error())
			}
			inter.Description = v
		default:
			// End of inline modifiers
			return inter, nil
		}
	}
}

func (p *parser) parseInteraction() (*InteractionNode, error) {
	from, err := p.readString()
	if err != nil {
		return nil, p.errorf("interaction from: %s", err.Error())
	}
	if err := p.expect("->"); err != nil {
		return nil, p.errorf("interaction: %s", err.Error())
	}
	to, err := p.readString()
	if err != nil {
		return nil, p.errorf("interaction to: %s", err.Error())
	}
	node := &InteractionNode{From: from, To: to}
	if err := p.expect("{"); err != nil {
		return nil, p.errorf("interaction %q->%q: %s", from, to, err.Error())
	}
	for {
		tok := p.peekToken()
		if tok == "}" || tok == "" {
			break
		}
		switch tok {
		case "description":
			p.readToken()
			v, err := p.readString()
			if err != nil {
				return nil, p.errorf("interaction description: %s", err.Error())
			}
			node.Description = v
		case "mode":
			p.readToken()
			v, err := p.readString()
			if err != nil {
				return nil, p.errorf("interaction mode: %s", err.Error())
			}
			node.Mode = v
		case "via":
			p.readToken()
			v, err := p.readString()
			if err != nil {
				return nil, p.errorf("interaction via: %s", err.Error())
			}
			node.Via = v
		default:
			p.readToken()
			return nil, p.errorf("interaction: unexpected field %q", tok)
		}
	}
	if err := p.expect("}"); err != nil {
		return nil, p.errorf("interaction %q->%q: %s", from, to, err.Error())
	}
	return node, nil
}

func (p *parser) parseDataAsset() (*DataAssetNode, error) {
	name, err := p.readString()
	if err != nil {
		return nil, p.errorf("data_asset: %s", err.Error())
	}
	node := &DataAssetNode{Name: name}
	if err := p.expect("{"); err != nil {
		return nil, p.errorf("data_asset %q: %s", name, err.Error())
	}
	for {
		tok := p.peekToken()
		if tok == "}" || tok == "" {
			break
		}
		switch tok {
		case "description":
			p.readToken()
			v, err := p.readString()
			if err != nil {
				return nil, p.errorf("data_asset description: %s", err.Error())
			}
			node.Description = v
		case "type":
			p.readToken()
			v, err := p.readString()
			if err != nil {
				return nil, p.errorf("data_asset type: %s", err.Error())
			}
			node.Type = v
		case "usedBy":
			p.readToken()
			target, err := p.readString()
			if err != nil {
				return nil, p.errorf("data_asset usedBy: %s", err.Error())
			}
			usage := DataAssetUsageNode{Target: target}
			if p.peekToken() == "access" {
				p.readToken()
				access, err := p.readString()
				if err != nil {
					return nil, p.errorf("data_asset usedBy access: %s", err.Error())
				}
				usage.Access = access
			}
			node.UsedBy = append(node.UsedBy, usage)
		case "producedBy":
			p.readToken()
			v, err := p.readString()
			if err != nil {
				return nil, p.errorf("data_asset producedBy: %s", err.Error())
			}
			node.ProducedBy = v
		case "consumedBy":
			p.readToken()
			v, err := p.readString()
			if err != nil {
				return nil, p.errorf("data_asset consumedBy: %s", err.Error())
			}
			node.ConsumedBy = append(node.ConsumedBy, v)
		default:
			p.readToken()
			return nil, p.errorf("data_asset: unexpected field %q", tok)
		}
	}
	if err := p.expect("}"); err != nil {
		return nil, p.errorf("data_asset %q: %s", name, err.Error())
	}
	return node, nil
}

func (p *parser) parseExternalDependency() (*ExternalDependencyNode, error) {
	name, err := p.readString()
	if err != nil {
		return nil, p.errorf("external_dependency: %s", err.Error())
	}
	node := &ExternalDependencyNode{Name: name}
	if err := p.expect("{"); err != nil {
		return nil, p.errorf("external_dependency %q: %s", name, err.Error())
	}
	for {
		tok := p.peekToken()
		if tok == "}" || tok == "" {
			break
		}
		switch tok {
		case "description":
			p.readToken()
			v, err := p.readString()
			if err != nil {
				return nil, p.errorf("external_dependency description: %s", err.Error())
			}
			node.Description = v
		case "usedBy":
			p.readToken()
			target, err := p.readString()
			if err != nil {
				return nil, p.errorf("external_dependency usedBy: %s", err.Error())
			}
			usage := ExternalDepUsageNode{Target: target}
			if p.peekToken() == ":" {
				p.readToken() // consume ":"
				desc, err := p.readString()
				if err != nil {
					return nil, p.errorf("external_dependency usedBy description: %s", err.Error())
				}
				usage.Description = desc
			}
			node.UsedBy = append(node.UsedBy, usage)
		default:
			p.readToken()
			return nil, p.errorf("external_dependency: unexpected field %q", tok)
		}
	}
	if err := p.expect("}"); err != nil {
		return nil, p.errorf("external_dependency %q: %s", name, err.Error())
	}
	return node, nil
}

func (p *parser) parseSignal() (*SignalNode, error) {
	name, err := p.readString()
	if err != nil {
		return nil, p.errorf("signal: %s", err.Error())
	}
	node := &SignalNode{Name: name}
	if err := p.expect("{"); err != nil {
		return nil, p.errorf("signal %q: %s", name, err.Error())
	}
	for {
		tok := p.peekToken()
		if tok == "}" || tok == "" {
			break
		}
		switch tok {
		case "description":
			p.readToken()
			v, err := p.readString()
			if err != nil {
				return nil, p.errorf("signal description: %s", err.Error())
			}
			node.Description = v
		case "category":
			p.readToken()
			v, err := p.readString()
			if err != nil {
				return nil, p.errorf("signal category: %s", err.Error())
			}
			node.Category = v
		case "severity":
			p.readToken()
			v, err := p.readString()
			if err != nil {
				return nil, p.errorf("signal severity: %s", err.Error())
			}
			node.Severity = v
		case "onEntity":
			p.readToken()
			v, err := p.readString()
			if err != nil {
				return nil, p.errorf("signal onEntity: %s", err.Error())
			}
			node.OnEntity = v
		case "affects":
			p.readToken()
			v, err := p.readString()
			if err != nil {
				return nil, p.errorf("signal affects: %s", err.Error())
			}
			node.Affected = append(node.Affected, v)
		default:
			p.readToken()
			return nil, p.errorf("signal: unexpected field %q", tok)
		}
	}
	if err := p.expect("}"); err != nil {
		return nil, p.errorf("signal %q: %s", name, err.Error())
	}
	return node, nil
}

// ---------------------------------------------------------------------------
// Import (5.5)
// ---------------------------------------------------------------------------

// parseImport parses:
//
//	import "path.unm"
//	import alias from "path.unm"
func (p *parser) parseImport() (*ImportNode, error) {
	tok := p.peekToken()
	if tok == "" {
		return nil, p.errorf("import: unexpected end of input")
	}

	// Check if it's a quoted path (simple import) or a bare identifier (named import)
	if strings.HasPrefix(tok, "\"") {
		// Simple import: import "path.unm"
		path, err := p.readString()
		if err != nil {
			return nil, p.errorf("import: %s", err.Error())
		}
		if path == "" {
			return nil, p.errorf("import: path must not be empty")
		}
		return &ImportNode{Path: path}, nil
	}

	// Named import: import alias from "path.unm"
	alias := p.readToken()
	if alias == "" {
		return nil, p.errorf("import: expected alias or path, got end of input")
	}
	if err := p.expect("from"); err != nil {
		return nil, p.errorf("import %q: expected \"from\", got %q", alias, p.peekToken())
	}
	path, err := p.readString()
	if err != nil {
		return nil, p.errorf("import %q from: %s", alias, err.Error())
	}
	if path == "" {
		return nil, p.errorf("import %q from: path must not be empty", alias)
	}
	return &ImportNode{Alias: alias, Path: path}, nil
}

// ---------------------------------------------------------------------------
// Inferred Mapping (5.6)
// ---------------------------------------------------------------------------

// parseInferredMapping parses:
//
//	inferred {
//	    from "need-name"
//	    to "capability-name"
//	    confidence 0.85
//	    evidence "..."
//	    status suggested
//	}
func (p *parser) parseInferredMapping() (*InferredMappingNode, error) {
	if err := p.expect("{"); err != nil {
		return nil, p.errorf("inferred: %s", err.Error())
	}
	node := &InferredMappingNode{}
	for {
		tok := p.peekToken()
		if tok == "}" || tok == "" {
			break
		}
		switch tok {
		case "from":
			p.readToken()
			v, err := p.readString()
			if err != nil {
				return nil, p.errorf("inferred from: %s", err.Error())
			}
			node.From = v
		case "to":
			p.readToken()
			v, err := p.readString()
			if err != nil {
				return nil, p.errorf("inferred to: %s", err.Error())
			}
			node.To = v
		case "confidence":
			p.readToken()
			v, err := p.readConfidence()
			if err != nil {
				return nil, p.errorf("inferred confidence: %s", err.Error())
			}
			node.Confidence = v
		case "evidence":
			p.readToken()
			v, err := p.readString()
			if err != nil {
				return nil, p.errorf("inferred evidence: %s", err.Error())
			}
			node.Evidence = v
		case "status":
			p.readToken()
			v, err := p.readString()
			if err != nil {
				return nil, p.errorf("inferred status: %s", err.Error())
			}
			node.Status = v
		default:
			p.readToken()
			return nil, p.errorf("inferred: unexpected field %q", tok)
		}
	}
	if err := p.expect("}"); err != nil {
		return nil, p.errorf("inferred: %s", err.Error())
	}
	return node, nil
}

// readConfidence reads the next bare token and parses it as a float64.
func (p *parser) readConfidence() (float64, error) {
	tok := p.readToken()
	if tok == "" {
		return 0, fmt.Errorf("unexpected end of input, expected confidence value")
	}
	var val float64
	_, err := fmt.Sscanf(tok, "%f", &val)
	if err != nil {
		return 0, fmt.Errorf("invalid confidence value %q: %s", tok, err.Error())
	}
	return val, nil
}

// readInt reads the next bare token and parses it as an int.
func (p *parser) readInt() (int, error) {
	tok := p.readToken()
	if tok == "" {
		return 0, fmt.Errorf("unexpected end of input, expected integer value")
	}
	var val int
	_, err := fmt.Sscanf(tok, "%d", &val)
	if err != nil {
		return 0, fmt.Errorf("invalid integer value %q: %s", tok, err.Error())
	}
	return val, nil
}

// ---------------------------------------------------------------------------
// Transition (5.7)
// ---------------------------------------------------------------------------

// parseTransition parses:
//
//	transition "name" {
//	    description "..."
//	    current { capability "X" ownedBy team "Y" ... }
//	    target  { capability "X" ownedBy team "Y" ... }
//	    step N "label" { action ... expected_outcome "..." }
//	}
func (p *parser) parseTransition() (*TransitionNode, error) {
	name, err := p.readString()
	if err != nil {
		return nil, p.errorf("transition: %s", err.Error())
	}
	node := &TransitionNode{Name: name}
	if err := p.expect("{"); err != nil {
		return nil, p.errorf("transition %q: %s", name, err.Error())
	}
	for {
		tok := p.peekToken()
		if tok == "}" || tok == "" {
			break
		}
		switch tok {
		case "description":
			p.readToken()
			v, err := p.readString()
			if err != nil {
				return nil, p.errorf("transition description: %s", err.Error())
			}
			node.Description = v
		case "current":
			p.readToken()
			bindings, err := p.parseTransitionBindings()
			if err != nil {
				return nil, p.errorf("transition current: %s", err.Error())
			}
			node.Current = bindings
		case "target":
			p.readToken()
			bindings, err := p.parseTransitionBindings()
			if err != nil {
				return nil, p.errorf("transition target: %s", err.Error())
			}
			node.Target = bindings
		case "step":
			p.readToken()
			step, err := p.parseTransitionStep()
			if err != nil {
				return nil, p.errorf("transition step: %s", err.Error())
			}
			node.Steps = append(node.Steps, step)
		default:
			p.readToken()
			return nil, p.errorf("transition: unexpected field %q", tok)
		}
	}
	if err := p.expect("}"); err != nil {
		return nil, p.errorf("transition %q: %s", name, err.Error())
	}
	return node, nil
}

// parseTransitionBindings parses:
//
//	{ capability "X" ownedBy team "Y" ... }
func (p *parser) parseTransitionBindings() ([]TransitionBindingNode, error) {
	if err := p.expect("{"); err != nil {
		return nil, err
	}
	var bindings []TransitionBindingNode
	for {
		tok := p.peekToken()
		if tok == "}" || tok == "" {
			break
		}
		if tok != "capability" {
			p.readToken()
			return nil, fmt.Errorf("expected \"capability\", got %q", tok)
		}
		p.readToken() // consume "capability"
		capName, err := p.readString()
		if err != nil {
			return nil, fmt.Errorf("capability name: %s", err.Error())
		}
		if err := p.expect("ownedBy"); err != nil {
			return nil, err
		}
		if err := p.expect("team"); err != nil {
			return nil, err
		}
		teamName, err := p.readString()
		if err != nil {
			return nil, fmt.Errorf("team name: %s", err.Error())
		}
		bindings = append(bindings, TransitionBindingNode{
			CapabilityName: capName,
			TeamName:       teamName,
		})
	}
	if err := p.expect("}"); err != nil {
		return nil, err
	}
	return bindings, nil
}

// parseTransitionStep parses:
//
//	N "label" { action ... expected_outcome "..." }
func (p *parser) parseTransitionStep() (TransitionStepNode, error) {
	// Read step number as a bare token
	numTok := p.readToken()
	if numTok == "" {
		return TransitionStepNode{}, fmt.Errorf("expected step number")
	}
	var num int
	_, err := fmt.Sscanf(numTok, "%d", &num)
	if err != nil {
		return TransitionStepNode{}, fmt.Errorf("invalid step number %q: %s", numTok, err.Error())
	}

	label, err := p.readString()
	if err != nil {
		return TransitionStepNode{}, fmt.Errorf("step label: %s", err.Error())
	}

	step := TransitionStepNode{Number: num, Label: label}
	if err := p.expect("{"); err != nil {
		return TransitionStepNode{}, err
	}
	for {
		tok := p.peekToken()
		if tok == "}" || tok == "" {
			break
		}
		switch tok {
		case "action":
			p.readToken()
			actionText := p.readRestOfLine()
			step.ActionText = actionText
		case "expected_outcome":
			p.readToken()
			v, err := p.readString()
			if err != nil {
				return TransitionStepNode{}, fmt.Errorf("expected_outcome: %s", err.Error())
			}
			step.ExpectedOutcome = v
		default:
			p.readToken()
			return TransitionStepNode{}, fmt.Errorf("unexpected field %q in step", tok)
		}
	}
	if err := p.expect("}"); err != nil {
		return TransitionStepNode{}, err
	}
	return step, nil
}

// readRestOfLine reads the remainder of the current line (trimmed).
func (p *parser) readRestOfLine() string {
	// Skip horizontal whitespace only
	for p.pos < len(p.src) && (p.src[p.pos] == ' ' || p.src[p.pos] == '\t') {
		p.pos++
	}
	start := p.pos
	for p.pos < len(p.src) && p.src[p.pos] != '\n' {
		p.pos++
	}
	return strings.TrimSpace(p.src[start:p.pos])
}

// ---------------------------------------------------------------------------
// Relationship
// ---------------------------------------------------------------------------

// parseRelationship reads: <target> [ : "description" | { description "..." role <role> } ]
func (p *parser) parseRelationship() (RelationshipNode, error) {
	target, err := p.readString()
	if err != nil {
		return RelationshipNode{}, p.errorf("relationship target: %s", err.Error())
	}
	rel := RelationshipNode{Target: target}

	// Colon shorthand: target : "description"
	if p.peekToken() == ":" {
		p.readToken() // consume ":"
		desc, err := p.readString()
		if err != nil {
			return RelationshipNode{}, p.errorf("relationship description: %s", err.Error())
		}
		rel.Description = desc
		return rel, nil
	}

	// Block form: target { description "..." role <role> }
	if p.peekToken() == "{" {
		p.readToken() // consume "{"
		for {
			tok := p.peekToken()
			if tok == "}" || tok == "" {
				break
			}
			switch tok {
			case "description":
				p.readToken()
				v, err := p.readString()
				if err != nil {
					return RelationshipNode{}, p.errorf("relationship description: %s", err.Error())
				}
				rel.Description = v
			case "role":
				p.readToken()
				v, err := p.readString()
				if err != nil {
					return RelationshipNode{}, p.errorf("relationship role: %s", err.Error())
				}
				rel.Role = v
			default:
				p.readToken()
				return RelationshipNode{}, p.errorf("relationship modifier: unexpected field %q", tok)
			}
		}
		if err := p.expect("}"); err != nil {
			return RelationshipNode{}, p.errorf("relationship: %s", err.Error())
		}
	}
	return rel, nil
}

// ---------------------------------------------------------------------------
// String list  ["a", "b", "c"]
// ---------------------------------------------------------------------------

func (p *parser) parseStringList() ([]string, error) {
	if err := p.expect("["); err != nil {
		return nil, p.errorf("string list: %s", err.Error())
	}
	var result []string
	for {
		tok := p.peekToken()
		if tok == "]" || tok == "" {
			break
		}
		// Skip commas between items
		if tok == "," {
			p.readToken()
			continue
		}
		v, err := p.readString()
		if err != nil {
			return nil, p.errorf("string list item: %s", err.Error())
		}
		result = append(result, v)
	}
	if err := p.expect("]"); err != nil {
		return nil, p.errorf("string list: %s", err.Error())
	}
	return result, nil
}

// ---------------------------------------------------------------------------
// Low-level token operations
// ---------------------------------------------------------------------------

// skipWhitespaceAndComments advances pos past whitespace and // comments.
func (p *parser) skipWhitespaceAndComments() {
	for p.pos < len(p.src) {
		ch := p.src[p.pos]
		if ch == '\n' {
			p.line++
			p.pos++
		} else if ch == ' ' || ch == '\t' || ch == '\r' {
			p.pos++
		} else if p.pos+1 < len(p.src) && p.src[p.pos] == '/' && p.src[p.pos+1] == '/' {
			// Skip to end of line
			for p.pos < len(p.src) && p.src[p.pos] != '\n' {
				p.pos++
			}
		} else {
			break
		}
	}
}

// peekToken returns the next token without consuming it.
func (p *parser) peekToken() string {
	saved := p.pos
	savedLine := p.line
	tok := p.readToken()
	p.pos = saved
	p.line = savedLine
	return tok
}

// readToken reads the next token (skipping whitespace/comments).
// Returns "" at end of input.
func (p *parser) readToken() string {
	p.skipWhitespaceAndComments()
	if p.pos >= len(p.src) {
		return ""
	}

	ch := p.src[p.pos]

	// Quoted string: return content between quotes (but keep as token for dispatch)
	if ch == '"' {
		start := p.pos
		p.pos++ // opening quote
		for p.pos < len(p.src) && p.src[p.pos] != '"' {
			if p.src[p.pos] == '\n' {
				p.line++
			}
			p.pos++
		}
		if p.pos < len(p.src) {
			p.pos++ // closing quote
		}
		return p.src[start:p.pos]
	}

	// Two-char token: ->
	if p.pos+1 < len(p.src) && p.src[p.pos] == '-' && p.src[p.pos+1] == '>' {
		p.pos += 2
		return "->"
	}

	// Single-char tokens
	if ch == '{' || ch == '}' || ch == '[' || ch == ']' || ch == ',' || ch == ':' {
		p.pos++
		return string(ch)
	}

	// Bare identifier/keyword: letters, digits, underscore, hyphen, dot, slash
	start := p.pos
	for p.pos < len(p.src) {
		c := rune(p.src[p.pos])
		if unicode.IsLetter(c) || unicode.IsDigit(c) || c == '_' || c == '-' || c == '.' || c == '/' {
			p.pos++
		} else {
			break
		}
	}
	if p.pos == start {
		// Unknown character — skip it to avoid infinite loop
		p.pos++
		return string(ch)
	}
	return p.src[start:p.pos]
}

// expect reads the next token and returns an error if it does not match want.
func (p *parser) expect(want string) error {
	tok := p.readToken()
	if tok != want {
		return fmt.Errorf("expected %q, got %q", want, tok)
	}
	return nil
}

// readString reads the next token and interprets it as a string value.
// Quoted tokens strip surrounding quotes; bare tokens are returned as-is.
func (p *parser) readString() (string, error) {
	tok := p.readToken()
	if tok == "" {
		return "", fmt.Errorf("unexpected end of input, expected string")
	}
	if strings.HasPrefix(tok, "\"") && strings.HasSuffix(tok, "\"") && len(tok) >= 2 {
		return tok[1 : len(tok)-1], nil
	}
	return tok, nil
}

package parser

import (
	"fmt"
	"io"

	"github.com/uber/unm-platform/internal/domain/entity"
	"github.com/uber/unm-platform/internal/infrastructure/parser/dsl"
)

// DSLParser implements the Parser interface for .unm DSL files.
type DSLParser struct{}

// NewDSLParser constructs a DSLParser.
func NewDSLParser() *DSLParser { return &DSLParser{} }

// Parse reads a .unm DSL file and returns a UNMModel.
func (p *DSLParser) Parse(r io.Reader) (*entity.UNMModel, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("dsl parser: read: %w", err)
	}
	ast, err := dsl.Parse(string(data))
	if err != nil {
		return nil, fmt.Errorf("dsl parser: parse: %w", err)
	}
	model, err := dsl.Transform(ast)
	if err != nil {
		return nil, fmt.Errorf("dsl parser: transform: %w", err)
	}
	return model, nil
}

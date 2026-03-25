package usecase

import (
	"io"

	"github.com/uber/unm-platform/internal/domain/entity"
	"github.com/uber/unm-platform/internal/domain/service"
)

// ModelParser parses a UNM model from a reader.
// Defined here so the infrastructure layer depends inward on the use case layer.
type ModelParser interface {
	Parse(r io.Reader) (*entity.UNMModel, error)
}

// ModelValidator validates a UNM model and returns a ValidationResult.
type ModelValidator interface {
	Validate(m *entity.UNMModel) service.ValidationResult
}

// ParseAndValidate orchestrates parsing a UNM model and validating it.
type ParseAndValidate struct {
	parser    ModelParser
	validator ModelValidator
}

// NewParseAndValidate constructs a ParseAndValidate use case.
func NewParseAndValidate(parser ModelParser, validator ModelValidator) *ParseAndValidate {
	return &ParseAndValidate{parser: parser, validator: validator}
}

// Execute parses the reader into a UNMModel and validates it.
// Returns the model and validation result. Returns an error only on parse failure.
func (uc *ParseAndValidate) Execute(r io.Reader) (*entity.UNMModel, service.ValidationResult, error) {
	model, err := uc.parser.Parse(r)
	if err != nil {
		return nil, service.ValidationResult{}, err
	}
	result := uc.validator.Validate(model)
	return model, result, nil
}

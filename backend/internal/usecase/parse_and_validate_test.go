package usecase_test

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/uber/unm-platform/internal/domain/entity"
	domainservice "github.com/uber/unm-platform/internal/domain/service"
	"github.com/uber/unm-platform/internal/usecase"
)

// stubParser simulates a ModelParser for testing.
type stubParser struct {
	model *entity.UNMModel
	err   error
}

func (s *stubParser) Parse(_ io.Reader) (*entity.UNMModel, error) {
	return s.model, s.err
}

// stubValidator simulates a ModelValidator for testing.
type stubValidator struct {
	result domainservice.ValidationResult
}

func (s *stubValidator) Validate(_ *entity.UNMModel) domainservice.ValidationResult {
	return s.result
}

func TestParseAndValidate_Success(t *testing.T) {
	model := entity.NewUNMModel("Test", "")
	parser := &stubParser{model: model}
	validator := &stubValidator{result: domainservice.ValidationResult{}}

	uc := usecase.NewParseAndValidate(parser, validator)
	gotModel, result, err := uc.Execute(strings.NewReader("ignored"))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotModel != model {
		t.Error("expected model returned from parser")
	}
	if !result.IsValid() {
		t.Error("expected valid result")
	}
}

func TestParseAndValidate_ParseError(t *testing.T) {
	parseErr := errors.New("bad YAML")
	parser := &stubParser{err: parseErr}
	validator := &stubValidator{}

	uc := usecase.NewParseAndValidate(parser, validator)
	_, _, err := uc.Execute(strings.NewReader(""))

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, parseErr) {
		t.Errorf("expected parse error, got %v", err)
	}
}

func TestParseAndValidate_ValidationErrors(t *testing.T) {
	model := entity.NewUNMModel("Test", "")
	parser := &stubParser{model: model}
	validator := &stubValidator{result: domainservice.ValidationResult{
		Errors: []domainservice.ValidationError{
			{Code: domainservice.ErrNeedNoCapability, Entity: "some-need", Message: "no capability"},
		},
	}}

	uc := usecase.NewParseAndValidate(parser, validator)
	_, result, err := uc.Execute(strings.NewReader("ignored"))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsValid() {
		t.Error("expected invalid result")
	}
	if len(result.Errors) != 1 {
		t.Errorf("expected 1 error, got %d", len(result.Errors))
	}
}

func TestParseAndValidate_ValidationWarnings(t *testing.T) {
	model := entity.NewUNMModel("Test", "")
	parser := &stubParser{model: model}
	validator := &stubValidator{result: domainservice.ValidationResult{
		Warnings: []domainservice.ValidationWarning{
			{Code: domainservice.WarnOrphanService, Entity: "svc", Message: "orphan"},
		},
	}}

	uc := usecase.NewParseAndValidate(parser, validator)
	_, result, err := uc.Execute(strings.NewReader("ignored"))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsValid() {
		t.Error("expected valid result (warnings don't make it invalid)")
	}
	if !result.HasWarnings() {
		t.Error("expected warnings")
	}
}

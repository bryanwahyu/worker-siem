package services_test

import (
	"testing"
	"time"

	"github.com/bryanwahyu/worker-siem/internal/domain/entities"
	"github.com/bryanwahyu/worker-siem/internal/domain/services"
	"github.com/bryanwahyu/worker-siem/internal/domain/valueobjects"
)

func TestEventValidatorValid(t *testing.T) {
	ip, _ := valueobjects.NewIP("10.0.0.1")
	status, _ := valueobjects.NewStatusCode(201)
	event := entities.NewEvent("id", time.Now(), "message", "source", "info", "host", ip, status, nil)
	validator := services.NewEventValidator()

	if err := validator.Validate(event); err != nil {
		t.Fatalf("expected valid event, got error: %v", err)
	}
}

func TestEventValidatorInvalid(t *testing.T) {
	validator := services.NewEventValidator()
	if err := validator.Validate(entities.Event{}); err == nil {
		t.Fatalf("expected validation error")
	}
}

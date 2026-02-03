package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthEndpoint(t *testing.T) {
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	handleHealth(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if body := w.Body.String(); body != "ok\n" {
		t.Fatalf("expected 'ok\\n', got %q", body)
	}
}

func TestExamplesEndpoint(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/examples", nil)
	w := httptest.NewRecorder()
	handleExamples(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var programs []PIOProgram
	if err := json.NewDecoder(w.Body).Decode(&programs); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if len(programs) != 3 {
		t.Fatalf("expected 3 examples, got %d", len(programs))
	}
}

func TestValidatePIO_ValidProgram(t *testing.T) {
	source := `.program squarewave
again:
    set pins, 1
    set pins, 0
    jmp again`

	result := validatePIO(source)
	if !result.Valid {
		t.Fatalf("expected valid, got errors: %v", result.Errors)
	}
	if len(result.Instructions) != 3 {
		t.Fatalf("expected 3 instructions, got %d", len(result.Instructions))
	}
}

func TestValidatePIO_InvalidOpcode(t *testing.T) {
	source := `    badop pins, 1`

	result := validatePIO(source)
	if result.Valid {
		t.Fatal("expected invalid program")
	}
	if len(result.Errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(result.Errors))
	}
}

func TestValidatePIO_TooManyInstructions(t *testing.T) {
	var lines string
	for i := 0; i < 33; i++ {
		lines += "    nop\n"
	}

	result := validatePIO(lines)
	if result.Valid {
		t.Fatal("expected invalid for >32 instructions")
	}
}

func TestValidatePIO_SideSetAndDelay(t *testing.T) {
	source := `    out pins, 1  side 0 [1]
    nop          side 1 [2]`

	result := validatePIO(source)
	if !result.Valid {
		t.Fatalf("expected valid, got errors: %v", result.Errors)
	}
	if len(result.Instructions) != 2 {
		t.Fatalf("expected 2 instructions, got %d", len(result.Instructions))
	}
}

func TestValidateEndpoint(t *testing.T) {
	body, _ := json.Marshal(map[string]string{
		"source": "    set pins, 1\n    jmp 0",
	})
	req := httptest.NewRequest("POST", "/api/validate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handleValidate(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var result ValidateResult
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if !result.Valid {
		t.Fatalf("expected valid, got errors: %v", result.Errors)
	}
}

func TestValidateEndpoint_MethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/validate", nil)
	w := httptest.NewRecorder()
	handleValidate(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

// PIOProgram represents a PIO assembly program.
type PIOProgram struct {
	Name        string   `json:"name"`
	Source      string   `json:"source"`
	Description string   `json:"description,omitempty"`
	Instructions []string `json:"instructions,omitempty"`
}

// PIOInstruction represents a parsed PIO instruction.
type PIOInstruction struct {
	Line    int    `json:"line"`
	Op      string `json:"op"`
	Args    string `json:"args,omitempty"`
	Comment string `json:"comment,omitempty"`
}

// ValidateResult holds the result of validating a PIO program.
type ValidateResult struct {
	Valid        bool             `json:"valid"`
	Instructions []PIOInstruction `json:"instructions"`
	Errors       []string         `json:"errors,omitempty"`
}

// Known PIO opcodes (RP2040 PIO instruction set).
var validOpcodes = map[string]bool{
	"jmp": true, "wait": true, "in": true, "out": true,
	"push": true, "pull": true, "mov": true, "irq": true, "set": true,
	"nop": true,
}

// Example PIO programs for reference.
var examples = []PIOProgram{
	{
		Name:        "squarewave",
		Description: "Simple square wave generator",
		Source: `.program squarewave
again:
    set pins, 1 [1]  ; Drive pin high and delay
    set pins, 0       ; Drive pin low
    jmp again         ; Loop`,
	},
	{
		Name:        "ws2812",
		Description: "WS2812 (Neopixel) LED driver",
		Source: `.program ws2812
.side_set 1
bitloop:
    out x, 1       side 0 [2]  ; Shift 1 bit, drive low
    jmp !x, do_zero side 1 [1] ; Branch on bit value
    jmp bitloop    side 1 [4]  ; Bit is 1: long pulse
do_zero:
    nop            side 0 [4]  ; Bit is 0: short pulse`,
	},
	{
		Name:        "spi_tx",
		Description: "SPI transmit-only master",
		Source: `.program spi_tx
.side_set 1
    out pins, 1  side 0 [1]  ; Write data, clock low
    nop          side 1 [1]  ; Clock high`,
	},
}

func main() {
	port := os.Getenv("TINYPIO_PORT")
	if port == "" {
		port = "8090"
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/api/examples", handleExamples)
	mux.HandleFunc("/api/validate", handleValidate)
	mux.HandleFunc("/", handleIndex)

	fmt.Printf("tinypio listening on :%s\n", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "ok")
}

func handleExamples(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(examples)
}

func handleValidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST required", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Source string `json:"source"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	result := validatePIO(req.Source)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func validatePIO(source string) ValidateResult {
	var instructions []PIOInstruction
	var errors []string

	lines := strings.Split(source, "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, ";") || strings.HasPrefix(trimmed, ".") {
			continue
		}

		// Strip inline comment FIRST (before label check, since comments can contain colons)
		comment := ""
		if idx := strings.Index(trimmed, ";"); idx >= 0 {
			comment = strings.TrimSpace(trimmed[idx+1:])
			trimmed = strings.TrimSpace(trimmed[:idx])
		}

		// Strip label prefix (e.g., "again:")
		if idx := strings.Index(trimmed, ":"); idx >= 0 {
			rest := strings.TrimSpace(trimmed[idx+1:])
			if rest == "" {
				continue
			}
			trimmed = rest
		}

		// Strip side_set and delay annotations
		if idx := strings.Index(trimmed, "side"); idx >= 0 {
			trimmed = strings.TrimSpace(trimmed[:idx])
		}
		// Strip delay [N]
		if idx := strings.Index(trimmed, "["); idx >= 0 {
			trimmed = strings.TrimSpace(trimmed[:idx])
		}

		parts := strings.Fields(trimmed)
		if len(parts) == 0 {
			continue
		}

		op := strings.ToLower(parts[0])
		args := ""
		if len(parts) > 1 {
			args = strings.Join(parts[1:], " ")
		}

		inst := PIOInstruction{
			Line:    i + 1,
			Op:      op,
			Args:    args,
			Comment: comment,
		}
		instructions = append(instructions, inst)

		if !validOpcodes[op] {
			errors = append(errors, fmt.Sprintf("line %d: unknown opcode '%s'", i+1, op))
		}
	}

	if len(instructions) > 32 {
		errors = append(errors, fmt.Sprintf("program has %d instructions, max is 32", len(instructions)))
	}

	return ValidateResult{
		Valid:        len(errors) == 0,
		Instructions: instructions,
		Errors:       errors,
	}
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, indexHTML)
}

const indexHTML = `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<title>tinypio</title>
<style>
  body { font-family: system-ui; max-width: 800px; margin: 2rem auto; padding: 0 1rem; }
  textarea { width: 100%; height: 200px; font-family: monospace; font-size: 14px; }
  pre { background: #f5f5f5; padding: 1rem; overflow-x: auto; }
  .error { color: #dc3545; }
  .valid { color: #28a745; }
  button { padding: 0.5rem 1rem; cursor: pointer; }
  .examples { display: flex; gap: 0.5rem; margin-bottom: 1rem; }
  .examples button { font-size: 0.85rem; }
</style>
</head>
<body>
<h1>tinypio - PIO Assembly Validator</h1>
<p>Validate RP2040/RP2350 PIO assembly programs.
<a href="https://github.com/tinygo-org/pio">TinyGo PIO</a></p>

<div class="examples">
  <span>Examples:</span>
  <button onclick="loadExample('squarewave')">Square Wave</button>
  <button onclick="loadExample('ws2812')">WS2812</button>
  <button onclick="loadExample('spi_tx')">SPI TX</button>
</div>

<textarea id="source" placeholder="Paste PIO assembly here..."></textarea>
<br><br>
<button onclick="validate()">Validate</button>
<div id="result"></div>

<script>
async function loadExample(name) {
  const resp = await fetch('/api/examples');
  const examples = await resp.json();
  const ex = examples.find(e => e.name === name);
  if (ex) document.getElementById('source').value = ex.source;
}

async function validate() {
  const source = document.getElementById('source').value;
  const resp = await fetch('/api/validate', {
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({source})
  });
  const data = await resp.json();
  let html = '';
  if (data.valid) {
    html += '<p class="valid">Valid PIO program (' + data.instructions.length + '/32 instructions)</p>';
  } else {
    html += '<p class="error">Invalid:</p><ul>';
    data.errors.forEach(e => html += '<li class="error">' + e + '</li>');
    html += '</ul>';
  }
  html += '<pre>' + JSON.stringify(data.instructions, null, 2) + '</pre>';
  document.getElementById('result').innerHTML = html;
}

loadExample('squarewave');
</script>
</body>
</html>`

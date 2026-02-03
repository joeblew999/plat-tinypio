package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

// PIOProgram represents a PIO assembly program.
type PIOProgram struct {
	Name        string   `json:"name"`
	Source      string   `json:"source"`
	Description string   `json:"description,omitempty"`
	Instructions []string `json:"instructions,omitempty"`
}

// CompileResult holds the result of compiling a PIO program with pioasm.
type CompileResult struct {
	Success bool     `json:"success"`
	Binary  []uint16 `json:"binary,omitempty"`
	Hex     string   `json:"hex,omitempty"`
	Go      string   `json:"go,omitempty"`
	Errors  []string `json:"errors,omitempty"`
}

// Driver represents a ready-to-use PIO driver from tinygo-org/pio.
type Driver struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Package     string `json:"package"`
	Example     string `json:"example,omitempty"`
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

// Available drivers from tinygo-org/pio/rp2-pio/piolib.
var drivers = []Driver{
	{
		Name:        "WS2812B",
		Description: "WS2812B (NeoPixel) RGB LED strip controller with DMA support",
		Package:     "github.com/tinygo-org/pio/rp2-pio/piolib",
		Example: `ws, err := piolib.NewWS2812B(sm, machine.GP0)
ws.PutRGB(255, 0, 0) // Red`,
	},
	{
		Name:        "SPI",
		Description: "SPI master implementation using PIO",
		Package:     "github.com/tinygo-org/pio/rp2-pio/piolib",
		Example:     `spi, err := piolib.NewSPI(sm, clkPin, mosiPin, misoPin)`,
	},
	{
		Name:        "Parallel",
		Description: "8-pin send-only parallel bus for displays",
		Package:     "github.com/tinygo-org/pio/rp2-pio/piolib",
		Example:     `bus, err := piolib.NewParallel8(sm, dataPin, clockPin)`,
	},
	{
		Name:        "I2S",
		Description: "I2S audio output driver",
		Package:     "github.com/tinygo-org/pio/rp2-pio/piolib",
		Example:     `i2s, err := piolib.NewI2S(sm, bclkPin, lrclkPin, dataPin)`,
	},
	{
		Name:        "Pulsar",
		Description: "Pulse-constrained square wave generator",
		Package:     "github.com/tinygo-org/pio/rp2-pio/piolib",
		Example:     `pulsar, err := piolib.NewPulsar(sm, pin, freq)`,
	},
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
	mux.HandleFunc("/api/compile", handleCompile)
	mux.HandleFunc("/api/drivers", handleDrivers)
	mux.HandleFunc("/api/status", handleStatus)
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

func handleCompile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST required", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Source string `json:"source"`
		Format string `json:"format"` // "hex", "go", or "binary" (default)
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	result := compilePIO(req.Source, req.Format)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func handleDrivers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(drivers)
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
	pioasmPath, _ := exec.LookPath("pioasm")
	status := map[string]interface{}{
		"validator":     true,
		"pioasm":        pioasmPath != "",
		"pioasm_path":   pioasmPath,
		"drivers":       len(drivers),
		"examples":      len(examples),
		"upstream":      "github.com/tinygo-org/pio",
		"max_instructions": 32,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func compilePIO(source, format string) CompileResult {
	// Check if pioasm is available
	pioasmPath, err := exec.LookPath("pioasm")
	if err != nil {
		return CompileResult{
			Success: false,
			Errors:  []string{"pioasm not found. Install from: https://github.com/raspberrypi/pico-sdk/tree/master/tools/pioasm"},
		}
	}

	// Write source to temp file
	tmpFile, err := os.CreateTemp("", "pio-*.pio")
	if err != nil {
		return CompileResult{Success: false, Errors: []string{err.Error()}}
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(source); err != nil {
		return CompileResult{Success: false, Errors: []string{err.Error()}}
	}
	tmpFile.Close()

	// Determine output format flag
	formatFlag := "-o"
	switch format {
	case "go":
		formatFlag = "-o"
	case "hex":
		formatFlag = "-o"
	default:
		format = "hex"
		formatFlag = "-o"
	}

	// Run pioasm
	outFile, err := os.CreateTemp("", "pio-out-*")
	if err != nil {
		return CompileResult{Success: false, Errors: []string{err.Error()}}
	}
	outFile.Close()
	defer os.Remove(outFile.Name())

	var stderr bytes.Buffer
	cmd := exec.Command(pioasmPath, formatFlag, format, tmpFile.Name(), outFile.Name())
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		errMsg := stderr.String()
		if errMsg == "" {
			errMsg = err.Error()
		}
		return CompileResult{Success: false, Errors: []string{errMsg}}
	}

	// Read output
	output, err := os.ReadFile(outFile.Name())
	if err != nil {
		return CompileResult{Success: false, Errors: []string{err.Error()}}
	}

	result := CompileResult{Success: true}
	switch format {
	case "go":
		result.Go = string(output)
	case "hex":
		result.Hex = string(output)
		// Parse hex to binary
		result.Binary = parseHexProgram(string(output))
	}

	return result
}

func parseHexProgram(hexOutput string) []uint16 {
	var binary []uint16
	lines := strings.Split(hexOutput, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "//") || strings.HasPrefix(line, "#") {
			continue
		}
		// Try to parse as hex value (0xNNNN format)
		line = strings.TrimPrefix(line, "0x")
		line = strings.TrimSuffix(line, ",")
		if len(line) >= 4 {
			b, err := hex.DecodeString(line[:4])
			if err == nil && len(b) >= 2 {
				binary = append(binary, uint16(b[0])<<8|uint16(b[1]))
			}
		}
	}
	return binary
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
<title>tinypio - PIO Development Toolkit</title>
<style>
  body { font-family: system-ui; max-width: 900px; margin: 2rem auto; padding: 0 1rem; }
  textarea { width: 100%; height: 200px; font-family: monospace; font-size: 14px; }
  pre { background: #f5f5f5; padding: 1rem; overflow-x: auto; font-size: 13px; }
  .error { color: #dc3545; }
  .valid { color: #28a745; }
  .warning { color: #ffc107; }
  button { padding: 0.5rem 1rem; cursor: pointer; margin-right: 0.5rem; }
  button.primary { background: #0066cc; color: white; border: none; }
  .examples { display: flex; gap: 0.5rem; margin-bottom: 1rem; align-items: center; }
  .examples button { font-size: 0.85rem; }
  .tabs { display: flex; gap: 0; margin-top: 1rem; border-bottom: 2px solid #ddd; }
  .tabs button { border: none; background: #f5f5f5; padding: 0.5rem 1rem; cursor: pointer; border-radius: 4px 4px 0 0; }
  .tabs button.active { background: #0066cc; color: white; }
  .tab-content { display: none; padding: 1rem 0; }
  .tab-content.active { display: block; }
  .status { font-size: 0.85rem; color: #666; margin-top: 2rem; padding: 1rem; background: #f9f9f9; border-radius: 4px; }
  .status .ok { color: #28a745; }
  .status .missing { color: #dc3545; }
  .driver-list { display: grid; gap: 1rem; margin-top: 1rem; }
  .driver { background: #f5f5f5; padding: 1rem; border-radius: 4px; }
  .driver h4 { margin: 0 0 0.5rem 0; }
  .driver code { background: #e0e0e0; padding: 0.2rem 0.4rem; border-radius: 2px; font-size: 0.8rem; }
  .actions { margin: 1rem 0; }
</style>
</head>
<body>
<h1>tinypio - PIO Development Toolkit</h1>
<p>Validate and compile RP2040/RP2350 PIO assembly programs.
Powered by <a href="https://github.com/tinygo-org/pio">TinyGo PIO</a>.</p>

<div class="examples">
  <span>Examples:</span>
  <button onclick="loadExample('squarewave')">Square Wave</button>
  <button onclick="loadExample('ws2812')">WS2812</button>
  <button onclick="loadExample('spi_tx')">SPI TX</button>
</div>

<textarea id="source" placeholder="Paste PIO assembly here..."></textarea>

<div class="actions">
  <button class="primary" onclick="validate()">Validate</button>
  <button onclick="compile('hex')">Compile (Hex)</button>
  <button onclick="compile('go')">Compile (Go)</button>
</div>

<div class="tabs">
  <button class="active" onclick="showTab('validation')">Validation</button>
  <button onclick="showTab('compiled')">Compiled Output</button>
  <button onclick="showTab('drivers')">Drivers</button>
</div>

<div id="validation" class="tab-content active">
  <div id="result"></div>
</div>

<div id="compiled" class="tab-content">
  <div id="compile-result"></div>
</div>

<div id="drivers" class="tab-content">
  <p>Ready-to-use PIO drivers from <code>github.com/tinygo-org/pio/rp2-pio/piolib</code>:</p>
  <div id="driver-list" class="driver-list"></div>
</div>

<div class="status" id="status">Loading status...</div>

<script>
async function loadExample(name) {
  const resp = await fetch('/api/examples');
  const examples = await resp.json();
  const ex = examples.find(e => e.name === name);
  if (ex) document.getElementById('source').value = ex.source;
}

async function validate() {
  showTab('validation');
  const source = document.getElementById('source').value;
  const resp = await fetch('/api/validate', {
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({source})
  });
  const data = await resp.json();
  let html = '';
  if (data.valid) {
    html += '<p class="valid">✓ Valid PIO program (' + data.instructions.length + '/32 instructions)</p>';
  } else {
    html += '<p class="error">✗ Invalid:</p><ul>';
    data.errors.forEach(e => html += '<li class="error">' + e + '</li>');
    html += '</ul>';
  }
  if (data.instructions && data.instructions.length > 0) {
    html += '<h4>Parsed Instructions:</h4>';
    html += '<pre>' + JSON.stringify(data.instructions, null, 2) + '</pre>';
  }
  document.getElementById('result').innerHTML = html;
}

async function compile(format) {
  showTab('compiled');
  const source = document.getElementById('source').value;
  const resp = await fetch('/api/compile', {
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({source, format})
  });
  const data = await resp.json();
  let html = '';
  if (data.success) {
    html += '<p class="valid">✓ Compilation successful</p>';
    if (data.go) {
      html += '<h4>Go Output:</h4><pre>' + escapeHtml(data.go) + '</pre>';
    }
    if (data.hex) {
      html += '<h4>Hex Output:</h4><pre>' + escapeHtml(data.hex) + '</pre>';
    }
    if (data.binary && data.binary.length > 0) {
      html += '<h4>Binary (' + data.binary.length + ' instructions):</h4>';
      html += '<pre>' + data.binary.map(b => '0x' + b.toString(16).padStart(4, '0')).join(', ') + '</pre>';
    }
  } else {
    html += '<p class="error">✗ Compilation failed:</p><ul>';
    data.errors.forEach(e => html += '<li class="error">' + e + '</li>');
    html += '</ul>';
  }
  document.getElementById('compile-result').innerHTML = html;
}

function escapeHtml(text) {
  const div = document.createElement('div');
  div.textContent = text;
  return div.innerHTML;
}

function showTab(name) {
  document.querySelectorAll('.tabs button').forEach(b => b.classList.remove('active'));
  document.querySelectorAll('.tab-content').forEach(c => c.classList.remove('active'));
  document.querySelector('.tabs button[onclick*="' + name + '"]').classList.add('active');
  document.getElementById(name).classList.add('active');
}

async function loadDrivers() {
  const resp = await fetch('/api/drivers');
  const drivers = await resp.json();
  let html = '';
  drivers.forEach(d => {
    html += '<div class="driver">';
    html += '<h4>' + d.name + '</h4>';
    html += '<p>' + d.description + '</p>';
    html += '<code>' + d.package + '</code>';
    if (d.example) {
      html += '<pre>' + escapeHtml(d.example) + '</pre>';
    }
    html += '</div>';
  });
  document.getElementById('driver-list').innerHTML = html;
}

async function loadStatus() {
  const resp = await fetch('/api/status');
  const s = await resp.json();
  let html = '<strong>Status:</strong> ';
  html += 'Validator <span class="ok">✓</span> | ';
  html += 'pioasm ' + (s.pioasm ? '<span class="ok">✓</span>' : '<span class="missing">✗ not installed</span>') + ' | ';
  html += s.drivers + ' drivers | ' + s.examples + ' examples | ';
  html += 'Max ' + s.max_instructions + ' instructions';
  if (!s.pioasm) {
    html += '<br><small>Install pioasm: <code>git clone pico-sdk && cd tools/pioasm && cmake . && make && sudo make install</code></small>';
  }
  document.getElementById('status').innerHTML = html;
}

loadExample('squarewave');
loadDrivers();
loadStatus();
</script>
</body>
</html>`

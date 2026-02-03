# Usage Guide

How to use the TinyPIO PIO development toolkit.

## Quick Start

```bash
# Install xplat
curl -fsSL https://raw.githubusercontent.com/joeblew999/xplat/main/install.sh | sh

# Clone and run
git clone https://github.com/joeblew999/plat-tinypio
cd plat-tinypio
xplat up
```

Open the web interface at http://localhost:8090

## Web Interface

1. Enter your PIO assembly code in the editor
2. Click **Validate** for syntax checking (no dependencies)
3. Click **Compile (Hex/Go)** for full compilation (requires pioasm)
4. Browse the **Drivers** tab for ready-to-use TinyGo drivers

## API Endpoints

### POST /api/validate

Validate PIO assembly syntax (fast, no external dependencies).

```bash
curl -X POST http://localhost:8090/api/validate \
  -H "Content-Type: application/json" \
  -d '{"source": ".program test\nset pins, 1\njmp 0"}'
```

Response:
```json
{
  "valid": true,
  "instructions": [
    {"line": 2, "op": "set", "args": "pins, 1"},
    {"line": 3, "op": "jmp", "args": "0"}
  ]
}
```

### POST /api/compile

Compile PIO assembly with pioasm (requires pioasm binary installed).

```bash
curl -X POST http://localhost:8090/api/compile \
  -H "Content-Type: application/json" \
  -d '{"source": ".program test\nset pins, 1", "format": "hex"}'
```

Formats: `hex`, `go`

### GET /api/examples

Get built-in example programs.

```bash
curl http://localhost:8090/api/examples
```

### GET /api/drivers

List available TinyGo PIO drivers.

```bash
curl http://localhost:8090/api/drivers
```

### GET /api/status

Check toolkit capabilities.

```bash
curl http://localhost:8090/api/status
```

### GET /health

Health check endpoint.

```bash
curl http://localhost:8090/health
```

## Supported Instructions

| Opcode | Description |
|--------|-------------|
| `jmp` | Jump (conditional/unconditional) |
| `wait` | Wait for GPIO/IRQ condition |
| `in` | Shift data into ISR |
| `out` | Shift data out of OSR |
| `push` | Push ISR to RX FIFO |
| `pull` | Pull from TX FIFO to OSR |
| `mov` | Move data between registers |
| `irq` | Set/clear/wait IRQ flags |
| `set` | Set pins/pindirs/register |
| `nop` | No operation |

## Example Programs

### Square Wave

```pio
.program squarewave
again:
    set pins, 1 [1]  ; Drive pin high and delay
    set pins, 0      ; Drive pin low
    jmp again        ; Loop
```

### WS2812 LED Driver

See the built-in examples in the web interface for a complete WS2812 driver.

## Installing pioasm

For full compilation support, install pioasm from pico-sdk:

```bash
git clone https://github.com/raspberrypi/pico-sdk.git
cd pico-sdk/tools/pioasm
cmake .
make
sudo make install
```

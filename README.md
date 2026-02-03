# plat-tinypio

A web-based **PIO (Programmable I/O) instruction validator** for RP2040 and RP2350 microcontrollers.

Validate your PIO assembly code instantly in the browser - check syntax, instruction counts, side-set/delay values, and more.

## Live Demo

**https://joeblew999.github.io/plat-tinypio/**

## Features

- **Real-time validation** - Instant feedback as you type
- **Full PIO instruction set** - Supports all RP2040/RP2350 PIO opcodes
- **Detailed error messages** - Line-by-line validation with specific error descriptions
- **Example programs** - Built-in examples including WS2812 LED driver, UART TX, and SPI
- **REST API** - Validate PIO code programmatically

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

## API Endpoints

### POST /validate
Validate PIO assembly code.

```bash
curl -X POST http://localhost:8090/validate \
  -H "Content-Type: application/json" \
  -d '{"code": "set pins, 1\nnop\nset pins, 0"}'
```

Response:
```json
{
  "valid": true,
  "instruction_count": 3,
  "errors": []
}
```

### GET /examples
Get built-in example programs.

```bash
curl http://localhost:8090/examples
```

### GET /health
Health check endpoint.

```bash
curl http://localhost:8090/health
```

## Development

### Prerequisites

- [xplat](https://github.com/joeblew999/xplat) - Cross-platform task runner

### Quick Start

```bash
# Install xplat
curl -fsSL https://raw.githubusercontent.com/joeblew999/xplat/main/install.sh | sh

# Clone and run
git clone https://github.com/joeblew999/plat-tinypio
cd plat-tinypio
xplat up
```

### Commands

```bash
xplat task build    # Build the binary
xplat task test     # Run tests
xplat task run      # Build and run
xplat task dev      # Run with hot reload (auto-rebuild on file changes)
xplat up            # Start with web UI dashboard
```

### Project Structure

```
plat-tinypio/
├── cmd/tinypio/
│   ├── main.go        # HTTP server and PIO validator
│   └── main_test.go   # Tests
├── xplat.yaml         # Project manifest
├── Taskfile.yml       # Task definitions
├── process-compose.yaml # Process orchestration
└── docs/              # Documentation (GitHub Pages)
```

## Configuration

Copy `.env.example` to `.env` to customize:

```bash
cp .env.example .env
```

## Links

- **Docs**: https://joeblew999.github.io/plat-tinypio/
- **Repo**: https://github.com/joeblew999/plat-tinypio
- **xplat**: https://github.com/joeblew999/xplat

## License

MIT

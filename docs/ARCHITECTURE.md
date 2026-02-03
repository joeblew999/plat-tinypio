# Architecture

TinyPIO is a Go web server providing PIO development tools.

## Components

```
plat-tinypio/
├── cmd/tinypio/         # HTTP server with validator, compiler, driver catalog
├── .src/pio/            # Cloned upstream tinygo-org/pio library
├── docs/                # Documentation (GitHub Pages)
├── xplat.yaml           # Project manifest
├── Taskfile.yml         # Build tasks
└── process-compose.yaml # Process orchestration
```

## Features

| Feature | Description | Dependencies |
|---------|-------------|--------------|
| Validator | Fast PIO syntax checking | None |
| Compiler | Full pioasm compilation | pioasm binary |
| Drivers | TinyGo driver catalog | Reference only |

## How It Works

1. **Web Interface** - Static HTML/JS served at `/`
2. **Validation API** - `/api/validate` - parses and validates PIO assembly
3. **Compile API** - `/api/compile` - calls pioasm for full compilation
4. **Driver Catalog** - `/api/drivers` - lists tinygo-org/pio drivers

## Validation

The validator checks:

| Check | Description |
|-------|-------------|
| Opcodes | Valid PIO instruction (jmp, wait, in, out, push, pull, mov, irq, set, nop) |
| Instruction count | Max 32 instructions per program |
| Comments | Strips `;` comments |
| Labels | Strips label definitions |
| Directives | Ignores `.program`, `.wrap`, `.side_set` |

## Upstream

Uses [tinygo-org/pio](https://github.com/tinygo-org/pio) for:

- Go assembler API (`AssemblerV0`)
- Ready-to-use drivers (WS2812B, SPI, I2S, Parallel)
- Platform support (RP2040, RP2350)

## Deployment

Single binary - no runtime dependencies for validation.

```bash
xplat task build
./tinypio
```

Compilation requires pioasm from pico-sdk.

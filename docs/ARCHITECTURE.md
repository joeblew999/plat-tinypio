# Architecture

TinyPIO is a simple Go web server that validates PIO assembly code.

## Components

```
plat-tinypio/
├── cmd/tinypio/      # Main entry point
├── bin/              # Built binary
├── docs/             # Documentation
└── Taskfile.yml      # Build tasks
```

## How It Works

1. **Web Interface** - Static HTML/JS served at `/`
2. **Validation API** - REST endpoint at `/api/validate`
3. **Parser** - Tokenizes PIO assembly
4. **Validator** - Checks instruction syntax and constraints

## Instruction Validation

The validator checks:

| Check | Description |
|-------|-------------|
| Syntax | Valid opcode and operands |
| Side-set | Within configured limits |
| Delay | Max 31 cycles |
| Labels | All jump targets exist |
| Directives | Valid `.program`, `.wrap`, etc. |

## Deployment

Runs as a single binary - no dependencies required.

```bash
xplat task build
./tinypio serve
```

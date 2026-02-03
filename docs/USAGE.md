# Usage Guide

How to use the TinyPIO validator.

## Quick Start

1. Open the web interface at http://localhost:8080
2. Enter your PIO assembly code in the editor
3. Click "Validate" to check for errors

## API Usage

### Validate Endpoint

```bash
curl -X POST http://localhost:8080/api/validate \
  -H "Content-Type: application/json" \
  -d '{"code": ".program test\njmp 0"}'
```

### Response Format

```json
{
  "valid": true,
  "errors": [],
  "instructions": 1
}
```

## Example Programs

### Blink LED

```pio
.program blink
.wrap_target
    set pins, 1   [31]  ; Turn LED on
    nop           [31]
    set pins, 0   [31]  ; Turn LED off
    nop           [31]
.wrap
```

### WS2812 LED Driver

See the built-in examples in the web interface for a complete WS2812 driver.

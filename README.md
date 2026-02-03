# plat-tinypio

PIO development toolkit for RP2040/RP2350 microcontrollers.

## Why?

**PIO (Programmable I/O)** is one of the most powerful features of RP2040/RP2350 chips - custom hardware interfaces in software. But developing PIO programs is painful:

- No instant feedback - you flash, test, repeat
- pioasm requires pico-sdk setup
- Finding working driver code means hunting through examples

**tinypio** solves this with a web-based toolkit:

1. **Validate instantly** - Check PIO syntax without any toolchain
2. **Compile to hex/Go** - Full pioasm compilation when you need it
3. **Browse drivers** - Ready-to-use TinyGo drivers for common protocols

Built on [tinygo-org/pio](https://github.com/tinygo-org/pio) - the Go library for PIO development. Thanks to [@soypat](https://github.com/soypat) for creating and maintaining the upstream library.

## Try It

**Live**: https://joeblew999.github.io/plat-tinypio/

**Local**:
```bash
xplat up
# Open http://localhost:8090
```

## Links

- [Usage Guide](https://joeblew999.github.io/plat-tinypio/docs/USAGE)
- [Architecture](https://joeblew999.github.io/plat-tinypio/docs/ARCHITECTURE)
- [tinygo-org/pio](https://github.com/tinygo-org/pio)
- [xplat](https://github.com/joeblew999/xplat)

## License

MIT

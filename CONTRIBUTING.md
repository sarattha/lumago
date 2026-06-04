# Contributing to LumaGo

LumaGo is currently a pre-alpha engine scaffold. Contributions should focus on the MVP roadmap and keep the game-facing API simple.

## Principles

1. Keep gameplay code independent from Vulkan.
2. Prefer batched data submission over per-object renderer calls.
3. Keep hot-path allocations low.
4. Build small, testable packages.
5. Document renderer assumptions clearly.

## Development Flow

```bash
make fmt
make test
make run
```

## Pull Request Checklist

- [ ] Code is formatted with `go fmt`.
- [ ] `go test ./...` passes.
- [ ] New engine concepts are documented.
- [ ] Renderer changes do not leak Vulkan details into game-facing packages.
- [ ] Hot-path allocations are considered.

# NephiteCoin Prototype

This repository captures an initial concept for a Bitcoin-inspired cryptocurrency that values balances with the Nephite monetary system from Alma 11.

- `docs/nephite-coin-spec.md` — technical specification covering consensus, monetary policy, and implementation roadmap.
- `cmd/nephite-denom/main.go` — Go CLI for converting between `leah` (atomic unit) and higher denominations such as `onti`, `limnah`, and `senum`.

## Quick Start

```bash
go run ./cmd/nephite-denom to-leah 3 limnah          # -> 168
go run ./cmd/nephite-denom breakdown 168             # -> 3 limnahs
go run ./cmd/nephite-denom breakdown 123 --preference gold
```

These tools illustrate how wallets and invoices can stay faithful to the Nephite denominations while leveraging Bitcoin-style infrastructure.

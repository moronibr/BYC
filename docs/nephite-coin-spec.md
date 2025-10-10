# NephiteCoin Technical Specification

## 1. Vision and Guiding Principles

NephiteCoin is a decentralized digital currency that preserves the trustless, peer-to-peer strengths of Bitcoin while expressing value through the Nephite monetary denominations recorded in Alma 11. The system aims to:

- Maintain a fixed, predictable supply schedule secured by proof-of-work.
- Use an unspent transaction output (UTXO) ledger with on-chain programmability limited to simple, auditable scripts.
- Present balances, fees, and rewards using canonical Nephite units (both gold and silver series) without losing the precision required for digital settlement.
- Remain open-source, permissionless, and resistant to censorship or centralized control.

## 2. Monetary Units

### 2.1 Base Unit

- **Atomic unit (`leah`)**: All ledger amounts are tracked as integers measured in `leah`. This avoids fractional arithmetic because every denomination is an integer multiple of `leah`.
- **Conversion**: 1 `senum` (silver) = 1 `senine` (gold) = 8 `leah`.

### 2.2 Gold Series

| Denomination | Leah multiplier | Senine equivalent | Notes |
| --- | ---: | ---: | --- |
| `senine` | 8 | 1 | Wage for a judge per day in Alma 11 |
| `seon` | 16 | 2 | Twice a `senine` |
| `shum` | 32 | 4 | Twice a `seon` |
| `limnah` | 56 | 7 | Value of all lesser gold pieces combined (`senine + seon + shum`) |
| `antion` | 12 | 1.5 | Equal to three `shiblons` (see §2.3) |

### 2.3 Silver Series and Fractions

| Denomination | Leah multiplier | Senum equivalent | Notes |
| --- | ---: | ---: | --- |
| `senum` | 8 | 1 | Equal to a `senine` |
| `amnor` | 16 | 2 | Twice a `senum` |
| `ezrom` | 32 | 4 | Twice an `amnor` |
| `onti` | 56 | 7 | Value of all lesser silver pieces combined |
| `shiblon` | 4 | 0.5 | Half a `senum` |
| `shiblum` | 2 | 0.25 | Half a `shiblon` |
| `leah` | 1 | 0.125 | Base unit |

Wallets and user interfaces SHOULD default to showing balances in the highest meaningful denomination, with optional toggles between gold and silver nomenclature.

## 3. Ledger Model and Consensus

- **Ledger**: UTXO model mirroring Bitcoin to enable stateless validation and signature aggregation.
- **Consensus**: Proof-of-Work using a memory-hard hash (placeholder `MosiahHash`) to encourage commodity hardware participation. Difficulty adjusts every 2016 blocks targeting a 10-minute interval.
- **Network**: Gossip protocol compatible with Bitcoin's P2P message formats for ease of implementation; new message types introduced for denomination metadata.
- **Script**: Bitcoin Script subset plus new opcodes for denomination-aware timelocks (e.g., `OP_CHECK_GOVERNANCE_HEIGHT`) while retaining determinism.

## 4. Monetary Policy and Issuance

- **Total cap**: 1,176,000,000 `leah` (equivalent to 21,000,000 `limnah`), inspired by Bitcoin's 21 million supply but expressed through Nephite units.
- **Block subsidy**: Starts at 2,240 `leah` (40 `senine`) per block, ensuring miner rewards can be represented with whole Nephite denominations.
- **Halving cadence**: Subsidy halves every 210,000 blocks (approximately 4 years) until reaching zero. Residual rounding always keeps rewards as multiples of `leah`.
- **Transaction fees**: Collected in `leah`, with fee estimation libraries exposing results in preferred denominations.

The emission curve mirrors Bitcoin's geometric decay, with 32 halving epochs delivering the capped subsidy.

## 5. Transaction Structure

- **Inputs/Outputs**: Each output records a value in `leah` plus a script. Wallets may include a `denomination_descriptor` metadata field for ease of display without impacting consensus-critical data.
- **Change decomposition**: Nodes accept any integer `leah` amount. Wallets SHOULD provide canonical change recommendations prioritizing `limnah`/`onti` pieces first, then descending denominations, to echo Alma 11's cultural context.
- **Fee negotiation**: Replace-By-Fee (RBF) enabled by default; mempool policies mirror Bitcoin Core 24.x.

## 6. Wallet and UX Guidelines

- **Display**: Primary mode shows balances in silver series (`onti`, `ezrom`, …, `leah`) with a toggle for gold denominations. Users can switch to raw `leah` (satoshi equivalent) for technical tasks.
- **Mnemonic and keys**: BIP-39/BIP-32 compatible key derivation with unique coin-type identifier (SLIP-0044 TBD).
- **QR encoding**: Payment URIs use the scheme `nephite:<address>?amount=<value>&unit=<denomination>`.
- **Denomination formatter**: Include library routines to convert raw `leah` amounts into mixed Nephite denominations for invoices and receipts.

## 7. Governance and Upgrades

- **Consensus changes**: BIP8-like process with miner signaling and user-activated soft forks. Hard forks require super-majority miner adoption plus a 6-month activation grace period.
- **Reference implementation**: Go-based full node and tooling suite with modular denomination layer and updated consensus constants. Light clients and SDKs also target Go to satisfy a single-language ecosystem.
- **Community stewardship**: Open Gnosis Safe (multi-sig) controlling pre-mine (if any) limited strictly to bootstrap funds; long-term goal is zero centralized treasury.

## 8. Security Considerations

- Reuse of hardened Bitcoin codebases minimizes novel attack surface.
- Memory-hard PoW mitigates ASIC concentration, but governance MUST monitor hardware evolution.
- Denomination metadata MUST remain non-consensus to avoid additional fork risk.
- Wallet UX should warn when invoices mix gold and silver naming to prevent user confusion or scams.

## 9. Implementation Roadmap

1. **Research and Prototype**
   - Finalize PoW algorithm parameters.
   - Implement denomination conversion library (CLI + SDK).
   - Draft SLIP-0044 registration and URI specification.
2. **Core Node Fork**
   - Update consensus constants (reward, cap, message prefixes).
   - Integrate denomination formatters into JSON-RPC.
   - Launch testnet with halving every 20,000 blocks for rapid evaluation.
3. **Wallet Ecosystem**
   - Release reference desktop/mobile wallet with denomination-aware UI.
   - Provide invoicing tools for merchants with preset Nephite amounts.
4. **Audit and Launch**
   - Commission third-party audits of consensus changes and wallet code.
   - Conduct public testnet trials, then lock mainnet genesis parameters.

This specification serves as the canonical reference for early-stage implementation and SHOULD evolve alongside community consensus.

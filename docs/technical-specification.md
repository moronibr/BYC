# Brigham Young Chain Technical Specification

## Overview

The Brigham Young Chain (Young Chain) is a dual-blockchain system inspired by the Nephite monetary system. This document provides technical details about the implementation, architecture, and protocols.

## System Architecture

### Dual Blockchain Structure

The system consists of two parallel blockchains:

1. **Golden Blockchain**
   - Contains gold-based transactions
   - Uses proof of work for consensus
   - Produces Ephraim coins upon block completion

2. **Silver Blockchain**
   - Contains silver-based transactions
   - Uses proof of work for consensus
   - Produces Manasseh coins upon block completion

### Block Structure

Each block contains:

```
Block Header:
- Version (4 bytes)
- Previous Block Hash (32 bytes)
- Merkle Root (32 bytes)
- Timestamp (8 bytes)
- Difficulty Target (4 bytes)
- Nonce (4 bytes)

Block Body:
- Transaction Count (varint)
- Transactions (variable)
```

### Transaction Structure

Transactions follow this format:

```
Transaction:
- Version (4 bytes)
- Input Count (varint)
- Inputs (variable)
- Output Count (varint)
- Outputs (variable)
- Lock Time (4 bytes)
- Coin Type (1 byte)
- Cross Block Flag (1 byte)
```

## Mining System

### Three-Tier Difficulty

1. **Leah Mining**
   - Base difficulty: 0x1d00ffff
   - Target block time: 10 minutes
   - Reward: 50 Leah coins

2. **Shiblum Mining**
   - Difficulty: 2x Leah difficulty
   - Target block time: 10 minutes
   - Reward: 100 Shiblum coins

3. **Shiblon Mining**
   - Difficulty: 4x Leah difficulty
   - Target block time: 10 minutes
   - Reward: 200 Shiblon coins

### Difficulty Adjustment

Difficulty is adjusted every 2016 blocks based on the actual time taken to mine the previous 2016 blocks:

```
New Difficulty = Old Difficulty * (Actual Time / Target Time)
Target Time = 2016 blocks * 10 minutes
```

## Coin System

### Mineable Coins

These coins can be mined in both chains:

1. **Leah**
   - Base unit
   - Smallest value
   - Easiest to mine

2. **Shiblum**
   - 2x Leah value
   - Medium difficulty
   - Medium value

3. **Shiblon**
   - 4x Leah value
   - Highest difficulty
   - Highest value

### Derived Coins

These coins are created through combinations:

#### Gold-based:
- Senine (derived from Leah)
- Seon (2x Senine)
- Shum (2x Seon)
- Limnah (value of all above)
- Antion (3x Shiblon, can cross chains)

#### Silver-based:
- Senum (derived from Leah)
- Amnor (2x Senum)
- Ezrom (4x Senum)
- Onti (value of all above)

### Special Coins

1. **Ephraim**
   - Total supply: 15,000,000
   - Maximum convertible to Joseph: 3,000,000
   - Produced by Golden Block completion
   - Cannot be transferred between chains

2. **Manasseh**
   - Total supply: 15,000,000
   - Maximum convertible to Joseph: 3,000,000
   - Produced by Silver Block completion
   - Cannot be transferred between chains

3. **Joseph**
   - Total supply: 3,000,000
   - Created by combining 1 Ephraim + 1 Manasseh
   - Limited by the number of convertible Ephraim and Manasseh coins
   - Can be transferred between chains

## Network Protocol

### Message Types

1. **Version (0x01)**
   - Protocol version
   - Node capabilities
   - Timestamp

2. **VerAck (0x02)**
   - Acknowledgment of version message

3. **Block (0x03)**
   - Block data
   - Block type (Golden/Silver)

4. **Transaction (0x04)**
   - Transaction data
   - Coin type

5. **GetBlocks (0x05)**
   - Block locator
   - Stop hash

6. **GetData (0x06)**
   - Inventory items

7. **Inventory (0x07)**
   - List of blocks/transactions

### Node Types

1. **Full Node**
   - Stores complete blockchain
   - Validates all blocks and transactions
   - Relays blocks and transactions

2. **Mining Node**
   - Full node capabilities
   - Participates in mining
   - Creates new blocks

3. **Light Node**
   - Stores block headers only
   - Validates using SPV
   - Limited transaction validation

## Security Considerations

1. **Double Spending Prevention**
   - UTXO model
   - Transaction validation
   - Block confirmation

2. **51% Attack Protection**
   - Proof of work requirement
   - Longest chain rule
   - Block confirmation depth

3. **Cross-Chain Security**
   - Special validation for Antion transfers
   - Atomic transactions
   - Verification of both chains

## Future Enhancements

1. **Smart Contracts**
   - Script language
   - Contract deployment
   - Execution environment

2. **Privacy Features**
   - Ring signatures
   - Confidential transactions
   - Zero-knowledge proofs

3. **Scalability Improvements**
   - Sharding
   - Layer 2 solutions
   - Off-chain transactions 
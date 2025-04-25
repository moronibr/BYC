# Brigham Young Chain User Guide

## Introduction

Welcome to the Brigham Young Chain (Young Chain), a dual-blockchain system inspired by the Nephite monetary system. This guide will help you understand how to use the blockchain, mine coins, and participate in the network.

## Getting Started

### Installation

1. **Prerequisites**
   - Go 1.16 or higher
   - Git
   - Basic understanding of blockchain concepts

2. **Download and Install**
   ```bash
   git clone https://github.com/yourusername/brigham-young-chain.git
   cd brigham-young-chain
   go mod download
   ```

### Running a Node

1. **Full Node**
   ```bash
   go run cmd/youngchain/main.go -type=full -port=8333
   ```
   - Stores complete blockchain
   - Validates all transactions
   - Relays blocks and transactions

2. **Mining Node**
   ```bash
   go run cmd/youngchain/main.go -type=miner -port=8334
   ```
   - Participates in mining
   - Creates new blocks
   - Earns mining rewards

3. **Light Node**
   ```bash
   go run cmd/youngchain/main.go -type=light -port=8335
   ```
   - Minimal storage requirements
   - Basic transaction validation
   - Suitable for mobile devices

## Understanding the Coin System

### Mineable Coins

1. **Leah**
   - Base unit
   - Easiest to mine
   - Present in both chains

2. **Shiblum**
   - 2x Leah value
   - Medium difficulty
   - Present in both chains

3. **Shiblon**
   - 4x Leah value
   - Highest difficulty
   - Present in both chains

### Derived Coins

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
   - Created when a Golden Block is completed
   - Limited to 11 million
   - Cannot be transferred between chains

2. **Manasseh**
   - Created when a Silver Block is completed
   - Limited to 11 million
   - Cannot be transferred between chains

3. **Joseph**
   - Created by combining 1 Ephraim + 1 Manasseh
   - Limited to 11 million
   - Highest value coin

## Mining Guide

### Getting Started with Mining

1. **Hardware Requirements**
   - CPU: Multi-core processor recommended
   - RAM: 4GB minimum, 8GB recommended
   - Storage: 100GB for full node, 10GB for light node

2. **Mining Software Setup**
   ```bash
   go run cmd/youngchain/main.go -type=miner -port=8334
   ```

3. **Selecting Mining Difficulty**
   - Leah: Easiest, suitable for basic hardware
   - Shiblum: Medium, requires better hardware
   - Shiblon: Hardest, requires dedicated mining hardware

### Mining Rewards

1. **Leah Mining**
   - Reward: 50 Leah coins per block
   - Block time: ~10 minutes

2. **Shiblum Mining**
   - Reward: 100 Shiblum coins per block
   - Block time: ~10 minutes

3. **Shiblon Mining**
   - Reward: 200 Shiblon coins per block
   - Block time: ~10 minutes

## Transactions

### Creating Transactions

1. **Standard Transaction**
   - Transfer coins within the same chain
   - Specify coin type (Leah, Shiblum, etc.)
   - Include recipient address and amount

2. **Cross-Chain Transaction**
   - Only possible with Antion coins
   - Converts to 3 Shiblons on the other chain
   - Requires special validation

3. **Special Coin Transaction**
   - Ephraim/Manasseh: Cannot be transferred between chains
   - Joseph: Can be transferred but cannot be split

### Transaction Fees

- Based on transaction size
- Paid in the same coin type as the transaction
- Minimum fee: 0.0001 of the coin type

## Wallet Management

### Creating a Wallet

1. **Generate Keys**
   ```bash
   youngchain-cli createwallet
   ```

2. **Backup Keys**
   - Save private key securely
   - Create multiple backups
   - Store in different locations

### Managing Addresses

1. **Generate Address**
   ```bash
   youngchain-cli getnewaddress
   ```

2. **View Balance**
   ```bash
   youngchain-cli getbalance
   ```

## Network Participation

### Connecting to Peers

1. **Add Peer**
   ```bash
   youngchain-cli addnode <ip>:<port>
   ```

2. **List Peers**
   ```bash
   youngchain-cli getpeerinfo
   ```

### Syncing the Blockchain

1. **Check Sync Status**
   ```bash
   youngchain-cli getblockchaininfo
   ```

2. **Force Resync**
   ```bash
   youngchain-cli resync
   ```

## Troubleshooting

### Common Issues

1. **Node Not Syncing**
   - Check internet connection
   - Verify firewall settings
   - Ensure sufficient disk space

2. **Mining Not Working**
   - Verify hardware meets requirements
   - Check mining difficulty settings
   - Ensure proper network connection

3. **Transaction Failed**
   - Verify sufficient funds
   - Check transaction fee
   - Ensure correct recipient address

### Getting Help

- Join our community forum
- Check the documentation
- Contact support team

## Security Best Practices

1. **Wallet Security**
   - Use strong passwords
   - Enable two-factor authentication
   - Regular backups

2. **Node Security**
   - Keep software updated
   - Use firewall
   - Limit RPC access

3. **Transaction Security**
   - Verify addresses carefully
   - Use appropriate fees
   - Confirm transaction details

## Conclusion

The Brigham Young Chain offers a unique dual-blockchain system with various coin types and mining options. By following this guide, you can participate in the network, mine coins, and conduct transactions securely. 
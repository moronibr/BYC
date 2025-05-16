# Brigham Young Chain (BYC)

Brigham Young Chain is a blockchain implementation with a unique three-tier mining system featuring Leah, Shiblum, and Shiblon coins, this blockchain is based on the Nephites monetary system found in The Book of Mormon, Alma 11.

## Quick Start

### Download Binaries

Download the appropriate binary for your operating system from the [releases page](https://github.com/yourusername/byc/releases).

### Running a BYC Node

To run a BYC node:

```bash
# Linux/macOS
./bycnode_linux_amd64_0.1.0 -type full -port 8333

# Windows
bycnode_windows_amd64_0.1.0.exe -type full -port 8333
```

Options:
- `-type`: Node type (full, miner, light)
- `-port`: Port to listen on (default: 8333)

### Running a BYC Miner

To run a BYC miner:

```bash
# Linux/macOS
./bycminer_linux_amd64_0.1.0 -node localhost:8333 -coin leah -threads 4

# Windows
bycminer_windows_amd64_0.1.0.exe -node localhost:8333 -coin leah -threads 4
```

Options:
- `-node`: BYC node address to connect to (default: localhost:8333)
- `-coin`: Coin type to mine (leah, shiblum, shiblon)
- `-threads`: Number of mining threads (default: number of CPU cores)
- `-type`: Mining type (solo, pool)
- `-pool`: Pool address (required for pool mining)
- `-wallet`: Wallet address to receive mining rewards (required for pool mining)

## Building from Source

If you prefer to build from source:

```bash
# Clone the repository
git clone https://github.com/yourusername/byc.git
cd byc

# Build the node
go build -o bycnode ./cmd/youngchain

# Build the miner
go build -o bycminer ./cmd/bycminer
```

## Development

### Prerequisites

- Go 1.16 or higher
- Git

### Project Structure

- `cmd/`: Command-line applications
  - `youngchain/`: BYC node implementation
  - `bycminer/`: BYC miner implementation
- `internal/`: Internal packages
  - `core/`: Core blockchain functionality
  - `network/`: Network communication
  - `storage/`: Data storage

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Features

### Dual Blockchain System
- Golden Block Chain
- Silver Block Chain
- Independent but interconnected economies

### Mining System
- Three-tier mining difficulty:
  - Leah (easiest)
  - Shiblum (medium)
  - Shiblon (hardest)
- Progressive difficulty adjustment
- Proof of Work consensus

### Coin System
#### Mineable Coins (Both Chains)
- Leah (base unit)
- Shiblum (2x difficulty)
- Shiblon (4x difficulty)

#### Gold-based Derived Units
- Senine
- Seon
- Shum
- Limnah
- Antion (cross-chain transfer)

#### Silver-based Derived Units
- Senum
- Amnor
- Ezrom
- Onti

#### Special Block Completion Coins
- Ephraim (Golden Block completion)
- Manasseh (Silver Block completion)
- Joseph (1 Ephraim + 1 Manasseh)

### Supply Limits
- 15 million Ephraim coins (only 3 million can be converted to Joseph)
- 15 million Manasseh coins (only 3 million can be converted to Joseph)
- 3 million Joseph coins (created by combining 1 Ephraim + 1 Manasseh)

### Network Features
- P2P network
- Node types (full, miner, light)
- Transparent transactions
- No pre-mining
- Open source

## Getting Started

### Prerequisites
- Go 1.16 or higher
- Git

### Installation
```bash
git clone https://github.com/yourusername/brigham-young-chain.git
cd brigham-young-chain
go mod download
```

### Running a Node
```bash
# Full node
go run cmd/youngchain/main.go -type=full -port=8333

# Mining node
go run cmd/youngchain/main.go -type=miner -port=8334

# Light node
go run cmd/youngchain/main.go -type=light -port=8335
```

## Architecture

### Block Structure
- Block header
- Transaction list
- Mining information
- Chain-specific data

### Mining Process
1. Select mining difficulty (Leah, Shiblum, Shiblon)
2. Create block with transactions
3. Perform proof of work
4. Broadcast to network

### Cross-Chain Transfers
- Only Antion can move between chains
- 1 Antion = 3 Shiblons
- Special validation for cross-chain transactions

## Contributing
Contributions are welcome! Please read our contributing guidelines before submitting pull requests.

## Acknowledgments
- The Book of Mormon was translated by the power of God by The Prophet Joseph Smith Junior
- This block chain has no link to The Church of Jesus Christ of Latter Day Saints
- Inspired by the Nephite monetary system
- Built on blockchain principles
- Community-driven development # BYC
# BYC

## Monitoring System

### Overview
The monitoring system (`internal/core/monitoring`) collects and aggregates metrics from various parts of the application, including system metrics (CPU, memory, disk), blockchain metrics, and application metrics. It also supports alerting based on configurable thresholds.

### Adding New Collectors
To add a new collector:
1. Implement the `MetricCollector` interface:
   ```go
   type MyCollector struct{}
   func (c *MyCollector) Collect() (Metric, error) {
       // Your custom logic here
       return Metric{
           Type:      "my_metric",
           Value:     42,
           Timestamp: time.Now(),
           Labels:    map[string]string{"info": "custom"},
       }, nil
   }
   ```
2. Register your collector in the monitor:
   ```go
   monitor.collectors["my_metric"] = &MyCollector{}
   ```

### Adding New Aggregators
To add a new aggregator:
1. Implement the `MetricAggregator` interface:
   ```go
   type MyAggregator struct{}
   func (a *MyAggregator) Aggregate(metrics []Metric) Metric {
       // Your custom aggregation logic here
       return Metric{
           Type:      metrics[0].Type,
           Value:     /* aggregated value */,
           Timestamp: time.Now(),
           Labels:    metrics[0].Labels,
       }
   }
   ```
2. Register your aggregator in the monitor:
   ```go
   monitor.aggregators["my_metric"] = &MyAggregator{}
   ```

### Running and Interpreting Tests
The monitoring system includes unit tests for alerting logic. To run the tests:
```sh
go test ./internal/core/monitoring
```
The tests verify that alerts are triggered correctly for various metrics (CPU, memory, disk, block time, mempool size, network peers). If a test fails, check the alert thresholds and logic in `checkAlerts`.

### Integrating with Prometheus/Grafana
To expose metrics for Prometheus:
1. Add a `/metrics` HTTP endpoint in your main application:
   ```go
   import (
       "github.com/prometheus/client_golang/prometheus/promhttp"
       "net/http"
   )
   func main() {
       // ... your setup ...
       http.Handle("/metrics", promhttp.Handler())
       go http.ListenAndServe(":2112", nil)
       // ... rest of your app ...
   }
   ```
2. Update your collectors to update Prometheus metrics as you collect them.
3. Configure Prometheus to scrape your `/metrics` endpoint.
4. Use Grafana to visualize your Prometheus data.

## Security Enhancements

### Overview
The following security features are planned for implementation:

- **Key Management System**: Secure storage and management of cryptographic keys.
- **Secure Address Generation**: Robust address generation for transactions.
- **Transaction Signing Improvements**: Enhanced transaction signing for better security.
- **Rate Limiting for RPC Endpoints**: Prevent abuse by limiting request rates.
- **Authentication System for RPC Calls**: Secure RPC access with authentication.

### Key Management System
The key management system will securely store and manage cryptographic keys used for transactions and authentication. This includes:
- Secure key storage (e.g., using hardware security modules or encrypted storage).
- Key rotation and backup mechanisms.
- Access control for key usage.

### Secure Address Generation
Secure address generation ensures that addresses are cryptographically secure and resistant to attacks. This includes:
- Using strong cryptographic algorithms (e.g., SHA-256, RIPEMD-160).
- Implementing address checksums to prevent typos.
- Supporting hierarchical deterministic (HD) wallets for better key management.

### Transaction Signing Improvements
Transaction signing improvements will enhance the security of transactions by:
- Using robust cryptographic algorithms (e.g., ECDSA, EdDSA).
- Implementing multi-signature (multisig) support for shared control.
- Adding replay protection and nonce management.

### Rate Limiting for RPC Endpoints
Rate limiting will prevent abuse of RPC endpoints by:
- Limiting the number of requests per IP address or user.
- Implementing token bucket or leaky bucket algorithms for fair rate limiting.
- Configuring rate limits based on endpoint sensitivity.

### Authentication System for RPC Calls
The authentication system will secure RPC access by:
- Implementing token-based authentication (e.g., JWT).
- Supporting role-based access control (RBAC).
- Providing secure password storage and management.

## Performance Optimization

### Overview
The following performance improvements are planned for implementation:

- **Caching Mechanisms**: Implement proper caching to reduce redundant computations and improve response times.
- **Database Indexing**: Add indexes to frequently queried fields to speed up database operations.
- **Batch Processing**: Optimize batch operations to reduce overhead and improve throughput.
- **Memory Management**: Improve memory usage and garbage collection to reduce latency and overhead.
- **Connection Pooling**: Implement connection pooling to reuse database connections and reduce connection overhead.

### Caching Mechanisms
Caching will be implemented to store frequently accessed data in memory, reducing the need for repeated computations or database queries. This includes:
- In-memory caching (e.g., using a library like `bigcache` or `go-cache`).
- Distributed caching (e.g., using Redis or Memcached) for multi-node deployments.
- Cache invalidation strategies to ensure data consistency.

### Database Indexing
Database indexing will be added to improve query performance by:
- Creating indexes on frequently queried fields (e.g., transaction IDs, block heights).
- Using composite indexes for complex queries.
- Regularly analyzing and optimizing index usage.

### Batch Processing
Batch processing will be optimized to reduce overhead by:
- Grouping similar operations (e.g., multiple transactions) into a single batch.
- Using bulk insert/update operations where possible.
- Implementing parallel processing for independent batches.

### Memory Management
Memory management improvements will focus on:
- Reducing memory leaks and excessive allocations.
- Optimizing garbage collection settings.
- Using object pooling for frequently created objects.

### Connection Pooling
Connection pooling will be implemented to reuse database connections, reducing the overhead of establishing new connections. This includes:
- Configuring a connection pool with appropriate min/max connections.
- Implementing connection timeouts and retry logic.
- Monitoring connection pool health and performance.

## Backup and Recovery

### Overview
The backup system (`internal/core/monitoring/backup.go`) is designed to ensure data integrity and availability through automated backups, verification, and recovery procedures. The following features are planned for implementation:

- **Automated Backup Scheduling**: Regularly scheduled backups to ensure data is consistently backed up.
- **Backup Verification**: Automated verification of backup integrity to ensure data can be restored.
- **Recovery Procedures**: Clear procedures for restoring data from backups in case of data loss.
- **Disaster Recovery Plans**: Comprehensive plans for handling catastrophic failures, including data center outages or hardware failures.

### Automated Backup Scheduling
Automated backups will be scheduled to run at regular intervals (e.g., daily, weekly) to ensure data is consistently backed up. This includes:
- Configurable backup schedules (e.g., using cron jobs or a scheduling library).
- Incremental backups to reduce storage and bandwidth usage.
- Backup retention policies to manage storage costs.

### Backup Verification
Backup verification will ensure that backups are valid and can be restored. This includes:
- Automated integrity checks (e.g., checksums or hash verification).
- Regular test restores to validate backup data.
- Logging and alerting for failed verifications.

### Recovery Procedures
Recovery procedures will provide clear steps for restoring data from backups. This includes:
- Step-by-step recovery guides for different scenarios (e.g., single file, full database).
- Role-based access control for recovery operations.
- Regular recovery drills to ensure procedures are effective.

### Disaster Recovery Plans
Disaster recovery plans will outline procedures for handling catastrophic failures. This includes:
- Data center redundancy and failover strategies.
- Hardware and software recovery procedures.
- Communication and escalation plans for major incidents.

## Configuration Management

### Overview
The configuration management system is designed to handle environment-specific configurations, dynamic updates, validation, and secure storage. The following features are planned for implementation:

- **Environment-Specific Configurations**: Support for different configurations based on the environment (e.g., development, staging, production).
- **Dynamic Configuration Updates**: Ability to update configurations at runtime without restarting the application.
- **Configuration Validation**: Automated validation of configuration values to ensure they meet required criteria.
- **Secure Configuration Storage**: Secure storage of sensitive configuration values (e.g., using environment variables, vaults, or encrypted files).

### Environment-Specific Configurations
Environment-specific configurations will allow the application to use different settings based on the current environment. This includes:
- Configuration files or environment variables for each environment (e.g., `config.dev.yaml`, `config.prod.yaml`).
- A configuration loader that selects the appropriate configuration based on the environment.
- Default fallback values for missing configurations.

### Dynamic Configuration Updates
Dynamic configuration updates will enable the application to update its configuration at runtime. This includes:
- A configuration manager that listens for changes (e.g., file watchers or API endpoints).
- Hot-reloading of configuration values without restarting the application.
- Logging and alerting for configuration changes.

### Configuration Validation
Configuration validation will ensure that configuration values are valid and meet required criteria. This includes:
- Automated validation of configuration values (e.g., using a validation library or custom logic).
- Clear error messages for invalid configurations.
- Preventing the application from starting with invalid configurations.

### Secure Configuration Storage
Secure configuration storage will protect sensitive configuration values. This includes:
- Using environment variables for sensitive values (e.g., API keys, passwords).
- Integrating with secure vaults (e.g., HashiCorp Vault, AWS Secrets Manager).
- Encrypting configuration files for additional security.

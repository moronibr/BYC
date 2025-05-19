# Configuration Guide

This guide provides detailed information about configuring your Book of Mormon Coin (BMC) node. All configuration options are available through the `config.json` file or environment variables.

## Configuration File

The configuration file is located at `~/.bmc/config.json` by default. You can specify a different location using the `--config` flag.

## Network Configuration

```json
{
  "network": {
    "type": "mainnet",
    "rpc_url": "http://localhost:8545",
    "p2p_port": 30303,
    "bootstrap_nodes": [],
    "block_time": "15s",
    "difficulty": 2,
    "max_block_size": 1048576,
    "max_connections": 50,
    "sync_timeout": "5m",
    "reconnect_delay": "30s"
  }
}
```

### Network Settings

| Setting | Description | Default | Environment Variable |
|---------|-------------|---------|---------------------|
| type | Network type (mainnet/testnet/devnet) | mainnet | BMC_NETWORK_TYPE |
| rpc_url | RPC server URL | http://localhost:8545 | BMC_RPC_URL |
| p2p_port | P2P network port | 30303 | BMC_P2P_PORT |
| bootstrap_nodes | List of bootstrap nodes | [] | BMC_BOOTSTRAP_NODES |
| block_time | Time between blocks | 15s | BMC_BLOCK_TIME |
| difficulty | Mining difficulty | 2 | BMC_DIFFICULTY |
| max_block_size | Maximum block size in bytes | 1048576 | BMC_MAX_BLOCK_SIZE |
| max_connections | Maximum P2P connections | 50 | BMC_MAX_CONNECTIONS |
| sync_timeout | Blockchain sync timeout | 5m | BMC_SYNC_TIMEOUT |
| reconnect_delay | P2P reconnect delay | 30s | BMC_RECONNECT_DELAY |

## Fee Configuration

```json
{
  "fee": {
    "base_fee": 0.001,
    "size_multiplier": 0.0001,
    "priority_multiplier": 0.01,
    "min_fee": 0.0001,
    "max_fee": 1.0,
    "fee_update_interval": "1h"
  }
}
```

### Fee Settings

| Setting | Description | Default | Environment Variable |
|---------|-------------|---------|---------------------|
| base_fee | Base transaction fee | 0.001 | BMC_BASE_FEE |
| size_multiplier | Fee multiplier per byte | 0.0001 | BMC_SIZE_MULTIPLIER |
| priority_multiplier | Priority fee multiplier | 0.01 | BMC_PRIORITY_MULTIPLIER |
| min_fee | Minimum transaction fee | 0.0001 | BMC_MIN_FEE |
| max_fee | Maximum transaction fee | 1.0 | BMC_MAX_FEE |
| fee_update_interval | Fee update interval | 1h | BMC_FEE_UPDATE_INTERVAL |

## Security Configuration

```json
{
  "security": {
    "encryption_algorithm": "aes-256-gcm",
    "key_derivation_cost": 32768,
    "key_rotation_interval": "720h",
    "max_login_attempts": 5,
    "session_timeout": "30m",
    "password_min_length": 12,
    "require_special_chars": true,
    "require_numbers": true,
    "require_uppercase": true
  }
}
```

### Security Settings

| Setting | Description | Default | Environment Variable |
|---------|-------------|---------|---------------------|
| encryption_algorithm | Wallet encryption algorithm | aes-256-gcm | BMC_ENCRYPTION_ALGORITHM |
| key_derivation_cost | Key derivation iterations | 32768 | BMC_KEY_DERIVATION_COST |
| key_rotation_interval | Key rotation interval | 720h | BMC_KEY_ROTATION_INTERVAL |
| max_login_attempts | Maximum login attempts | 5 | BMC_MAX_LOGIN_ATTEMPTS |
| session_timeout | Session timeout | 30m | BMC_SESSION_TIMEOUT |
| password_min_length | Minimum password length | 12 | BMC_PASSWORD_MIN_LENGTH |
| require_special_chars | Require special characters | true | BMC_REQUIRE_SPECIAL_CHARS |
| require_numbers | Require numbers | true | BMC_REQUIRE_NUMBERS |
| require_uppercase | Require uppercase letters | true | BMC_REQUIRE_UPPERCASE |

## Environment Configuration

```json
{
  "environment": {
    "log_level": "info",
    "data_dir": "data",
    "backup_dir": "backups",
    "temp_dir": "temp",
    "max_log_size": 104857600,
    "max_log_files": 5,
    "debug_mode": false,
    "metrics_port": 9090,
    "enable_profiling": false
  }
}
```

### Environment Settings

| Setting | Description | Default | Environment Variable |
|---------|-------------|---------|---------------------|
| log_level | Logging level | info | BMC_LOG_LEVEL |
| data_dir | Data directory | data | BMC_DATA_DIR |
| backup_dir | Backup directory | backups | BMC_BACKUP_DIR |
| temp_dir | Temporary directory | temp | BMC_TEMP_DIR |
| max_log_size | Maximum log file size | 104857600 | BMC_MAX_LOG_SIZE |
| max_log_files | Maximum number of log files | 5 | BMC_MAX_LOG_FILES |
| debug_mode | Enable debug mode | false | BMC_DEBUG_MODE |
| metrics_port | Metrics server port | 9090 | BMC_METRICS_PORT |
| enable_profiling | Enable profiling | false | BMC_ENABLE_PROFILING |

## Configuration Examples

### Development Environment

```json
{
  "network": {
    "type": "devnet",
    "rpc_url": "http://localhost:8545",
    "p2p_port": 30304
  },
  "environment": {
    "log_level": "debug",
    "debug_mode": true,
    "enable_profiling": true
  }
}
```

### Production Environment

```json
{
  "network": {
    "type": "mainnet",
    "p2p_port": 30303,
    "max_connections": 100
  },
  "security": {
    "key_derivation_cost": 65536,
    "password_min_length": 16
  },
  "environment": {
    "log_level": "info",
    "max_log_files": 10
  }
}
```

### High-Security Environment

```json
{
  "security": {
    "key_derivation_cost": 131072,
    "key_rotation_interval": "168h",
    "max_login_attempts": 3,
    "session_timeout": "15m",
    "password_min_length": 16,
    "require_special_chars": true,
    "require_numbers": true,
    "require_uppercase": true
  }
}
```

## Environment Variables

All configuration options can be set using environment variables. The format is `BMC_<SECTION>_<SETTING>` in uppercase.

Example:
```bash
export BMC_NETWORK_TYPE=mainnet
export BMC_RPC_URL=http://localhost:8545
export BMC_SECURITY_PASSWORD_MIN_LENGTH=16
```

## Configuration Validation

The configuration is validated on startup. Common validation errors include:

- Invalid network type
- Invalid port numbers
- Invalid time durations
- Invalid file paths
- Invalid security parameters

## Best Practices

1. **Security**
   - Use strong passwords
   - Enable all security features in production
   - Regularly rotate keys
   - Use secure RPC endpoints

2. **Performance**
   - Adjust max_connections based on resources
   - Configure appropriate log levels
   - Enable profiling in development

3. **Monitoring**
   - Enable metrics collection
   - Configure appropriate log rotation
   - Set up monitoring alerts

## Troubleshooting

If you encounter configuration issues:

1. Check the configuration file syntax
2. Verify environment variables
3. Check file permissions
4. Review the logs
5. Validate the configuration

For more help, see the [Troubleshooting Guide](./troubleshooting.md). 
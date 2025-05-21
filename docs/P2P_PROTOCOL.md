# BYC Blockchain P2P Protocol

This document describes the peer-to-peer protocol used in the BYC Blockchain network.

## Overview

The BYC Blockchain uses a custom P2P protocol for node discovery, block propagation, and transaction broadcasting. The protocol is designed to be efficient, secure, and resilient to network partitions.

## Network Architecture

### Node Types

1. **Full Nodes**
   - Maintain complete blockchain history
   - Participate in block validation
   - Relay transactions and blocks
   - Can mine new blocks

2. **Light Nodes**
   - Maintain only block headers
   - Can verify transactions
   - Cannot mine blocks
   - Rely on full nodes for data

3. **Bootstrap Nodes**
   - Maintain a list of known peers
   - Help new nodes join the network
   - Provide initial peer discovery

### Connection Types

1. **Inbound Connections**
   - Initiated by other nodes
   - Limited to prevent DoS attacks
   - Subject to rate limiting

2. **Outbound Connections**
   - Initiated by the local node
   - Used for block and transaction propagation
   - Maintained for peer discovery

## Protocol Messages

### Message Format

All messages are JSON-encoded with the following structure:

```json
{
    "type": "message_type",
    "from": "sender_id",
    "to": "recipient_id",
    "data": {},
    "timestamp": "2024-03-14T12:00:00Z"
}
```

### Message Types

1. **Peer Discovery**
   - `get_peers`: Request list of known peers
   - `peer_list`: Response with list of peers
   - `ping`: Check peer connectivity
   - `pong`: Response to ping

2. **Block Propagation**
   - `new_block`: Announce new block
   - `get_block`: Request block data
   - `block_data`: Block data response
   - `block_headers`: Block headers for light nodes

3. **Transaction Broadcasting**
   - `new_transaction`: Announce new transaction
   - `get_transaction`: Request transaction data
   - `transaction_data`: Transaction data response

4. **Network Management**
   - `version`: Exchange protocol versions
   - `disconnect`: Graceful disconnection
   - `error`: Error notification

## Peer Discovery

### Bootstrap Process

1. Node starts with list of bootstrap nodes
2. Connects to bootstrap nodes
3. Requests peer list from each bootstrap node
4. Connects to discovered peers
5. Repeats process periodically

### Peer Maintenance

1. Regular ping/pong to check connectivity
2. Remove inactive peers
3. Maintain minimum number of connections
4. Rotate connections to prevent stale network

## Network Partition Handling

### Detection

1. Monitor peer connectivity
2. Track block propagation delays
3. Detect fork conditions
4. Monitor network health metrics

### Recovery

1. Maintain connection to multiple network segments
2. Use bootstrap nodes to reconnect
3. Sync with longest chain
4. Replay transactions from orphaned blocks

## Security Measures

### Connection Security

1. Rate limiting per IP
2. Maximum connections per IP
3. Connection timeouts
4. Message size limits

### Message Validation

1. Verify message format
2. Check message size
3. Validate timestamps
4. Verify sender identity

### Anti-DoS Protection

1. Connection limits
2. Rate limiting
3. Resource quotas
4. Blacklisting

## Implementation Details

### Configuration

```go
type NetworkConfig struct {
    ListenPort     int
    BootstrapPeers []string
    MaxPeers       int
    PingInterval   time.Duration
    DialTimeout    time.Duration
    ReadTimeout    time.Duration
    WriteTimeout   time.Duration
}
```

### Peer Management

```go
type Peer struct {
    ID          string
    Address     string
    Port        int
    LastSeen    time.Time
    Latency     int64
    Version     string
    IsActive    bool
    IsBootstrap bool
}
```

### Message Handling

```go
type NetworkMessage struct {
    Type      string
    From      string
    To        string
    Data      json.RawMessage
    Timestamp time.Time
}
```

## Best Practices

1. **Connection Management**
   - Maintain diverse peer connections
   - Regular peer rotation
   - Monitor connection health
   - Implement backoff strategies

2. **Resource Usage**
   - Limit message queue size
   - Implement message prioritization
   - Monitor memory usage
   - Clean up inactive connections

3. **Error Handling**
   - Graceful disconnection
   - Automatic reconnection
   - Error logging and monitoring
   - Circuit breaker pattern

4. **Performance**
   - Use connection pooling
   - Implement message batching
   - Optimize message serialization
   - Monitor network metrics

## Monitoring and Metrics

### Key Metrics

1. **Network Health**
   - Active peer count
   - Connection success rate
   - Message latency
   - Network partition detection

2. **Performance**
   - Message throughput
   - Bandwidth usage
   - CPU usage
   - Memory usage

3. **Security**
   - Failed connection attempts
   - Rate limit violations
   - Malformed messages
   - Blacklisted peers

### Logging

1. **Connection Events**
   - New connections
   - Disconnections
   - Connection errors
   - Peer discovery

2. **Message Events**
   - Message received
   - Message sent
   - Message errors
   - Message validation

3. **System Events**
   - Resource usage
   - Performance metrics
   - Security events
   - Network status

## Future Improvements

1. **Protocol Enhancements**
   - Message compression
   - Binary protocol
   - Zero-copy message handling
   - Improved partition detection

2. **Security Features**
   - TLS encryption
   - Peer authentication
   - Message signing
   - Advanced DoS protection

3. **Performance Optimizations**
   - Connection multiplexing
   - Message pipelining
   - Improved peer selection
   - Better resource management 
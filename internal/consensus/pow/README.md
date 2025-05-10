# Proof of Work Implementation

This package provides a robust implementation of the Proof of Work (PoW) consensus mechanism with adaptive mining capabilities and production-ready features.

## Features

- Adaptive worker count based on system resources
- Circuit breaker pattern for failure protection
- Graceful shutdown mechanism
- Configurable mining parameters
- Comprehensive error handling
- Resource management
- Performance monitoring
- Difficulty adjustment

## Quick Start

```go
// Create a new block
block := &MyBlock{
    Data: []byte("transaction data"),
    // ... other block fields
}

// Create a new Proof of Work instance with default configuration
pow, err := NewProofOfWork(block, nil)
if err != nil {
    log.Fatal(err)
}

// Mine the block with adaptive worker count
nonce, hash, err := pow.RunAdaptive()
if err != nil {
    log.Fatal(err)
}

// Validate the proof of work
if pow.Validate() {
    fmt.Println("Block successfully mined!")
}
```

## Configuration

### Proof of Work Configuration

```go
config := &Config{
    MiningTimeout:   time.Minute * 5,  // Maximum time for mining
    MaxRetries:      3,                // Maximum retry attempts
    RecoveryEnabled: true,             // Enable panic recovery
    LoggingEnabled:  true,             // Enable detailed logging
}
```

### Circuit Breaker Configuration

```go
circuitBreakerConfig := &CircuitBreakerConfig{
    FailureThreshold:    5,            // Failures before opening circuit
    ResetTimeout:        time.Second * 30,  // Time before attempting reset
    HalfOpenMaxRequests: 3,            // Max requests in half-open state
}
```

## Performance Tuning

### Worker Count

The system automatically adjusts the number of workers based on:
- Available CPU cores
- Memory usage
- Current system load
- Historical performance

### Resource Limits

Set resource limits to prevent system overload:
```go
pow.rm.SetResourceLimits(
    80,  // Maximum CPU percentage
    70,  // Maximum memory percentage
)
```

### Mining Timeout

Adjust the mining timeout based on your network's difficulty:
```go
config := &Config{
    MiningTimeout: time.Minute * 10,  // Adjust based on network difficulty
}
```

## Error Handling

The implementation includes comprehensive error handling for:
- Resource exhaustion
- Mining timeouts
- Circuit breaker states
- System shutdown
- Invalid blocks

## Graceful Shutdown

```go
// Initiate graceful shutdown
err := pow.Shutdown()
if err != nil {
    log.Printf("Shutdown error: %v", err)
}
```

## Best Practices

1. **Resource Management**
   - Monitor system resources
   - Set appropriate resource limits
   - Use adaptive worker count

2. **Error Handling**
   - Always check returned errors
   - Implement proper error recovery
   - Use circuit breaker for failure protection

3. **Performance**
   - Monitor mining performance
   - Adjust difficulty as needed
   - Use appropriate timeout values

4. **Security**
   - Validate all input data
   - Protect against resource exhaustion
   - Implement proper access controls

## API Reference

### Types

#### Block Interface
```go
type Block interface {
    GetHash() []byte
    GetPrevHash() []byte
    GetTimestamp() int64
    GetData() []byte
    GetNonce() int64
    GetHeight() int64
    SetNonce(nonce int64)
    SetHash(hash []byte)
}
```

#### Config
```go
type Config struct {
    MiningTimeout   time.Duration
    MaxRetries      int
    RecoveryEnabled bool
    LoggingEnabled  bool
}
```

#### CircuitBreakerConfig
```go
type CircuitBreakerConfig struct {
    FailureThreshold    int
    ResetTimeout        time.Duration
    HalfOpenMaxRequests int
}
```

### Functions

#### NewProofOfWork
```go
func NewProofOfWork(block Block, config *Config) (*ProofOfWork, error)
```
Creates a new Proof of Work instance.

#### RunAdaptive
```go
func (pow *ProofOfWork) RunAdaptive() (int64, []byte, error)
```
Performs mining with adaptive worker count.

#### Validate
```go
func (pow *ProofOfWork) Validate() bool
```
Validates the proof of work.

#### Shutdown
```go
func (pow *ProofOfWork) Shutdown() error
```
Initiates graceful shutdown.

## Contributing

1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Create a new Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details. 
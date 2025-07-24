# Go UDP Kit

This project is a sophisticated, modular UDP (User Datagram Protocol) framework in Go, featuring automated packet reassembly, customizable retry and timeout mechanisms, and packet prioritization with Quality of Service (QoS) support.

## Table of Contents

1. Features
2. Installation
3. Usage
4. API Reference
5. Configuration
6. Examples
7. Metrics Integration
8. Contributing
9. License

## Features

- Automated packet reassembly for data integrity and order
- Customizable retry and timeout mechanisms
- Packet prioritization and QoS
- Bulk data transfer
- Compression and decompression
- Simple encryption and decryption
- Simulated packet loss for testing
- Real-time statistics tracking
- Prometheus metrics integration

## Installation

Install the Go UDP Kit using:

```
go get github.com/1cbyc/goudpkit
```

## Usage

Example usage of the Go UDP Kit:

```go
package main

import (
	"log"
	"net"
	"time"

	"github.com/1cbyc/goudpkit"
)

func main() {
	retryConfig := goudpkit.RetryConfig{
		MaxRetries:  3,
		BaseTimeout: time.Second,
		BackoffRate: 1.5,
	}

	qosConfig := goudpkit.QoSConfig{
		PriorityLevels: 3,
		PriorityQueues: make([][]goudpkit.Packet, 3),
	}

	bufferConfig := goudpkit.BufferConfig{
		MaxBufferSize: 1024,
		FlushInterval: 2 * time.Second,
	}

	kit, err := goudpkit.NewGoUDPKit(":8080", retryConfig, qosConfig, bufferConfig)
	if err != nil {
		log.Fatal(err)
	}
	defer kit.Close()

	packet := goudpkit.Packet{
		SequenceNumber: 1,
		Priority:       2,
		Data:           []byte("Hello, UDP!"),
	}
	destAddr, _ := net.ResolveUDPAddr("udp", "localhost:9090")
	err = kit.SendPacket(packet, destAddr)
	if err != nil {
		log.Printf("Error sending packet: %v", err)
	}

	for {
		data, addr, err := kit.ReceivePacket()
		if err != nil {
			log.Printf("Error receiving packet: %v", err)
			continue
		}
		log.Printf("Received from %v: %s", addr, string(data))
	}
}
```

## API Reference

### Types

#### GoUDPKit

The main struct for the Go UDP Kit. All methods are attached to this type.

#### Packet

```
type Packet struct {
	SequenceNumber uint32
	Priority       int
	Data           []byte
	RetryCount     int
	Timestamp      time.Time
}
```

#### RetryConfig

```
type RetryConfig struct {
	MaxRetries  int
	BaseTimeout time.Duration
	BackoffRate float64
}
```

#### QoSConfig

```
type QoSConfig struct {
	PriorityLevels int
	PriorityQueues [][]Packet
}
```

#### BufferConfig

```
type BufferConfig struct {
	MaxBufferSize int
	FlushInterval time.Duration
}
```

#### Stats

```
type Stats struct {
	PacketsSent     uint64
	PacketsReceived uint64
	PacketsDropped  uint64
	RetryCount      uint64
}
```

### Functions

- `NewGoUDPKit(addr string, retryConfig RetryConfig, qosConfig QoSConfig, bufferConfig BufferConfig) (*GoUDPKit, error)`
- `SendPacket(packet Packet, destAddr *net.UDPAddr) error`
- `ReceivePacket() ([]byte, *net.UDPAddr, error)`
- `SendBulkData(data []byte, packetSize int, destAddr *net.UDPAddr) error`
- `ReceiveBulkData(expectedPackets int) ([]byte, error)`
- `Compress(data []byte) []byte`
- `Decompress(data []byte) []byte`
- `EncryptData(data []byte, key []byte) []byte`
- `DecryptData(data []byte, key []byte) []byte`
- `SimulatePacketLoss(lossPercentage int)`
- `GetStats() Stats`
- `Close() error`
- `RegisterMetrics()`
- `ExportMetricsHTTP(addr string) error`

## Configuration

- **RetryConfig**: MaxRetries, BaseTimeout, BackoffRate
- **QoSConfig**: PriorityLevels, PriorityQueues
- **BufferConfig**: MaxBufferSize, FlushInterval

## Examples

### Sending a High-Priority Packet

```go
packet := goudpkit.Packet{
	SequenceNumber: 1,
	Priority:       2,
	Data:           []byte("Important message"),
}
destAddr, _ := net.ResolveUDPAddr("udp", "localhost:9090")
err := kit.SendPacket(packet, destAddr)
```

### Receiving and Reassembling Packets

```go
for {
	data, addr, err := kit.ReceivePacket()
	if err != nil {
		log.Printf("Error receiving packet: %v", err)
		continue
	}
	log.Printf("Reassembled message from %v: %s", addr, string(data))
}
```

## Metrics Integration

The kit provides built-in Prometheus metrics for packets sent, received, dropped, and retry count.

### Register and Export Metrics

```go
import "github.com/1cbyc/goudpkit"

func main() {
	goudpkit.RegisterMetrics()
	go goudpkit.ExportMetricsHTTP(":2112")
	// ... rest of your app
}
```

Visit `http://localhost:2112/metrics` to view real-time stats.

## Contributing

Contributions are welcome! Please submit a Pull Request.

## License

This project is licensed under the MIT License.

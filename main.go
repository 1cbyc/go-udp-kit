package main

import (
	"log"
	"net"
	"time"

	"github.com/1cbyc/go-udp-kit/goudpkit"
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

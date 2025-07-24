package main

import (
	"log"
	"net"
	"time"

	"github.com/1cbyc/udpframework"
)

func main() {
	retryConfig := udpframework.RetryConfig{
		MaxRetries:  3,
		BaseTimeout: time.Second,
		BackoffRate: 1.5,
	}

	qosConfig := udpframework.QoSConfig{
		PriorityLevels: 3,
		PriorityQueues: make([][]udpframework.Packet, 3),
	}

	bufferConfig := udpframework.BufferConfig{
		MaxBufferSize: 1024,
		FlushInterval: 2 * time.Second,
	}

	framework, err := udpframework.UdGo(":8080", retryConfig, qosConfig, bufferConfig)
	if err != nil {
		log.Fatal(err)
	}
	defer framework.Close()

	packet := udpframework.Packet{
		SequenceNumber: 1,
		Priority:       2,
		Data:           []byte("Hello, UDP!"),
	}
	destAddr, _ := net.ResolveUDPAddr("udp", "localhost:9090")
	err = framework.SendPacket(packet, destAddr)
	if err != nil {
		log.Printf("Error sending packet: %v", err)
	}

	for {
		data, addr, err := framework.ReceivePacket()
		if err != nil {
			log.Printf("Error receiving packet: %v", err)
			continue
		}
		log.Printf("Received from %v: %s", addr, string(data))
	}
}

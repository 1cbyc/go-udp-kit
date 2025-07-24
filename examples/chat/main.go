package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/1cbyc/udpframework"
)

func main() {
	mode := flag.String("mode", "receive", "Mode: send or receive")
	addr := flag.String("addr", ":9000", "UDP address to listen/send")
	msg := flag.String("msg", "hello", "Message to send (send mode)")
	flag.Parse()

	retryConfig := udpframework.RetryConfig{MaxRetries: 3, BaseTimeout: 100 * time.Millisecond, BackoffRate: 1.5}
	qosConfig := udpframework.QoSConfig{PriorityLevels: 1, PriorityQueues: make([][]udpframework.Packet, 1)}
	bufferConfig := udpframework.BufferConfig{MaxBufferSize: 1024, FlushInterval: 2 * time.Second}

	if *mode == "receive" {
		uf, err := udpframework.UdGo(*addr, retryConfig, qosConfig, bufferConfig)
		if err != nil {
			log.Fatal(err)
		}
		defer uf.Close()
		for {
			data, remote, err := uf.ReceivePacket()
			if err != nil {
				continue
			}
			if data != nil {
				fmt.Printf("Received from %v: %s\n", remote, string(data))
				os.Stdout.Sync()
			}
		}
	} else if *mode == "send" {
		uf, err := udpframework.UdGo(":0", retryConfig, qosConfig, bufferConfig)
		if err != nil {
			log.Fatal(err)
		}
		defer uf.Close()
		packet := udpframework.Packet{SequenceNumber: 1, Priority: 0, Data: []byte(*msg), Timestamp: time.Now()}
		destAddr, err := net.ResolveUDPAddr("udp", *addr)
		if err != nil {
			log.Fatal(err)
		}
		err = uf.SendPacket(packet, destAddr)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Message sent")
	}
}

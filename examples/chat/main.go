package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/1cbyc/go-udp-kit/goudpkit"
)

func main() {
	mode := flag.String("mode", "receive", "Mode: send or receive")
	addr := flag.String("addr", ":9000", "UDP address to listen/send")
	msg := flag.String("msg", "hello", "Message to send (send mode)")
	flag.Parse()

	retryConfig := goudpkit.RetryConfig{MaxRetries: 3, BaseTimeout: 100 * time.Millisecond, BackoffRate: 1.5}
	qosConfig := goudpkit.QoSConfig{PriorityLevels: 1, PriorityQueues: make([][]goudpkit.Packet, 1)}
	bufferConfig := goudpkit.BufferConfig{MaxBufferSize: 1024, FlushInterval: 2 * time.Second}

	if *mode == "receive" {
		kit, err := goudpkit.NewGoUDPKit(*addr, retryConfig, qosConfig, bufferConfig)
		if err != nil {
			log.Fatal(err)
		}
		defer kit.Close()
		for {
			data, remote, err := kit.ReceivePacket()
			if err != nil {
				continue
			}
			if data != nil {
				fmt.Printf("Received from %v: %s\n", remote, string(data))
				os.Stdout.Sync()
			}
		}
	} else if *mode == "send" {
		kit, err := goudpkit.NewGoUDPKit(":0", retryConfig, qosConfig, bufferConfig)
		if err != nil {
			log.Fatal(err)
		}
		defer kit.Close()
		packet := goudpkit.Packet{SequenceNumber: 1, Priority: 0, Data: []byte(*msg), Timestamp: time.Now()}
		destAddr, err := net.ResolveUDPAddr("udp", *addr)
		if err != nil {
			log.Fatal(err)
		}
		err = kit.SendPacket(packet, destAddr)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Message sent")
	}
}

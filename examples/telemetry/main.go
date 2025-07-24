package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/1cbyc/go-udp-kit/goudpkit"
)

func main() {
	mode := flag.String("mode", "receive", "Mode: send or receive")
	addr := flag.String("addr", ":9100", "UDP address to listen/send")
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
				fmt.Printf("Received metric from %v: %s\n", remote, string(data))
				os.Stdout.Sync()
			}
		}
	} else if *mode == "send" {
		kit, err := goudpkit.NewGoUDPKit(":0", retryConfig, qosConfig, bufferConfig)
		if err != nil {
			log.Fatal(err)
		}
		defer kit.Close()
		destAddr, err := net.ResolveUDPAddr("udp", *addr)
		if err != nil {
			log.Fatal(err)
		}
		rand.Seed(time.Now().UnixNano())
		for i := 0; i < 10; i++ {
			metric := "cpu=" + strconv.Itoa(rand.Intn(100)) + ",mem=" + strconv.Itoa(rand.Intn(10000))
			packet := goudpkit.Packet{SequenceNumber: uint32(i), Priority: 0, Data: []byte(metric), Timestamp: time.Now()}
			err = kit.SendPacket(packet, destAddr)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("Sent metric: %s\n", metric)
			time.Sleep(500 * time.Millisecond)
		}
	}
}

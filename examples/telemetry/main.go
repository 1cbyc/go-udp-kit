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

	"github.com/1cbyc/udpframework"
)

func main() {
	mode := flag.String("mode", "receive", "Mode: send or receive")
	addr := flag.String("addr", ":9100", "UDP address to listen/send")
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
				fmt.Printf("Received metric from %v: %s\n", remote, string(data))
				os.Stdout.Sync()
			}
		}
	} else if *mode == "send" {
		uf, err := udpframework.UdGo(":0", retryConfig, qosConfig, bufferConfig)
		if err != nil {
			log.Fatal(err)
		}
		defer uf.Close()
		destAddr, err := net.ResolveUDPAddr("udp", *addr)
		if err != nil {
			log.Fatal(err)
		}
		rand.Seed(time.Now().UnixNano())
		for i := 0; i < 10; i++ {
			metric := "cpu=" + strconv.Itoa(rand.Intn(100)) + ",mem=" + strconv.Itoa(rand.Intn(10000))
			packet := udpframework.Packet{SequenceNumber: uint32(i), Priority: 0, Data: []byte(metric), Timestamp: time.Now()}
			err = uf.SendPacket(packet, destAddr)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("Sent metric: %s\n", metric)
			time.Sleep(500 * time.Millisecond)
		}
	}
}

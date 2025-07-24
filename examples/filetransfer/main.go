package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"

	"github.com/1cbyc/udpframework"
)

func main() {
	mode := flag.String("mode", "receive", "Mode: send or receive")
	addr := flag.String("addr", ":9200", "UDP address to listen/send")
	filePath := flag.String("file", "file.dat", "File path to send/receive")
	flag.Parse()

	retryConfig := udpframework.RetryConfig{MaxRetries: 3, BaseTimeout: 100 * time.Millisecond, BackoffRate: 1.5}
	qosConfig := udpframework.QoSConfig{PriorityLevels: 1, PriorityQueues: make([][]udpframework.Packet, 1)}
	bufferConfig := udpframework.BufferConfig{MaxBufferSize: 4096, FlushInterval: 2 * time.Second}

	if *mode == "receive" {
		uf, err := udpframework.UdGo(*addr, retryConfig, qosConfig, bufferConfig)
		if err != nil {
			log.Fatal(err)
		}
		defer uf.Close()
		f, err := os.Create(*filePath)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		fmt.Println("Waiting for file...")
		data, err := uf.ReceiveBulkData(10000)
		if err != nil {
			log.Fatal(err)
		}
		_, err = f.Write(data)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("File received and written to", *filePath)
	} else if *mode == "send" {
		uf, err := udpframework.UdGo(":0", retryConfig, qosConfig, bufferConfig)
		if err != nil {
			log.Fatal(err)
		}
		defer uf.Close()
		f, err := os.Open(*filePath)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		fileInfo, err := f.Stat()
		if err != nil {
			log.Fatal(err)
		}
		fileSize := fileInfo.Size()
		data := make([]byte, fileSize)
		_, err = io.ReadFull(f, data)
		if err != nil {
			log.Fatal(err)
		}
		destAddr, err := net.ResolveUDPAddr("udp", *addr)
		if err != nil {
			log.Fatal(err)
		}
		err = uf.SendBulkData(data, 1024, destAddr)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("File sent")
	}
}

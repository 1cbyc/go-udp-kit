package goudpkit

import (
	"encoding/binary"
	"errors"
	"net"
	"sync"
	"time"

	"github.com/1cbyc/goudpkit"
)

type GoUDPKit struct {
	conn            *net.UDPConn
	reassemblyQueue map[uint32]*Packet
	retryConfig     RetryConfig
	qosConfig       QoSConfig
	bufferConfig    BufferConfig
	stats           Stats
	mu              sync.Mutex
}

type RetryConfig struct {
	MaxRetries  int
	BaseTimeout time.Duration
	BackoffRate float64
}

type QoSConfig struct {
	PriorityLevels int
	PriorityQueues [][]Packet
}

type BufferConfig struct {
	MaxBufferSize int
	FlushInterval time.Duration
}

type Stats struct {
	PacketsSent     uint64
	PacketsReceived uint64
	PacketsDropped  uint64
	RetryCount      uint64
}

func NewGoUDPKit(addr string, retryConfig RetryConfig, qosConfig QoSConfig, bufferConfig BufferConfig) (*GoUDPKit, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return nil, err
	}

	kit := &GoUDPKit{
		conn:            conn,
		reassemblyQueue: make(map[uint32]*Packet),
		retryConfig:     retryConfig,
		qosConfig:       qosConfig,
		bufferConfig:    bufferConfig,
		mu:              sync.Mutex{},
	}

	go kit.flushBufferPeriodically()

	return kit, nil
}

func (kit *GoUDPKit) SendPacket(packet Packet, destAddr *net.UDPAddr) error {
	kit.mu.Lock()
	defer kit.mu.Unlock()

	kit.qosConfig.PriorityQueues[packet.Priority] = append(kit.qosConfig.PriorityQueues[packet.Priority], packet)

	for i := kit.qosConfig.PriorityLevels - 1; i >= 0; i-- {
		for len(kit.qosConfig.PriorityQueues[i]) > 0 {
			p := kit.qosConfig.PriorityQueues[i][0]
			kit.qosConfig.PriorityQueues[i] = kit.qosConfig.PriorityQueues[i][1:]

			err := kit.sendWithRetry(p, destAddr)
			if err != nil {
				kit.stats.PacketsDropped++
				goudpkit.IncPacketsDropped()
				return err
			}
			kit.stats.PacketsSent++
			goudpkit.IncPacketsSent()
		}
	}

	return nil
}

func (kit *GoUDPKit) sendWithRetry(packet Packet, destAddr *net.UDPAddr) error {
	timeout := kit.retryConfig.BaseTimeout
	for retry := 0; retry < kit.retryConfig.MaxRetries; retry++ {
		_, err := kit.conn.WriteToUDP(packet.Data, destAddr)
		if err == nil {
			return nil
		}

		kit.stats.RetryCount++
		goudpkit.IncRetryCount()
		time.Sleep(timeout)
		timeout = time.Duration(float64(timeout) * kit.retryConfig.BackoffRate)
	}

	return errors.New("maximum retries reached")
}

func (kit *GoUDPKit) ReceivePacket() ([]byte, *net.UDPAddr, error) {
	buffer := make([]byte, 1500)
	n, remoteAddr, err := kit.conn.ReadFromUDP(buffer)
	if err != nil {
		return nil, nil, err
	}

	kit.stats.PacketsReceived++
	goudpkit.IncPacketsReceived()

	packet := Packet{
		SequenceNumber: binary.BigEndian.Uint32(buffer[:4]),
		Data:           buffer[4:n],
		Timestamp:      time.Now(),
	}

	kit.mu.Lock()
	defer kit.mu.Unlock()

	kit.reassemblyQueue[packet.SequenceNumber] = &packet

	assembledData := kit.tryReassemble()
	if assembledData != nil {
		return assembledData, remoteAddr, nil
	}

	return nil, remoteAddr, nil
}

func (kit *GoUDPKit) tryReassemble() []byte {
	var assembledData []byte
	expectedSeq := uint32(0)

	for {
		packet, exists := kit.reassemblyQueue[expectedSeq]
		if !exists {
			break
		}

		assembledData = append(assembledData, packet.Data...)
		delete(kit.reassemblyQueue, expectedSeq)
		expectedSeq++
	}

	if len(assembledData) > 0 {
		return assembledData
	}

	return nil
}

func (kit *GoUDPKit) flushBufferPeriodically() {
	ticker := time.NewTicker(kit.bufferConfig.FlushInterval)
	defer ticker.Stop()

	for range ticker.C {
		kit.flushBuffer()
	}
}

func (kit *GoUDPKit) flushBuffer() {
	kit.mu.Lock()
	defer kit.mu.Unlock()

	now := time.Now()
	for seq, packet := range kit.reassemblyQueue {
		if now.Sub(packet.Timestamp) > kit.bufferConfig.FlushInterval {
			delete(kit.reassemblyQueue, seq)
			kit.stats.PacketsDropped++
			goudpkit.IncPacketsDropped()
		}
	}
}

func (kit *GoUDPKit) GetStats() Stats {
	kit.mu.Lock()
	defer kit.mu.Unlock()
	return kit.stats
}

func (kit *GoUDPKit) Close() error {
	return kit.conn.Close()
}

package goudpkit

import (
	"encoding/binary"
	"errors"
	"net"
	"sync"
	"time"
)

type UDPConn interface {
	WriteToUDP(b []byte, addr *net.UDPAddr) (int, error)
	ReadFromUDP(b []byte) (int, *net.UDPAddr, error)
	SetReadDeadline(t time.Time) error
	Close() error
	LocalAddr() net.Addr
}

type GoUDPKit struct {
	conn            UDPConn
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

func NewGoUDPKit(addr string, retryConfig RetryConfig, qosConfig QoSConfig, bufferConfig BufferConfig, customConn ...UDPConn) (*GoUDPKit, error) {
	var conn UDPConn
	if len(customConn) > 0 && customConn[0] != nil {
		conn = customConn[0]
	} else {
		udpAddr, err := net.ResolveUDPAddr("udp", addr)
		if err != nil {
			return nil, err
		}
		conn, err = net.ListenUDP("udp", udpAddr)
		if err != nil {
			return nil, err
		}
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

func (kit *GoUDPKit) SendPacket(packet Packet, addr *net.UDPAddr) error {
	buf := make([]byte, 4+len(packet.Data))
	binary.BigEndian.PutUint32(buf[:4], packet.SequenceNumber)
	copy(buf[4:], packet.Data)
	_, err := kit.conn.WriteToUDP(buf, addr)
	if err == nil {
		kit.stats.PacketsSent++
	}
	return err
}

func (kit *GoUDPKit) sendWithRetry(packet Packet, destAddr *net.UDPAddr) error {
	timeout := kit.retryConfig.BaseTimeout
	for retry := 0; retry < kit.retryConfig.MaxRetries; retry++ {
		_, err := kit.conn.WriteToUDP(packet.Data, destAddr)
		if err == nil {
			return nil
		}

		kit.stats.RetryCount++
		time.Sleep(timeout)
		timeout = time.Duration(float64(timeout) * kit.retryConfig.BackoffRate)
	}

	return errors.New("maximum retries reached")
}

func (kit *GoUDPKit) ReceivePacket() ([]byte, *net.UDPAddr, error) {
	buf := make([]byte, 65535)
	n, addr, err := kit.conn.ReadFromUDP(buf)
	if err != nil {
		kit.stats.PacketsDropped++
		return nil, nil, err
	}
	if n < 4 {
		kit.stats.PacketsDropped++
		return nil, addr, errors.New("packet too short")
	}
	data := buf[4:n]
	kit.stats.PacketsReceived++
	return data, addr, nil
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

func (kit *GoUDPKit) Conn() UDPConn {
	return kit.conn
}

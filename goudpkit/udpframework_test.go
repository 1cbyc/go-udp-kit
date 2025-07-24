package goudpkit

import (
	"errors"
	"net"
	"testing"
	"time"
)

type mockUDPConn struct {
	writeCh chan []byte
	readCh  chan []byte
	closed  bool
}

func newMockUDPConn() *mockUDPConn {
	return &mockUDPConn{
		writeCh: make(chan []byte, 10),
		readCh:  make(chan []byte, 10),
	}
}

func (m *mockUDPConn) WriteToUDP(b []byte, _ *net.UDPAddr) (int, error) {
	if m.closed {
		return 0, errors.New("closed")
	}
	// Simulate real UDP: prepend 4-byte sequence number (big endian) if not already present
	if len(b) < 4 {
		return 0, errors.New("data too short for sequence number")
	}
	m.writeCh <- append([]byte{}, b...)
	return len(b), nil
}

func (m *mockUDPConn) ReadFromUDP(b []byte) (int, *net.UDPAddr, error) {
	if m.closed {
		return 0, nil, errors.New("closed")
	}
	data := <-m.writeCh
	copy(b, data)
	return len(data), &net.UDPAddr{}, nil
}

func (m *mockUDPConn) SetReadDeadline(t time.Time) error { return nil }
func (m *mockUDPConn) Close() error                      { m.closed = true; return nil }
func (m *mockUDPConn) LocalAddr() net.Addr               { return &net.UDPAddr{} }

func TestNewGoUDPKitInitialization(t *testing.T) {
	t.Parallel()
	retryConfig := RetryConfig{MaxRetries: 2, BaseTimeout: time.Millisecond * 10, BackoffRate: 1.2}
	qosConfig := QoSConfig{PriorityLevels: 2, PriorityQueues: make([][]Packet, 2)}
	bufferConfig := BufferConfig{MaxBufferSize: 128, FlushInterval: time.Millisecond * 50}
	mockConn := newMockUDPConn()
	kit, err := NewGoUDPKit(":0", retryConfig, qosConfig, bufferConfig, mockConn)
	if err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}
	defer kit.Close()
}

func TestSendAndReceivePacket(t *testing.T) {
	t.Parallel()
	retryConfig := RetryConfig{MaxRetries: 5, BaseTimeout: time.Millisecond * 20, BackoffRate: 1.5}
	qosConfig := QoSConfig{PriorityLevels: 2, PriorityQueues: make([][]Packet, 2)}
	bufferConfig := BufferConfig{MaxBufferSize: 128, FlushInterval: time.Millisecond * 50}
	mockConn := newMockUDPConn()
	recvKit, err := NewGoUDPKit(":0", retryConfig, qosConfig, bufferConfig, mockConn)
	if err != nil {
		t.Fatalf("Failed to initialize receiver: %v", err)
	}
	defer recvKit.Close()

	sendKit, err := NewGoUDPKit(":0", retryConfig, qosConfig, bufferConfig, mockConn)
	if err != nil {
		t.Fatalf("Failed to initialize sender: %v", err)
	}
	defer sendKit.Close()

	packet := Packet{SequenceNumber: 1, Priority: 1, Data: []byte("test-data"), Timestamp: time.Now()}
	err = sendKit.SendPacket(packet, &net.UDPAddr{})
	if err != nil {
		t.Fatalf("SendPacket failed: %v", err)
	}

	data, _, err := recvKit.ReceivePacket()
	if err != nil {
		t.Fatalf("ReceivePacket failed: %v", err)
	}
	if string(data) != "test-data" {
		t.Fatalf("Expected 'test-data', got '%s'", string(data))
	}
}

func TestCompressDecompress(t *testing.T) {
	t.Parallel()
	retryConfig := RetryConfig{MaxRetries: 1, BaseTimeout: time.Millisecond * 10, BackoffRate: 1.1}
	qosConfig := QoSConfig{PriorityLevels: 1, PriorityQueues: make([][]Packet, 1)}
	bufferConfig := BufferConfig{MaxBufferSize: 64, FlushInterval: time.Millisecond * 20}
	kit, err := NewGoUDPKit(":0", retryConfig, qosConfig, bufferConfig)
	if err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}
	defer kit.Close()

	original := []byte("aaabbbccccccdddddddeee")
	compressed := kit.Compress(original)
	decompressed := kit.Decompress(compressed)
	if string(decompressed) != string(original) {
		t.Fatalf("Compress/Decompress failed: got '%s', want '%s'", string(decompressed), string(original))
	}
}

func TestEncryptDecrypt(t *testing.T) {
	t.Parallel()
	retryConfig := RetryConfig{MaxRetries: 1, BaseTimeout: time.Millisecond * 10, BackoffRate: 1.1}
	qosConfig := QoSConfig{PriorityLevels: 1, PriorityQueues: make([][]Packet, 1)}
	bufferConfig := BufferConfig{MaxBufferSize: 64, FlushInterval: time.Millisecond * 20}
	kit, err := NewGoUDPKit(":0", retryConfig, qosConfig, bufferConfig)
	if err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}
	defer kit.Close()

	key := []byte("secret")
	plaintext := []byte("encrypt-this-data")
	cipher := kit.EncryptData(plaintext, key)
	decrypted := kit.DecryptData(cipher, key)
	if string(decrypted) != string(plaintext) {
		t.Fatalf("Encrypt/Decrypt failed: got '%s', want '%s'", string(decrypted), string(plaintext))
	}
}

func TestSendAndReceiveBulkData(t *testing.T) {
	t.Parallel()
	retryConfig := RetryConfig{MaxRetries: 5, BaseTimeout: time.Millisecond * 20, BackoffRate: 1.5}
	qosConfig := QoSConfig{PriorityLevels: 2, PriorityQueues: make([][]Packet, 2)}
	bufferConfig := BufferConfig{MaxBufferSize: 256, FlushInterval: time.Millisecond * 50}
	mockConn := newMockUDPConn()
	recvKit, err := NewGoUDPKit(":0", retryConfig, qosConfig, bufferConfig, mockConn)
	if err != nil {
		t.Fatalf("Failed to initialize receiver: %v", err)
	}
	defer recvKit.Close()

	sendKit, err := NewGoUDPKit(":0", retryConfig, qosConfig, bufferConfig, mockConn)
	if err != nil {
		t.Fatalf("Failed to initialize sender: %v", err)
	}
	defer sendKit.Close()

	bulkData := []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	packetSize := 10
	err = sendKit.SendBulkData(bulkData, packetSize, &net.UDPAddr{})
	if err != nil {
		t.Fatalf("SendBulkData failed: %v", err)
	}

	received, err := recvKit.ReceiveBulkData((len(bulkData) + packetSize - 1) / packetSize)
	if err != nil {
		t.Fatalf("ReceiveBulkData failed: %v", err)
	}
	if string(received) != string(bulkData) {
		t.Fatalf("Bulk data mismatch: got '%s', want '%s'", string(received), string(bulkData))
	}
}

func TestSimulatePacketLoss(t *testing.T) {
	t.Parallel()
	retryConfig := RetryConfig{MaxRetries: 1, BaseTimeout: time.Millisecond * 10, BackoffRate: 1.1}
	qosConfig := QoSConfig{PriorityLevels: 1, PriorityQueues: make([][]Packet, 1)}
	bufferConfig := BufferConfig{MaxBufferSize: 64, FlushInterval: time.Millisecond * 20}
	kit, err := NewGoUDPKit(":0", retryConfig, qosConfig, bufferConfig)
	if err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}
	defer kit.Close()

	initialDropped := kit.GetStats().PacketsDropped
	for i := 0; i < 100; i++ {
		kit.SimulatePacketLoss(100)
	}
	finalDropped := kit.GetStats().PacketsDropped
	if finalDropped-initialDropped != 100 {
		t.Fatalf("SimulatePacketLoss failed: expected 100 drops, got %d", finalDropped-initialDropped)
	}
}

func TestStatsTracking(t *testing.T) {
	t.Parallel()
	retryConfig := RetryConfig{MaxRetries: 5, BaseTimeout: time.Millisecond * 20, BackoffRate: 1.5}
	qosConfig := QoSConfig{PriorityLevels: 2, PriorityQueues: make([][]Packet, 2)}
	bufferConfig := BufferConfig{MaxBufferSize: 128, FlushInterval: time.Millisecond * 50}
	mockConn := newMockUDPConn()
	recvKit, err := NewGoUDPKit(":0", retryConfig, qosConfig, bufferConfig, mockConn)
	if err != nil {
		t.Fatalf("Failed to initialize receiver: %v", err)
	}
	defer recvKit.Close()

	sendKit, err := NewGoUDPKit(":0", retryConfig, qosConfig, bufferConfig, mockConn)
	if err != nil {
		t.Fatalf("Failed to initialize sender: %v", err)
	}
	defer sendKit.Close()

	packet := Packet{SequenceNumber: 1, Priority: 1, Data: []byte("stat-data"), Timestamp: time.Now()}
	err = sendKit.SendPacket(packet, &net.UDPAddr{})
	if err != nil {
		t.Fatalf("SendPacket failed: %v", err)
	}

	_, _, err = recvKit.ReceivePacket()
	if err != nil {
		t.Fatalf("ReceivePacket failed: %v", err)
	}

	s := sendKit.GetStats()
	if s.PacketsSent == 0 {
		t.Fatalf("Stats tracking failed: PacketsSent should be > 0")
	}
}

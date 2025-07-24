package goudpkit

import (
	"net"
	"time"
)

func (kit *GoUDPKit) SendBulkData(data []byte, packetSize int, destAddr *net.UDPAddr) error {
	totalPackets := (len(data) + packetSize - 1) / packetSize
	for i := 0; i < totalPackets; i++ {
		start := i * packetSize
		end := start + packetSize
		if end > len(data) {
			end = len(data)
		}

		packet := Packet{
			SequenceNumber: uint32(i),
			Priority:       0,
			Data:           data[start:end],
			Timestamp:      time.Now(),
		}

		err := kit.SendPacket(packet, destAddr)
		if err != nil {
			return err
		}
	}

	return nil
}

func (kit *GoUDPKit) ReceiveBulkData(expectedPackets int) ([]byte, error) {
	var bulkData []byte
	receivedPackets := 0

	for receivedPackets < expectedPackets {
		data, _, err := kit.ReceivePacket()
		if err != nil {
			return nil, err
		}

		if data != nil {
			bulkData = append(bulkData, data...)
			receivedPackets++
		}
	}

	return bulkData, nil
}

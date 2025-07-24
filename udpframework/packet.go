package goudpkit

import "time"

type Packet struct {
	SequenceNumber uint32
	Priority       int
	Data           []byte
	RetryCount     int
	Timestamp      time.Time
}

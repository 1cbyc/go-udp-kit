package goudpkit

import (
	"math/rand"
	"time"
)

func (kit *GoUDPKit) SimulatePacketLoss(lossPercentage int) {
	rand.Seed(time.Now().UnixNano())
	if rand.Intn(100) < lossPercentage {
		kit.stats.PacketsDropped++
		return
	}
}

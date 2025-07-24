package main

import (
	"fmt"

	"github.com/1cbyc/udpframework"
	"github.com/spf13/cobra"
)

func init() {
	var loss int
	var count int

	simCmd := &cobra.Command{
		Use:   "simulate-loss",
		Short: "Simulate packet loss and show dropped count",
		PreRun: func(cmd *cobra.Command, args []string) {
			loadConfig()
			if !cmd.Flags().Changed("loss") {
				loss = 50
			}
			if !cmd.Flags().Changed("count") {
				count = 100
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			retryConfig := udpframework.RetryConfig{MaxRetries: 1, BaseTimeout: 1, BackoffRate: 1.0}
			qosConfig := udpframework.QoSConfig{PriorityLevels: 1, PriorityQueues: make([][]udpframework.Packet, 1)}
			bufferConfig := udpframework.BufferConfig{MaxBufferSize: 1, FlushInterval: 1}
			uf, _ := udpframework.UdGo(":0", retryConfig, qosConfig, bufferConfig)
			defer uf.Close()
			before := uf.GetStats().PacketsDropped
			for i := 0; i < count; i++ {
				uf.SimulatePacketLoss(loss)
			}
			after := uf.GetStats().PacketsDropped
			fmt.Printf("Simulated %d packets with %d%% loss: %d dropped\n", count, loss, after-before)
		},
	}

	simCmd.Flags().IntVar(&loss, "loss", 0, "Loss percentage (0-100)")
	simCmd.Flags().IntVar(&count, "count", 0, "Number of packets to simulate")

	rootCmd.AddCommand(simCmd)
}

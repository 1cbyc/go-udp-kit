package main

import (
	"fmt"
	"time"

	"github.com/1cbyc/go-udp-kit/goudpkit"
	"github.com/spf13/cobra"
)

func init() {
	var addr string
	var timeout int

	statsCmd := &cobra.Command{
		Use:   "stats",
		Short: "Show UDP framework stats after receiving packets",
		PreRun: func(cmd *cobra.Command, args []string) {
			loadConfig()
			if !cmd.Flags().Changed("addr") {
				addr = cliConfig.Addr
			}
			if !cmd.Flags().Changed("timeout") {
				timeout = 10
			}
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			retryConfig := goudpkit.RetryConfig{MaxRetries: 3, BaseTimeout: 100 * time.Millisecond, BackoffRate: 1.5}
			qosConfig := goudpkit.QoSConfig{PriorityLevels: 1, PriorityQueues: make([][]goudpkit.Packet, 1)}
			bufferConfig := goudpkit.BufferConfig{MaxBufferSize: 1024, FlushInterval: 2 * time.Second}
			kit, err := goudpkit.NewGoUDPKit(addr, retryConfig, qosConfig, bufferConfig)
			if err != nil {
				return err
			}
			defer kit.Close()

			deadline := time.Now().Add(time.Duration(timeout) * time.Second)
			for {
				if timeout > 0 && time.Now().After(deadline) {
					break
				}
				kit.Conn().SetReadDeadline(time.Now().Add(1 * time.Second))
				_, _, err := kit.ReceivePacket()
				if err != nil {
					continue
				}
			}
			stats := kit.GetStats()
			fmt.Printf("Packets Sent: %d\nPackets Received: %d\nPackets Dropped: %d\nRetry Count: %d\n", stats.PacketsSent, stats.PacketsReceived, stats.PacketsDropped, stats.RetryCount)
			return nil
		},
	}

	statsCmd.Flags().StringVar(&addr, "addr", "", "UDP address to listen on")
	statsCmd.Flags().IntVar(&timeout, "timeout", 0, "Timeout in seconds")

	rootCmd.AddCommand(statsCmd)
}

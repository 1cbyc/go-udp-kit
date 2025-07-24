package main

import (
	"net"
	"time"

	"github.com/1cbyc/go-udp-kit/goudpkit"
	"github.com/spf13/cobra"
)

func init() {
	var addr string
	var data string
	var priority int
	var retries int
	var timeout int
	var backoff float64

	sendCmd := &cobra.Command{
		Use:   "send",
		Short: "Send a UDP packet",
		PreRun: func(cmd *cobra.Command, args []string) {
			loadConfig()
			if !cmd.Flags().Changed("addr") {
				addr = cliConfig.Addr
			}
			if !cmd.Flags().Changed("data") {
				data = cliConfig.Data
			}
			if !cmd.Flags().Changed("priority") {
				priority = cliConfig.Priority
			}
			if !cmd.Flags().Changed("retries") {
				retries = cliConfig.Retries
			}
			if !cmd.Flags().Changed("timeout") {
				timeout = cliConfig.Timeout
			}
			if !cmd.Flags().Changed("backoff") {
				backoff = cliConfig.Backoff
			}
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			retryConfig := goudpkit.RetryConfig{
				MaxRetries:  retries,
				BaseTimeout: time.Duration(timeout) * time.Millisecond,
				BackoffRate: backoff,
			}
			qosConfig := goudpkit.QoSConfig{
				PriorityLevels: priority + 1,
				PriorityQueues: make([][]goudpkit.Packet, priority+1),
			}
			bufferConfig := goudpkit.BufferConfig{
				MaxBufferSize: 1024,
				FlushInterval: 2 * time.Second,
			}
			kit, err := goudpkit.NewGoUDPKit(":0", retryConfig, qosConfig, bufferConfig)
			if err != nil {
				return err
			}
			defer kit.Close()

			packet := goudpkit.Packet{
				SequenceNumber: 1,
				Priority:       priority,
				Data:           []byte(data),
				Timestamp:      time.Now(),
			}
			destAddr, err := net.ResolveUDPAddr("udp", addr)
			if err != nil {
				return err
			}
			return kit.SendPacket(packet, destAddr)
		},
	}

	sendCmd.Flags().StringVar(&addr, "addr", "", "Destination UDP address")
	sendCmd.Flags().StringVar(&data, "data", "", "Data to send")
	sendCmd.Flags().IntVar(&priority, "priority", 0, "Packet priority")
	sendCmd.Flags().IntVar(&retries, "retries", 0, "Max retries")
	sendCmd.Flags().IntVar(&timeout, "timeout", 0, "Base timeout (ms)")
	sendCmd.Flags().Float64Var(&backoff, "backoff", 0, "Backoff rate")

	rootCmd.AddCommand(sendCmd)
}

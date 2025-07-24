package main

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/1cbyc/go-udp-kit/goudpkit"
	"github.com/spf13/cobra"
)

func init() {
	var addr string
	var timeout int

	receiveCmd := &cobra.Command{
		Use:   "receive",
		Short: "Receive UDP packets",
		PreRun: func(cmd *cobra.Command, args []string) {
			loadConfig()
			if !cmd.Flags().Changed("addr") {
				addr = cliConfig.Addr
			}
			if !cmd.Flags().Changed("timeout") {
				timeout = 0
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
				data, remote, err := kit.ReceivePacket()
				if err != nil {
					if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
						continue
					}
					return err
				}
				if data != nil {
					fmt.Printf("Received from %v: %s\n", remote, string(data))
					os.Stdout.Sync()
				}
			}
			return nil
		},
	}

	receiveCmd.Flags().StringVar(&addr, "addr", "", "UDP address to listen on")
	receiveCmd.Flags().IntVar(&timeout, "timeout", 0, "Timeout in seconds (0 for no timeout)")

	rootCmd.AddCommand(receiveCmd)
}

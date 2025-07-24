package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	shellCmd := &cobra.Command{
		Use:   "shell",
		Short: "Start an interactive UDP CLI shell",
		Run: func(cmd *cobra.Command, args []string) {
			loadConfig()
			scanner := bufio.NewScanner(os.Stdin)
			for {
				fmt.Print("udpcli> ")
				if !scanner.Scan() {
					break
				}
				line := strings.TrimSpace(scanner.Text())
				if line == "exit" || line == "quit" {
					break
				}
				fields := strings.Fields(line)
				if len(fields) == 0 {
					continue
				}
				switch fields[0] {
				case "send":
					args := []string{"send"}
					if len(fields) > 1 {
						args = append(args, fields[1:]...)
					}
					rootCmd.SetArgs(args)
					_ = rootCmd.Execute()
				case "receive":
					args := []string{"receive"}
					if len(fields) > 1 {
						args = append(args, fields[1:]...)
					}
					rootCmd.SetArgs(args)
					_ = rootCmd.Execute()
				case "stats":
					args := []string{"stats"}
					if len(fields) > 1 {
						args = append(args, fields[1:]...)
					}
					rootCmd.SetArgs(args)
					_ = rootCmd.Execute()
				case "simulate-loss":
					args := []string{"simulate-loss"}
					if len(fields) > 1 {
						args = append(args, fields[1:]...)
					}
					rootCmd.SetArgs(args)
					_ = rootCmd.Execute()
				default:
					fmt.Println("Unknown command. Available: send, receive, stats, simulate-loss, exit")
				}
			}
		},
	}
	rootCmd.AddCommand(shellCmd)
}

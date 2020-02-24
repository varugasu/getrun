package cmd

import (
	"fmt"
	"getrun/src/connection"
	"os"

	"github.com/spf13/cobra"
)

var (
	target   string
	port     int
	username string
	password string
	rootCmd  = &cobra.Command{
		Use:   "getrun",
		Short: "getrun returns the running configuration for a given Cisco host",
		Long:  "A package created to learn about golang. Feel free to use or modify it.",
		Run: func(cmd *cobra.Command, args []string) {
			h, err := connection.Connect(target, port, username, password)
			if err != nil {
				fmt.Println(err)
				return
			}
			defer h.Session.Close()
			defer h.Client.Close()
			out, err := h.SendCommand("sh run")
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println(out)
		},
	}
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&target, "target", "t", "127.0.0.1", "cisco device address to be connected")
	rootCmd.PersistentFlags().IntVarP(&port, "port", "P", 22, "device port")
	rootCmd.PersistentFlags().StringVarP(&username, "username", "u", "admin", "")
	rootCmd.PersistentFlags().StringVarP(&password, "password", "p", "admin", "")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

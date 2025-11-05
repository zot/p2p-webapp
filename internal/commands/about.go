package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

// AboutCmd represents the about command
var AboutCmd = &cobra.Command{
	Use:   "about",
	Short: "Display information about p2p-webapp",
	Long:  `Display project URL and license information for p2p-webapp.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("p2p-webapp - A Go application to host peer-to-peer applications")
		fmt.Println()
		fmt.Println("Copyright (C) 2025, Bill Burdick")
		fmt.Println("MIT Licensed")
		fmt.Println("Project URL: https://github.com/zot/p2p-webapp")
	},
}

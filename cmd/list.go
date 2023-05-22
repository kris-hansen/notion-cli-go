// This code is licensed under the Apache License, Version 2.0 (the "License").
// You may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0

package cmd

import (
	"fmt"
	"notioncli/utils"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tasks",
	Long:  `List all tasks in the Notion page`,
	Run: func(cmd *cobra.Command, args []string) {
		notionAPIKey, pageID := utils.SetAPIConfig()
		localTimezone, err := utils.GetLocalTimeZone()
		brightWhite := color.New(color.FgHiWhite).SprintFunc()
		if err != nil {
			fmt.Println("Error getting the local time zone: ", err)
			os.Exit(1)
		}
		blocks, err := utils.GetToDoBlocks(notionAPIKey, pageID, localTimezone)
		if err != nil {
			fmt.Println("Error getting blocks from the pageID: ", err)

			os.Exit(1)
		}
		for _, block := range blocks {
			fmt.Println(brightWhite(block))
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}

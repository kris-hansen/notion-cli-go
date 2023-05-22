// This code is licensed under the Apache License, Version 2.0 (the "License").
// You may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0

package cmd

import (
	"fmt"
	"notioncli/utils"
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <item order>",
	Short: "Remove a task from the task list",
	Long:  `Completely delete a task, e.g., delete 2 (removes the second task)`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		order, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Printf("Could not convert %q to an integer: %v", args[0], err)
			os.Exit(1)
		}
		notionAPIKey, pageID := utils.SetAPIConfig()
		result := utils.DeleteToDoBlock(notionAPIKey, pageID, order)
		if result != nil {
			fmt.Printf("Error removing task %d : %v\n", order, result)
			os.Exit(1)
		}
		fmt.Printf("Task %d removed.\n", order)

	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.Flags().Int("order", 0, "numeric order of the task to delete")
}

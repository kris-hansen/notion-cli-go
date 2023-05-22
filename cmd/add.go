package cmd

import (
	"fmt"
	"notioncli/utils"
	"os"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <task>",
	Short: "Add a new task",
	Long:  `Add a new task to the Notion ToDo task list page`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		text := args[0]
		notionAPIKey, pageID := utils.SetAPIConfig()
		result := utils.AddNewToDoItem(notionAPIKey, pageID, text)
		if result != nil {
			fmt.Printf("Error adding new task: %s\n", result)
			os.Exit(1)
		}
		fmt.Printf("Task %s added.\n", text)

	},
}

func init() {
	rootCmd.AddCommand(addCmd)
	checkCmd.Flags().String("text", "", "Text for the new task")
}

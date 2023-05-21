// This code is licensed under the Apache License, Version 2.0 (the "License").
// You may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0

package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "notioncli",
	Short: "notioncli provides a CLI interface to track your tasks in a Notion page",
	Long: `notioncli works with the official Notion API to extend a Notion page with to-dos into your command line environment.
	
		This version supports the following options:
		--list (to list tasks)
		--create <task> (create a new task)
		--done <number> (mark a task done)
		--undone <number> (mark a task as not done)
		--delete <number> (permanently remove a task)
		--help (get some help)`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(listCmd)
}

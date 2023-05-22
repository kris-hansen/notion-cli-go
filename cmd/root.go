// This code is licensed under the Apache License, Version 2.0 (the "License").
// You may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0

package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "notioncli",
	Short: "Notioncli provides a CLI interface to track your tasks in a Notion page",
	Long: `Notioncli is a tool that utilizes the official Notion API to enable the integration of to-do lists from Notion pages into your command line interface.
	
		This version supports the following options:
		  list (to list tasks)
		  add <task> (create a new task)
		  check <number> (mark a task done)
		  uncheck <number> (mark a task as not done)
		  delete <number> (permanently remove a task)
		  help (get some help)`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	boldBlue := color.New(color.Bold, color.FgBlue).SprintFunc()
	fmt.Println(boldBlue("----=[ NotionCLI ]=----"))
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {

}

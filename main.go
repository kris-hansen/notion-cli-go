/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"fmt"
	"notioncli/cmd"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()

	if err != nil {
		fmt.Println("Error loading .env file")
		os.Exit(1)
	}
	notionAPIKey, ok := os.LookupEnv("NOTION_API_KEY")
	if !ok {
		fmt.Println("NOTION_API_KEY environment variable not found")
		os.Exit(1)
	}
	pageID, ok := os.LookupEnv("NOTION_PAGE_ID")
	if !ok {
		fmt.Println("NOTION_PAGE_ID environment variable not found")
		os.Exit(1)
	}
	blocks, err := cmd.GetBlocks(notionAPIKey, pageID)
	if err != nil {
		fmt.Println("Error getting blocks from the pageID")
		os.Exit(1)
	}
	fmt.Println(blocks)
	cmd.Execute()
}

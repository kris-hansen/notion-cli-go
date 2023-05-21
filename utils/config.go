package utils

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

func SetAPIConfig() (string, string) {
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
	return notionAPIKey, pageID
}

func GetLocalTimeZone() (*time.Location, error) {
	err := godotenv.Load()

	if err != nil {
		fmt.Println("Error loading .env file")
		os.Exit(1)
	}
	localTimeZone, ok := os.LookupEnv("LOCAL_TIMEZONE")
	if !ok {
		return nil, fmt.Errorf("LOCAL_TIMEZONE environment variable not found")
	}
	location, err := time.LoadLocation(localTimeZone)
	if err != nil {
		return nil, err
	}
	return location, nil
}

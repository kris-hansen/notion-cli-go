package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
)

func SetAPIConfig() (string, string) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error getting home directory: ", err)
		os.Exit(1)
	}

	envPathHomeDir := filepath.Join(homeDir, ".config/notioncli/.env")
	workingDir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current directory: ", err)
		os.Exit(1)
	}

	envPathWorkingDir := filepath.Join(workingDir, ".env")
	err = godotenv.Load(envPathWorkingDir)

	if err != nil {
		// If the env file is not found in the working directory, try to load it from the home directory
		err = godotenv.Load(envPathHomeDir)
		if err != nil {
			fmt.Println("Error loading .env file: ", err)
			os.Exit(1)
		}
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
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error getting home directory: ", err)
		os.Exit(1)
	}

	envPathHomeDir := filepath.Join(homeDir, ".config/notioncli/.env")
	workingDir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current directory: ", err)
		os.Exit(1)
	}

	envPathWorkingDir := filepath.Join(workingDir, ".env")
	err = godotenv.Load(envPathWorkingDir)

	if err != nil {
		// If the env file is not found in the working directory, try to load it from the home directory
		err = godotenv.Load(envPathHomeDir)
		if err != nil {
			fmt.Println("Error loading .env file: ", err)
			os.Exit(1)
		}
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

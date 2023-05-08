package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const baseURL = "https://api.notion.com/v1/"

var blocks []Block

type Page struct {
	Object         string                 `json:"object"`
	ID             string                 `json:"id"`
	CreatedTime    string                 `json:"created_time"`
	LastEditedTime string                 `json:"last_edited_time"`
	Title          string                 `json:"title"`
	Properties     map[string]interface{} `json:"properties"`
}

// Define a struct for a block
type Block struct {
	Object         string `json:"object"`
	ID             string `json:"id"`
	CreatedTime    string `json:"created_time"`
	LastEditedTime string `json:"last_edited_time"`
	Type           string `json:"type"`
	HasChildren    bool   `json:"has_children"`
	Paragraph      *struct {
		Text []struct {
			Type string `json:"type"`
			Text struct {
				Content string `json:"content"`
			} `json:"text"`
		} `json:"text"`
	} `json:"paragraph,omitempty"`
	ToDo *struct {
		Text []struct {
			Type string `json:"type"`
			Text struct {
				Content string `json:"content"`
			} `json:"text"`
		} `json:"text"`
		Checked bool `json:"checked"`
	} `json:"to_do,omitempty"`
}

func getNotionPage(notionAPIKey, pageID string) (*Page, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", baseURL+"pages/"+pageID, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("Notion-Version", "2022-06-28")
	req.Header.Set("Authorization", "Bearer "+notionAPIKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var page Page
	err = json.Unmarshal(body, &page)
	if err != nil {
		return nil, err
	}

	return &page, nil
}

func GetBlocks(notionAPIKey, blockID string) ([]Block, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", baseURL+"blocks/"+blockID+"/children", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("Notion-Version", "2022-06-28")
	req.Header.Set("Authorization", "Bearer "+notionAPIKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}

	results := data["results"].([]interface{})
	var blocks []Block
	for _, result := range results {
		var block Block
		blockJSON, err := json.Marshal(result)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(blockJSON, &block)
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, block)
	}

	return blocks, nil
}

func getToDoBlocks(blocks []byte) ([]Block, error) {
	var toDoBlocks []Block
	var data map[string]interface{}
	err := json.Unmarshal(blocks, &data)
	if err != nil {
		return nil, err
	}
	results := data["results"].([]interface{})
	for _, result := range results {
		var block Block
		blockJSON, err := json.Marshal(result)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(blockJSON, &block)
		if err != nil {
			return nil, err
		}
		if block.Type == "to_do" {
			toDoBlocks = append(toDoBlocks, block)
		}
	}
	return toDoBlocks, nil
}

func main() {

	fmt.Printf("Blocks: %s\n", blocks)

	var toDoBlocks []Block
	for _, block := range blocks {
		if block.Type == "to_do" {
			toDoBlocks = append(toDoBlocks, block)
		}
	}

	fmt.Printf("To-Do Blocks: %s\n", toDoBlocks)

}

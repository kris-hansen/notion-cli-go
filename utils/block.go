// This code is licensed under the Apache License, Version 2.0 (the "License").
// You may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0

package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

var baseURL = "https://api.notion.com/v1"

var blocks []Block

type ToDo struct {
	Checked  bool       `json:"checked"`
	Color    string     `json:"color"`
	RichText []RichText `json:"rich_text"`
}

type RichText struct {
	Annotations Annotation  `json:"annotations"`
	Href        interface{} `json:"href"`
	PlainText   string      `json:"plain_text"`
	Text        Text        `json:"text"`
	Type        string      `json:"type"`
}

type Annotation struct {
	Bold          bool   `json:"bold"`
	Code          bool   `json:"code"`
	Color         string `json:"color"`
	Italic        bool   `json:"italic"`
	Strikethrough bool   `json:"strikethrough"`
	Underline     bool   `json:"underline"`
}

type Text struct {
	Content string      `json:"content"`
	Link    interface{} `json:"link"`
}

type Block struct {
	Object         string `json:"object"`
	ID             string `json:"id"`
	CreatedTime    string `json:"created_time"`
	LastEditedTime string `json:"last_edited_time"`
	Type           string `json:"type"`
	HasChildren    bool   `json:"has_children"`
	ToDo           *ToDo  `json:"to_do,omitempty"`
}

type BlockList struct {
	Object          string   `json:"object"`
	Results         []Block  `json:"results"`
	NextCursor      string   `json:"next_cursor"`
	HasMore         bool     `json:"has_more"`
	Type            string   `json:"type"`
	Block           struct{} `json:"block"`
	DeveloperSurvey string   `json:"developer_survey"`
}

func GetBlocks(notionAPIKey, pageID string) ([]Block, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", baseURL+"/blocks/"+pageID+"/children", nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("Notion-Version", "2022-06-28")
	req.Header.Set("Authorization", "Bearer "+notionAPIKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var blockList BlockList

	err = json.NewDecoder(resp.Body).Decode(&blockList)
	if err != nil {
		return nil, err
	}

	var blocks []Block
	for _, result := range blockList.Results {
		if result.Object == "block" && result.ToDo != nil && len(result.ToDo.RichText) > 0 {
			blocks = append(blocks, result)
		}
	}
	return blocks, nil
}

func GetToDoBlocks(notionAPIKey, blockID string, localTimezone *time.Location) ([]string, error) {
	blocks, err := GetBlocks(notionAPIKey, blockID)
	if err != nil {
		return nil, err
	}
	var todoBlocks []string
	for _, block := range blocks {
		if block.ToDo != nil {
			var checked string
			if block.ToDo.Checked {
				checked = "X"
			} else {
				checked = " "
			}
			lastEditedTime, err := time.Parse(time.RFC3339, block.LastEditedTime)
			if err != nil {
				return nil, err
			}
			truncatedTime := lastEditedTime.In(localTimezone).Truncate(time.Minute)

			element := fmt.Sprintf("%d [%s] %s (%s)", len(todoBlocks)+1, checked, block.ToDo.RichText[0].PlainText, truncatedTime.Format("2006-01-02 15:04"))
			todoBlocks = append(todoBlocks, element)
		}
	}

	return todoBlocks, nil
}

func AddNewToDoItem(notionAPIKey, pageID, text string) error {
	client := &http.Client{}
	reqBody, err := json.Marshal(map[string]interface{}{
		"children": []map[string]interface{}{
			{
				"object": "block",
				"type":   "to_do",
				"to_do": map[string]interface{}{
					"rich_text": []map[string]interface{}{
						{
							"type": "text",
							"text": map[string]interface{}{
								"content": text,
							},
						},
					},
				},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("error marshalling request body: %v", err)
	}

	req, err := http.NewRequest("PATCH", baseURL+"/blocks/"+pageID+"/children", bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	req.Header.Add("Notion-Version", "2022-06-28")
	req.Header.Set("Authorization", "Bearer "+notionAPIKey)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d, message: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

func GetBlockID(notionAPIKey, pageID string, order int) (string, error) {
	if order < 1 {
		return "", fmt.Errorf("order must be greater than 0")
	}
	client := &http.Client{}
	req, err := http.NewRequest("GET", baseURL+"/blocks/"+pageID+"/children", nil)
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("Notion-Version", "2022-06-28")
	req.Header.Set("Authorization", "Bearer "+notionAPIKey)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var blockList BlockList
	err = json.NewDecoder(resp.Body).Decode(&blockList)
	if err != nil {
		return "", err
	}

	if order > len(blockList.Results) {
		return "", fmt.Errorf("order number exceeds the number of blocks")
	}

	return blockList.Results[order-1].ID, nil

}

func MarkToDoBlockChecked(notionAPIKey, pageID string, order int) error {

	blockID, err := GetBlockID(notionAPIKey, pageID, order)
	if err != nil {
		return err
	}
	client := &http.Client{}
	reqBody, err := json.Marshal(map[string]interface{}{
		"to_do": map[string]interface{}{
			"checked": true,
		},
	})
	if err != nil {
		return fmt.Errorf("error marshalling request body: %v", err)
	}

	req, err := http.NewRequest("PATCH", baseURL+"/blocks/"+blockID, bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	req.Header.Add("Notion-Version", "2022-06-28")
	req.Header.Set("Authorization", "Bearer "+notionAPIKey)

	resp, err := client.Do(req)

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func MarkToDoBlockUnChecked(notionAPIKey, pageID string, order int) error {

	blockID, err := GetBlockID(notionAPIKey, pageID, order)
	if err != nil {
		return err
	}
	client := &http.Client{}
	reqBody, err := json.Marshal(map[string]interface{}{
		"to_do": map[string]interface{}{
			"checked": false,
		},
	})
	if err != nil {
		return fmt.Errorf("error marshalling request body: %v", err)
	}

	req, err := http.NewRequest("PATCH", baseURL+"/blocks/"+blockID, bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	req.Header.Add("Notion-Version", "2022-06-28")
	req.Header.Set("Authorization", "Bearer "+notionAPIKey)

	resp, err := client.Do(req)

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func DeleteToDoBlock(notionAPIKey, pageID string, order int) error {
	blockID, err := GetBlockID(notionAPIKey, pageID, order)
	if err != nil {
		return err
	}
	client := &http.Client{}
	req, err := http.NewRequest("DELETE", baseURL+"/blocks/"+blockID, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Add("Notion-Version", "2022-06-28")
	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Authorization", "Bearer "+notionAPIKey)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

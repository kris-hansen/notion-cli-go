// This code is licensed under the Apache License, Version 2.0 (the "License").
// You may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0

package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const baseURL = "https://api.notion.com/v1/"

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

func GetBlocks(notionAPIKey, blockID string) ([]Block, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", baseURL+"blocks/"+blockID+"/children", nil)
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
		if result.Object == "block" && result.ToDo != nil {
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
			todoBlocks = append(todoBlocks, fmt.Sprintf("%d [%s] %s (%s)", len(todoBlocks)+1, checked, block.ToDo.RichText[0].PlainText, truncatedTime.Format("2006-01-02 15:04")))
		}
	}

	return todoBlocks, nil
}

func AddToDoBlock(notionAPIKey, pageID, content string) error {
	client := &http.Client{}
	reqBody := map[string]interface{}{
		"parent": map[string]interface{}{
			"page_id": pageID,
		},
		"to_do": map[string]interface{}{
			"title": []map[string]interface{}{
				{
					"text": map[string]interface{}{
						"content": content,
					},
				},
			},
		},
	}
	reqBodyJSON, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", baseURL+"blocks", bytes.NewBuffer(reqBodyJSON))
	if err != nil {
		return err
	}
	req.Header.Add("accept", "application/json")
	req.Header.Add("Notion-Version", "2022-06-28")
	req.Header.Set("Authorization", "Bearer "+notionAPIKey)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

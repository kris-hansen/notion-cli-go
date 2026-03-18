// This code is licensed under the Apache License, Version 2.0 (the "License").
// You may not use this file except in compliance with the License.
// You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0

package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"time"
)

var baseURL = "https://api.notion.com/v1"

var blocks []Block

// BlockTypeInfo contains display information for a block type
type BlockTypeInfo struct {
	Icon  string
	Color string
}

// SupportedBlockTypes defines all supported block types with their display info
var SupportedBlockTypes = map[string]BlockTypeInfo{
	"paragraph":          {Icon: "¶", Color: "white"},
	"heading_1":          {Icon: "H1", Color: "cyan"},
	"heading_2":          {Icon: "H2", Color: "cyan"},
	"heading_3":          {Icon: "H3", Color: "cyan"},
	"bulleted_list_item": {Icon: "•", Color: "white"},
	"numbered_list_item": {Icon: "#", Color: "white"},
	"to_do":              {Icon: "☐", Color: "green"},
	"toggle":             {Icon: "▸", Color: "magenta"},
	"quote":              {Icon: "❝", Color: "yellow"},
	"callout":            {Icon: "💡", Color: "yellow"},
	"divider":            {Icon: "—", Color: "white"},
	"code":               {Icon: "<>", Color: "blue"},
}

// GetSupportedBlockTypeNames returns a sorted list of supported block type names
func GetSupportedBlockTypeNames() []string {
	names := make([]string, 0, len(SupportedBlockTypes))
	for name := range SupportedBlockTypes {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// IsValidBlockType checks if a block type is supported
func IsValidBlockType(blockType string) bool {
	_, ok := SupportedBlockTypes[blockType]
	return ok
}

type ToDo struct {
	Checked  bool       `json:"checked"`
	Color    string     `json:"color"`
	RichText []RichText `json:"rich_text"`
}

// RichTextBlock represents a block with rich_text content
type RichTextBlock struct {
	RichText []RichText `json:"rich_text"`
	Color    string     `json:"color,omitempty"`
	Language string     `json:"language,omitempty"` // for code blocks
	Checked  bool       `json:"checked,omitempty"`  // for to_do blocks
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
	Object           string         `json:"object"`
	ID               string         `json:"id"`
	CreatedTime      string         `json:"created_time"`
	LastEditedTime   string         `json:"last_edited_time"`
	Type             string         `json:"type"`
	HasChildren      bool           `json:"has_children"`
	ToDo             *ToDo          `json:"to_do,omitempty"`
	Paragraph        *RichTextBlock `json:"paragraph,omitempty"`
	Heading1         *RichTextBlock `json:"heading_1,omitempty"`
	Heading2         *RichTextBlock `json:"heading_2,omitempty"`
	Heading3         *RichTextBlock `json:"heading_3,omitempty"`
	BulletedListItem *RichTextBlock `json:"bulleted_list_item,omitempty"`
	NumberedListItem *RichTextBlock `json:"numbered_list_item,omitempty"`
	Toggle           *RichTextBlock `json:"toggle,omitempty"`
	Quote            *RichTextBlock `json:"quote,omitempty"`
	Callout          *RichTextBlock `json:"callout,omitempty"`
	Code             *RichTextBlock `json:"code,omitempty"`
	Divider          *struct{}      `json:"divider,omitempty"`
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
		bodyBytes, _ := io.ReadAll(resp.Body)
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

// GetAllBlocks retrieves all blocks under a page, optionally filtered by type
// Handles pagination for pages with more than 100 blocks
func GetAllBlocks(notionAPIKey, pageID string, filterType string) ([]Block, error) {
	client := &http.Client{}
	var result []Block
	var cursor string

	for {
		url := baseURL + "/blocks/" + pageID + "/children"
		if cursor != "" {
			url += "?start_cursor=" + cursor
		}

		req, err := http.NewRequest("GET", url, nil)
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

		var blockList BlockList
		err = json.NewDecoder(resp.Body).Decode(&blockList)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}

		for _, block := range blockList.Results {
			if block.Object == "block" {
				// If no filter or matches filter, include block
				if filterType == "" || block.Type == filterType {
					result = append(result, block)
				}
			}
		}

		// Check if there are more pages
		if !blockList.HasMore || blockList.NextCursor == "" {
			break
		}
		cursor = blockList.NextCursor
	}

	return result, nil
}

// GetBlockContent extracts text content from any block type
func GetBlockContent(block Block) string {
	switch block.Type {
	case "divider":
		return "───────────"
	case "to_do":
		if block.ToDo != nil && len(block.ToDo.RichText) > 0 {
			return block.ToDo.RichText[0].PlainText
		}
	case "paragraph":
		if block.Paragraph != nil && len(block.Paragraph.RichText) > 0 {
			return block.Paragraph.RichText[0].PlainText
		}
	case "heading_1":
		if block.Heading1 != nil && len(block.Heading1.RichText) > 0 {
			return block.Heading1.RichText[0].PlainText
		}
	case "heading_2":
		if block.Heading2 != nil && len(block.Heading2.RichText) > 0 {
			return block.Heading2.RichText[0].PlainText
		}
	case "heading_3":
		if block.Heading3 != nil && len(block.Heading3.RichText) > 0 {
			return block.Heading3.RichText[0].PlainText
		}
	case "bulleted_list_item":
		if block.BulletedListItem != nil && len(block.BulletedListItem.RichText) > 0 {
			return block.BulletedListItem.RichText[0].PlainText
		}
	case "numbered_list_item":
		if block.NumberedListItem != nil && len(block.NumberedListItem.RichText) > 0 {
			return block.NumberedListItem.RichText[0].PlainText
		}
	case "toggle":
		if block.Toggle != nil && len(block.Toggle.RichText) > 0 {
			return block.Toggle.RichText[0].PlainText
		}
	case "quote":
		if block.Quote != nil && len(block.Quote.RichText) > 0 {
			return block.Quote.RichText[0].PlainText
		}
	case "callout":
		if block.Callout != nil && len(block.Callout.RichText) > 0 {
			return block.Callout.RichText[0].PlainText
		}
	case "code":
		if block.Code != nil && len(block.Code.RichText) > 0 {
			return block.Code.RichText[0].PlainText
		}
	}
	return "(empty)"
}

// GetBlockIcon returns the display icon for a block
func GetBlockIcon(block Block) string {
	// Special case for to_do: show checked/unchecked
	if block.Type == "to_do" && block.ToDo != nil {
		if block.ToDo.Checked {
			return "☑"
		}
		return "☐"
	}

	if info, ok := SupportedBlockTypes[block.Type]; ok {
		return info.Icon
	}
	return "?"
}

// FormatAllBlocks returns formatted strings for all blocks
func FormatAllBlocks(notionAPIKey, pageID string, localTimezone *time.Location, filterType string) ([]string, map[string]int, error) {
	blocks, err := GetAllBlocks(notionAPIKey, pageID, filterType)
	if err != nil {
		return nil, nil, err
	}

	var formatted []string
	typeCounts := make(map[string]int)

	for index, block := range blocks {
		lastEditedTime, err := time.Parse(time.RFC3339, block.LastEditedTime)
		if err != nil {
			return nil, nil, err
		}
		truncatedTime := lastEditedTime.In(localTimezone).Truncate(time.Minute)

		icon := GetBlockIcon(block)
		content := GetBlockContent(block)

		// Truncate long content
		if len(content) > 50 {
			content = content[:47] + "..."
		}

		element := fmt.Sprintf("%4d %s  [%-20s] %s (%s)",
			index+1,
			icon,
			block.Type,
			content,
			truncatedTime.Format("2006-01-02 15:04"))
		formatted = append(formatted, element)
		typeCounts[block.Type]++
	}

	return formatted, typeCounts, nil
}

// AddBlock adds a new block of any supported type
func AddBlock(notionAPIKey, pageID, blockType, text string) error {
	if !IsValidBlockType(blockType) {
		return fmt.Errorf("unsupported block type: %s", blockType)
	}

	client := &http.Client{}

	var blockContent map[string]interface{}

	if blockType == "divider" {
		blockContent = map[string]interface{}{
			"object":  "block",
			"type":    "divider",
			"divider": map[string]interface{}{},
		}
	} else {
		// Most blocks use rich_text
		richText := []map[string]interface{}{
			{
				"type": "text",
				"text": map[string]interface{}{
					"content": text,
				},
			},
		}

		innerContent := map[string]interface{}{
			"rich_text": richText,
		}

		// Add language for code blocks
		if blockType == "code" {
			innerContent["language"] = "plain text"
		}

		blockContent = map[string]interface{}{
			"object":  "block",
			"type":    blockType,
			blockType: innerContent,
		}
	}

	reqBody, err := json.Marshal(map[string]interface{}{
		"children": []map[string]interface{}{blockContent},
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
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d, message: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// DeleteBlock deletes any block by its index (1-based)
func DeleteBlock(notionAPIKey, pageID string, order int) error {
	// Get all blocks to find the one at the given index
	blocks, err := GetAllBlocks(notionAPIKey, pageID, "")
	if err != nil {
		return err
	}

	if order < 1 || order > len(blocks) {
		return fmt.Errorf("block number %d out of range (1-%d)", order, len(blocks))
	}

	blockID := blocks[order-1].ID

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
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d, message: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

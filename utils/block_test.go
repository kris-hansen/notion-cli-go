package utils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Mock server
var (
	mockServer    *httptest.Server
	deletedBlocks map[string]bool
)

func SetBaseURL(url string) {
	baseURL = url
}

func mockBlock(texts []string) Block {

	richTexts := make([]RichText, len(texts))
	for i, text := range texts {
		richTexts[i] = RichText{PlainText: text}
	}

	return Block{
		Object: "block",
		ID:     "blockID",
		Type:   "to_do",
		ToDo: &ToDo{
			Checked:  false,
			RichText: richTexts,
		},
	}
}

func setup() {
	deletedBlocks = make(map[string]bool)
	mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("{}")) // Send back an empty JSON body for DELETE requests
		case r.URL.Path == "/blocks/pageID/children":

			result := BlockList{
				Results: []Block{
					mockBlock([]string{"test todo"}),
				},
			}
			err := json.NewEncoder(w).Encode(result)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

		case r.URL.Path == "/blocks/pageWithToDoWithNoContent/children":

			result := BlockList{
				Results: []Block{
					mockBlock([]string{}),
					mockBlock([]string{"test todo"}),
				},
			}
			err := json.NewEncoder(w).Encode(result)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		case r.URL.Path == "/blocks/blockID":

			result := Block{
				Object: "block",
				ID:     "blockID",
				Type:   "to_do",
				ToDo: &ToDo{
					Checked:  true,
					RichText: []RichText{{PlainText: "test todo"}},
				},
			}
			err := json.NewEncoder(w).Encode(result)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}
	}))
	SetBaseURL(mockServer.URL)
}

func teardown() {
	mockServer.Close()
}

func TestGetBlocks(t *testing.T) {
	setup()
	defer teardown()

	notionAPIKey := "fakeKey"
	pageID := "pageID"
	blocks, err := GetBlocks(notionAPIKey, pageID)

	if err != nil {
		t.Errorf("Got error: %v", err)
	}

	if len(blocks) != 1 {
		t.Errorf("Expected 1 block, got: %v", len(blocks))
	}
}

func TestGetBlocksIfRichTextIsEmpty(t *testing.T) {
	setup()
	defer teardown()
	notionAPIKey := "fakeKey"
	pageID := "pageWithToDoWithNoContent"
	blocks, err := GetBlocks(notionAPIKey, pageID)

	if err != nil {
		t.Errorf("Got error: %v", err)
	}

	if len(blocks) != 1 {
		t.Errorf("Expected 1 block, got: %v", len(blocks))
	}

}

func TestAddNewToDoItem(t *testing.T) {
	setup()
	defer teardown()

	notionAPIKey := "fakeKey"
	pageID := "pageID"
	toDoText := "new todo"
	err := AddNewToDoItem(notionAPIKey, pageID, toDoText)

	if err != nil {
		t.Errorf("Got error: %v", err)
	}
}

func TestDeleteToDoBlock(t *testing.T) {
	setup()
	defer teardown()

	notionAPIKey := "fakeKey"
	pageID := "pageID"
	order := 1

	err := DeleteToDoBlock(notionAPIKey, pageID, order)

	if err != nil {
		t.Errorf("Error deleting to-do block with ID %s and order %d: %v", pageID, order, err)
	}

}

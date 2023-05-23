package utils

import (
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

func setup() {
	deletedBlocks = make(map[string]bool)
	mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("{}")) // Send back an empty JSON body for DELETE requests
		case r.URL.Path == "/blocks/pageID/children":
			response := `{"object": "list", "results": [{"object": "block", "id": "blockID", "type": "to_do", "to_do": {"checked": false, "rich_text": [{"plain_text": "test todo"}]}}]}`
			w.Write([]byte(response))
		case r.URL.Path == "/blocks/blockID":
			response := `{"object": "block", "id": "blockID", "type": "to_do", "to_do": {"checked": true, "rich_text": [{"plain_text": "test todo"}]}}`
			w.Write([]byte(response))
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

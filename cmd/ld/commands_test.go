package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rodstewart/linkding-cli/internal/models"
)

// executeCommand executes a command with the given arguments and returns the output
func executeCommand(t *testing.T, args ...string) (string, error) {
	t.Helper()

	// Create pipes to capture stdout and stderr
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	rOut, wOut, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create stdout pipe: %v", err)
	}
	rErr, wErr, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create stderr pipe: %v", err)
	}
	os.Stdout = wOut
	os.Stderr = wErr

	// Create channels to capture the output
	outC := make(chan string)
	errC := make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, rOut)
		outC <- buf.String()
	}()
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, rErr)
		errC <- buf.String()
	}()

	// Set command arguments
	rootCmd.SetArgs(args)

	// Execute the command
	cmdErr := rootCmd.Execute()

	// Restore stdout/stderr and close the writers
	wOut.Close()
	wErr.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	// Get the captured output
	stdout := <-outC
	stderr := <-errC

	// Combine stdout and stderr
	output := stdout + stderr

	// Reset args and global flags for next test (but keep commands registered)
	rootCmd.SetArgs(nil)
	jsonOutput = false
	debugMode = false

	return output, cmdErr
}

// setupMockServer creates a mock LinkDing API server for testing
func setupMockServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(func() { server.Close() })
	return server
}

// setTestEnv sets up environment variables for testing
func setTestEnv(t *testing.T, url, token string) {
	t.Helper()
	os.Setenv("LINKDING_URL", url)
	os.Setenv("LINKDING_TOKEN", token)
	t.Cleanup(func() {
		os.Unsetenv("LINKDING_URL")
		os.Unsetenv("LINKDING_TOKEN")
	})
}

// mockBookmark creates a sample bookmark for testing
func mockBookmark(id int, url, title string, tags []string) models.Bookmark {
	return models.Bookmark{
		ID:           id,
		URL:          url,
		Title:        title,
		Description:  fmt.Sprintf("Description for %s", title),
		TagNames:     tags,
		DateAdded:    time.Now(),
		DateModified: time.Now(),
		Unread:       false,
		Shared:       false,
		IsArchived:   false,
	}
}

// TestAddCommand tests the 'ld add' command
func TestAddCommand(t *testing.T) {
	// Create a mock server
	server := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/bookmarks/" && r.Method == "POST" {
			// Parse request body
			var create models.BookmarkCreate
			if err := json.NewDecoder(r.Body).Decode(&create); err != nil {
				t.Fatalf("Failed to decode request: %v", err)
			}

			// Return created bookmark
			bookmark := mockBookmark(1, create.URL, create.Title, create.TagNames)
			if create.Title == "" {
				bookmark.Title = "Auto-fetched Title"
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated) // 201 for CREATE operations
			json.NewEncoder(w).Encode(bookmark)
			return
		}
		http.NotFound(w, r)
	})

	setTestEnv(t, server.URL, "test-token")

	t.Run("add bookmark with minimal args", func(t *testing.T) {
		output, err := executeCommand(t, "add", "https://example.com")
		if err != nil {
			t.Fatalf("Command failed: %v", err)
		}

		if !strings.Contains(output, "✓ Bookmark added:") {
			t.Errorf("Expected success message, got: %s", output)
		}
		if !strings.Contains(output, "ID: 1") {
			t.Errorf("Expected bookmark ID in output, got: %s", output)
		}
	})

	t.Run("add bookmark with title and tags", func(t *testing.T) {
		output, err := executeCommand(t, "add", "https://example.com/test",
			"--title", "Test Bookmark",
			"--tags", "test,example")
		if err != nil {
			t.Fatalf("Command failed: %v", err)
		}

		if !strings.Contains(output, "✓ Bookmark added:") {
			t.Errorf("Expected success message, got: %s", output)
		}
		if !strings.Contains(output, "Tags:") {
			t.Errorf("Expected tags in output, got: %s", output)
		}
	})

	t.Run("add bookmark with json output", func(t *testing.T) {
		output, err := executeCommand(t, "add", "https://example.com/json",
			"--json")
		if err != nil {
			t.Fatalf("Command failed: %v", err)
		}

		var bookmark models.Bookmark
		if err := json.Unmarshal([]byte(output), &bookmark); err != nil {
			t.Errorf("Expected valid JSON output, got error: %v, output: %s", err, output)
		}
		if bookmark.ID != 1 {
			t.Errorf("Expected bookmark ID 1, got: %d", bookmark.ID)
		}
	})
}

// TestListCommand tests the 'ld list' command
func TestListCommand(t *testing.T) {
	server := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/bookmarks/" && r.Method == "GET" {
			bookmarks := []models.Bookmark{
				mockBookmark(1, "https://example.com", "Example Site", []string{"example"}),
				mockBookmark(2, "https://test.com", "Test Site", []string{"test"}),
			}

			response := models.BookmarkList{
				Count:    2,
				Next:     nil,
				Previous: nil,
				Results:  bookmarks,
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}
		http.NotFound(w, r)
	})

	setTestEnv(t, server.URL, "test-token")

	t.Run("list bookmarks table format", func(t *testing.T) {
		output, err := executeCommand(t, "list")
		if err != nil {
			t.Fatalf("Command failed: %v", err)
		}

		// Check for table headers and content
		if !strings.Contains(output, "ID") || !strings.Contains(output, "TITLE") {
			t.Errorf("Expected table headers in output, got: %s", output)
		}
		if !strings.Contains(output, "Example Site") {
			t.Errorf("Expected 'Example Site' in output, got: %s", output)
		}
		if !strings.Contains(output, "Test Site") {
			t.Errorf("Expected 'Test Site' in output, got: %s", output)
		}
	})

	t.Run("list bookmarks json format", func(t *testing.T) {
		output, err := executeCommand(t, "list", "--json")
		if err != nil {
			t.Fatalf("Command failed: %v", err)
		}

		var bookmarkList models.BookmarkList
		if err := json.Unmarshal([]byte(output), &bookmarkList); err != nil {
			t.Errorf("Expected valid JSON, got error: %v, output: %s", err, output)
		}
		if len(bookmarkList.Results) != 2 {
			t.Errorf("Expected 2 bookmarks, got: %d", len(bookmarkList.Results))
		}
		if bookmarkList.Count != 2 {
			t.Errorf("Expected count of 2, got: %d", bookmarkList.Count)
		}
	})
}

// TestGetCommand tests the 'ld get' command
func TestGetCommand(t *testing.T) {
	server := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/bookmarks/") && r.Method == "GET" {
			bookmark := mockBookmark(1, "https://example.com", "Example Site", []string{"example", "test"})
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(bookmark)
			return
		}
		http.NotFound(w, r)
	})

	setTestEnv(t, server.URL, "test-token")

	t.Run("get bookmark by id", func(t *testing.T) {
		output, err := executeCommand(t, "get", "1")
		if err != nil {
			t.Fatalf("Command failed: %v", err)
		}

		if !strings.Contains(output, "Example Site") {
			t.Errorf("Expected bookmark title in output, got: %s", output)
		}
		if !strings.Contains(output, "https://example.com") {
			t.Errorf("Expected bookmark URL in output, got: %s", output)
		}
		if !strings.Contains(output, "Tags:") {
			t.Errorf("Expected tags section in output, got: %s", output)
		}
	})

	t.Run("get bookmark json output", func(t *testing.T) {
		output, err := executeCommand(t, "get", "1", "--json")
		if err != nil {
			t.Fatalf("Command failed: %v", err)
		}

		var bookmark models.Bookmark
		if err := json.Unmarshal([]byte(output), &bookmark); err != nil {
			t.Errorf("Expected valid JSON, got error: %v, output: %s", err, output)
		}
		if bookmark.ID != 1 {
			t.Errorf("Expected bookmark ID 1, got: %d", bookmark.ID)
		}
	})
}

// TestUpdateCommand tests the 'ld update' command
func TestUpdateCommand(t *testing.T) {
	server := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/bookmarks/") {
			if r.Method == "GET" {
				// Return existing bookmark for merge operations
				bookmark := mockBookmark(1, "https://example.com", "Example Site", []string{"existing"})
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(bookmark)
				return
			} else if r.Method == "PATCH" {
				// Parse update request
				var update models.BookmarkUpdate
				if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
					t.Fatalf("Failed to decode request: %v", err)
				}

				// Return updated bookmark
				bookmark := mockBookmark(1, "https://example.com", "Example Site", []string{"existing", "new"})
				if update.TagNames != nil {
					bookmark.TagNames = *update.TagNames
				}

				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(bookmark)
				return
			}
		}
		http.NotFound(w, r)
	})

	setTestEnv(t, server.URL, "test-token")

	t.Run("update with add-tags", func(t *testing.T) {
		output, err := executeCommand(t, "update", "1", "--add-tags", "new")
		if err != nil {
			t.Fatalf("Command failed: %v", err)
		}

		if !strings.Contains(output, "✓ Bookmark updated") {
			t.Errorf("Expected success message, got: %s", output)
		}
	})
}

// TestDeleteCommand tests the 'ld delete' command
func TestDeleteCommand(t *testing.T) {
	server := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/bookmarks/") && r.Method == "DELETE" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		http.NotFound(w, r)
	})

	setTestEnv(t, server.URL, "test-token")

	t.Run("delete with force flag", func(t *testing.T) {
		output, err := executeCommand(t, "delete", "1", "--force")
		if err != nil {
			t.Fatalf("Command failed: %v", err)
		}

		if !strings.Contains(output, "✓ Bookmark") && !strings.Contains(output, "deleted") {
			t.Errorf("Expected success message, got: %s", output)
		}
	})
}

// TestExportCommand tests the 'ld export' command
func TestExportCommand(t *testing.T) {
	server := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/bookmarks/" && r.Method == "GET" {
			bookmarks := []models.Bookmark{
				mockBookmark(1, "https://example.com", "Example", []string{"test"}),
			}

			response := models.BookmarkList{
				Count:    1,
				Next:     nil,
				Previous: nil,
				Results:  bookmarks,
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}
		http.NotFound(w, r)
	})

	setTestEnv(t, server.URL, "test-token")

	t.Run("export to csv", func(t *testing.T) {
		output, err := executeCommand(t, "export", "-f", "csv")
		if err != nil {
			t.Fatalf("Command failed: %v", err)
		}

		// Check for CSV headers
		if !strings.Contains(output, "url,title,description,tags") {
			t.Errorf("Expected CSV headers in output, got: %s", output)
		}
		if !strings.Contains(output, "https://example.com") {
			t.Errorf("Expected bookmark data in output, got: %s", output)
		}
	})
}

// TestBackupCommand tests the 'ld backup' command
func TestBackupCommand(t *testing.T) {
	server := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/bookmarks/" && r.Method == "GET" {
			bookmarks := []models.Bookmark{
				mockBookmark(1, "https://example.com", "Example", []string{"test"}),
			}

			response := models.BookmarkList{
				Count:    1,
				Next:     nil,
				Previous: nil,
				Results:  bookmarks,
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}
		http.NotFound(w, r)
	})

	setTestEnv(t, server.URL, "test-token")

	t.Run("backup creates file", func(t *testing.T) {
		tmpDir := t.TempDir()
		output, err := executeCommand(t, "backup", "--output", tmpDir)
		if err != nil {
			t.Fatalf("Command failed: %v", err)
		}

		if !strings.Contains(output, "Backup created:") {
			t.Errorf("Expected success message, got: %s", output)
		}

		// Verify file was created
		files, err := os.ReadDir(tmpDir)
		if err != nil {
			t.Fatalf("Failed to read temp dir: %v", err)
		}
		if len(files) != 1 {
			t.Errorf("Expected 1 backup file, got: %d", len(files))
		}

		// Verify filename contains timestamp pattern
		filename := files[0].Name()
		if !strings.HasPrefix(filename, "linkding-backup-") || !strings.HasSuffix(filename, ".json") {
			t.Errorf("Unexpected backup filename: %s", filename)
		}
	})

	t.Run("backup json output", func(t *testing.T) {
		tmpDir := t.TempDir()
		output, err := executeCommand(t, "backup", "--output", tmpDir, "--json")
		if err != nil {
			t.Fatalf("Command failed: %v", err)
		}

		var result map[string]string
		if err := json.Unmarshal([]byte(output), &result); err != nil {
			t.Errorf("Expected valid JSON, got error: %v, output: %s", err, output)
		}
		if result["file"] == "" {
			t.Errorf("Expected file path in JSON output")
		}

		// Verify the file path exists
		if _, err := os.Stat(result["file"]); os.IsNotExist(err) {
			t.Errorf("Backup file does not exist: %s", result["file"])
		}
	})
}

// TestTagsCommand tests the 'ld tags' command
func TestTagsCommand(t *testing.T) {
	server := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/tags/" && r.Method == "GET" {
			// Return tags list
			tags := []models.Tag{
				{ID: 1, Name: "golang", DateAdded: time.Now()},
				{ID: 2, Name: "cli", DateAdded: time.Now()},
				{ID: 3, Name: "testing", DateAdded: time.Now()},
			}

			response := models.TagList{
				Count:    3,
				Next:     nil,
				Previous: nil,
				Results:  tags,
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}

		if r.URL.Path == "/api/bookmarks/" && r.Method == "GET" {
			// Return bookmarks with various tags
			bookmarks := []models.Bookmark{
				mockBookmark(1, "https://example.com", "Example 1", []string{"golang", "cli"}),
				mockBookmark(2, "https://example.com/2", "Example 2", []string{"golang", "testing"}),
				mockBookmark(3, "https://example.com/3", "Example 3", []string{"cli"}),
			}

			response := models.BookmarkList{
				Count:    3,
				Next:     nil,
				Previous: nil,
				Results:  bookmarks,
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}
		http.NotFound(w, r)
	})

	setTestEnv(t, server.URL, "test-token")

	t.Run("list tags with counts", func(t *testing.T) {
		output, err := executeCommand(t, "tags")
		if err != nil {
			t.Fatalf("Command failed: %v", err)
		}

		// Check for table headers
		if !strings.Contains(output, "TAG") || !strings.Contains(output, "COUNT") {
			t.Errorf("Expected table headers in output, got: %s", output)
		}

		// Check for tags and counts
		if !strings.Contains(output, "golang") {
			t.Errorf("Expected 'golang' tag in output, got: %s", output)
		}
		if !strings.Contains(output, "cli") {
			t.Errorf("Expected 'cli' tag in output, got: %s", output)
		}
	})

	t.Run("list tags json output", func(t *testing.T) {
		output, err := executeCommand(t, "tags", "--json")
		if err != nil {
			t.Fatalf("Command failed: %v", err)
		}

		var tags []models.TagWithCount
		if err := json.Unmarshal([]byte(output), &tags); err != nil {
			t.Errorf("Expected valid JSON array, got error: %v, output: %s", err, output)
		}
		if len(tags) == 0 {
			t.Errorf("Expected tags in output, got none")
		}
	})
}

// TestConfigCommand tests the 'ld config show' command
func TestConfigCommand(t *testing.T) {
	t.Run("config show redacts token", func(t *testing.T) {
		// Set up environment with token
		setTestEnv(t, "https://linkding.example.com", "supersecrettoken123")

		output, err := executeCommand(t, "config", "show")
		if err != nil {
			t.Fatalf("Command failed: %v", err)
		}

		// Verify URL is shown
		if !strings.Contains(output, "https://linkding.example.com") {
			t.Errorf("Expected URL in output, got: %s", output)
		}

		// Verify token is redacted
		if strings.Contains(output, "supersecrettoken123") {
			t.Errorf("Token should be redacted in output, got: %s", output)
		}
		if !strings.Contains(output, "...") && !strings.Contains(output, "***") && !strings.Contains(output, "[REDACTED]") {
			t.Errorf("Expected token redaction indicator, got: %s", output)
		}
	})
}

// TestMissingConfig tests commands when config is not set
func TestMissingConfig(t *testing.T) {
	// Ensure no config environment variables are set
	os.Unsetenv("LINKDING_URL")
	os.Unsetenv("LINKDING_TOKEN")

	// Also ensure no config file is found by using a non-existent path
	tmpDir := t.TempDir()
	nonExistentConfig := filepath.Join(tmpDir, "nonexistent", "config.yaml")

	t.Run("command fails with missing config", func(t *testing.T) {
		_, err := executeCommand(t, "--config", nonExistentConfig, "list")
		if err == nil {
			t.Error("Expected error with missing config, got nil")
		}
	})
}

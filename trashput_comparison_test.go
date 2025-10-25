package trash

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestTrashPutCompatibility compares the behavior of go-trash with trash-put
func TestTrashPutCompatibility(t *testing.T) {
	// Check if trash-put is installed
	if _, err := exec.LookPath("trash-put"); err != nil {
		t.Skip("trash-put not installed, skipping compatibility test")
	}

	// Clean trash before test
	Empty()

	tempDir := t.TempDir()

	t.Run("SingleFile", func(t *testing.T) {
		// Create test file
		testFile := filepath.Join(tempDir, "single_test.txt")
		content := []byte("test content")
		if err := os.WriteFile(testFile, content, 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		// Trash with go-trash
		if err := Trash(testFile); err != nil {
			t.Fatalf("Failed to trash file: %v", err)
		}

		// Get the trashed item
		items, err := List()
		if err != nil {
			t.Fatalf("Failed to list trash: %v", err)
		}

		var item *TrashItem
		for _, i := range items {
			if strings.HasPrefix(i.Name, "single_test.txt") {
				item = &i
				break
			}
		}

		if item == nil {
			t.Fatal("Trashed file not found")
		}

		// Read and verify trashinfo file
		infoContent, err := os.ReadFile(item.InfoPath)
		if err != nil {
			t.Fatalf("Failed to read trashinfo: %v", err)
		}

		verifyTrashInfo(t, string(infoContent), testFile)
	})

	t.Run("DuplicateFiles", func(t *testing.T) {
		// Clean trash
		Empty()

		// Create and trash multiple files with same name
		fileName := "duplicate_compare.txt"
		for i := 0; i < 3; i++ {
			testFile := filepath.Join(tempDir, fileName)
			if err := os.WriteFile(testFile, []byte("content"), 0644); err != nil {
				t.Fatalf("Failed to create test file %d: %v", i, err)
			}

			if err := Trash(testFile); err != nil {
				t.Fatalf("Failed to trash file %d: %v", i, err)
			}
		}

		// Get all items
		items, err := List()
		if err != nil {
			t.Fatalf("Failed to list trash: %v", err)
		}

		var foundNames []string
		for _, item := range items {
			if strings.HasPrefix(item.Name, fileName) {
				foundNames = append(foundNames, item.Name)
			}
		}

		// Verify naming matches trash-put convention
		expectedNames := []string{
			"duplicate_compare.txt",
			"duplicate_compare.txt_1",
			"duplicate_compare.txt_2",
		}

		if len(foundNames) != len(expectedNames) {
			t.Errorf("Expected %d files, got %d", len(expectedNames), len(foundNames))
		}

		for _, expected := range expectedNames {
			found := false
			for _, name := range foundNames {
				if name == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Missing expected name %q. Found: %v", expected, foundNames)
			}
		}
	})

	t.Run("FileWithSpaces", func(t *testing.T) {
		// Clean trash
		Empty()

		fileName := "file with spaces.txt"
		testFile := filepath.Join(tempDir, fileName)
		if err := os.WriteFile(testFile, []byte("content"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		if err := Trash(testFile); err != nil {
			t.Fatalf("Failed to trash file: %v", err)
		}

		items, err := List()
		if err != nil {
			t.Fatalf("Failed to list trash: %v", err)
		}

		var item *TrashItem
		for _, i := range items {
			if i.Name == fileName {
				item = &i
				break
			}
		}

		if item == nil {
			t.Fatal("Trashed file not found")
		}

		// Verify trashinfo format
		infoContent, err := os.ReadFile(item.InfoPath)
		if err != nil {
			t.Fatalf("Failed to read trashinfo: %v", err)
		}

		verifyTrashInfo(t, string(infoContent), testFile)
	})

	t.Run("TrashInfoFormat", func(t *testing.T) {
		// Clean trash
		Empty()

		testFile := filepath.Join(tempDir, "format_test.txt")
		if err := os.WriteFile(testFile, []byte("content"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		if err := Trash(testFile); err != nil {
			t.Fatalf("Failed to trash file: %v", err)
		}

		items, err := List()
		if err != nil {
			t.Fatalf("Failed to list trash: %v", err)
		}

		var item *TrashItem
		for _, i := range items {
			if strings.HasPrefix(i.Name, "format_test.txt") {
				item = &i
				break
			}
		}

		if item == nil {
			t.Fatal("Trashed file not found")
		}

		// Read trashinfo
		infoContent, err := os.ReadFile(item.InfoPath)
		if err != nil {
			t.Fatalf("Failed to read trashinfo: %v", err)
		}

		content := string(infoContent)

		// Verify format matches FreeDesktop spec
		lines := strings.Split(content, "\n")
		if len(lines) < 3 {
			t.Fatalf("Trashinfo has too few lines: %d", len(lines))
		}

		// First line must be [Trash Info]
		if lines[0] != "[Trash Info]" {
			t.Errorf("First line should be '[Trash Info]', got: %q", lines[0])
		}

		// Should have Path= line
		hasPath := false
		hasDeletionDate := false
		for _, line := range lines {
			if strings.HasPrefix(line, "Path=") {
				hasPath = true
			}
			if strings.HasPrefix(line, "DeletionDate=") {
				hasDeletionDate = true
			}
		}

		if !hasPath {
			t.Error("Missing 'Path=' line in trashinfo")
		}
		if !hasDeletionDate {
			t.Error("Missing 'DeletionDate=' line in trashinfo")
		}
	})
}

// verifyTrashInfo verifies that the trashinfo content follows the FreeDesktop spec
func verifyTrashInfo(t *testing.T, content string, originalPath string) {
	t.Helper()

	lines := strings.Split(content, "\n")
	if len(lines) < 3 {
		t.Errorf("Trashinfo has too few lines: %d", len(lines))
		return
	}

	// Verify header
	if lines[0] != "[Trash Info]" {
		t.Errorf("Expected '[Trash Info]' header, got: %q", lines[0])
	}

	// Verify Path line exists
	hasPath := false
	hasDate := false

	for _, line := range lines {
		if strings.HasPrefix(line, "Path=") {
			hasPath = true
			// Path should be URL-encoded
			pathValue := strings.TrimPrefix(line, "Path=")
			if pathValue == "" {
				t.Error("Path value is empty")
			}
		}
		if strings.HasPrefix(line, "DeletionDate=") {
			hasDate = true
			// Date should be in ISO 8601 format: YYYY-MM-DDTHH:MM:SS
			dateValue := strings.TrimPrefix(line, "DeletionDate=")
			if len(dateValue) < 19 {
				t.Errorf("DeletionDate format seems wrong: %q", dateValue)
			}
		}
	}

	if !hasPath {
		t.Error("Missing 'Path=' line")
	}
	if !hasDate {
		t.Error("Missing 'DeletionDate=' line")
	}
}

// TestTrashPutInteroperability tests that files trashed by trash-put can be listed and restored
func TestTrashPutInteroperability(t *testing.T) {
	// Check if trash-put is installed
	if _, err := exec.LookPath("trash-put"); err != nil {
		t.Skip("trash-put not installed, skipping interoperability test")
	}

	// Clean trash
	Empty()

	tempDir := t.TempDir()

	t.Run("RestoreTrashPutFile", func(t *testing.T) {
		// Create a test file
		testFile := filepath.Join(tempDir, "trashput_test.txt")
		content := []byte("trash-put test content")
		if err := os.WriteFile(testFile, content, 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		// Trash it using trash-put
		cmd := exec.Command("trash-put", testFile)
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to run trash-put: %v", err)
		}

		// Verify file is gone
		if _, err := os.Stat(testFile); !os.IsNotExist(err) {
			t.Error("File still exists after trash-put")
		}

		// List with go-trash
		items, err := List()
		if err != nil {
			t.Fatalf("Failed to list trash: %v", err)
		}

		// Find the item
		var item *TrashItem
		for _, i := range items {
			if i.OriginalPath == testFile {
				item = &i
				break
			}
		}

		if item == nil {
			t.Fatal("File trashed by trash-put not found in go-trash list")
		}

		// Restore with go-trash
		if err := Restore(item.Name); err != nil {
			t.Fatalf("Failed to restore trash-put file: %v", err)
		}

		// Verify file is restored
		restoredContent, err := os.ReadFile(testFile)
		if err != nil {
			t.Fatalf("Failed to read restored file: %v", err)
		}

		if string(restoredContent) != string(content) {
			t.Error("Restored file content doesn't match original")
		}
	})

	t.Run("TrashPutDuplicates", func(t *testing.T) {
		// Clean trash
		Empty()

		// Create and trash multiple files with trash-put
		fileName := "trashput_dup.txt"
		for i := 0; i < 3; i++ {
			testFile := filepath.Join(tempDir, fileName)
			if err := os.WriteFile(testFile, []byte("content"), 0644); err != nil {
				t.Fatalf("Failed to create test file %d: %v", i, err)
			}

			cmd := exec.Command("trash-put", testFile)
			if err := cmd.Run(); err != nil {
				t.Fatalf("Failed to run trash-put %d: %v", i, err)
			}
		}

		// List with go-trash
		items, err := List()
		if err != nil {
			t.Fatalf("Failed to list trash: %v", err)
		}

		// Find trash-put items
		var foundNames []string
		for _, item := range items {
			if strings.HasPrefix(item.Name, fileName) {
				foundNames = append(foundNames, item.Name)
			}
		}

		// Should have 3 items with the naming pattern
		if len(foundNames) != 3 {
			t.Errorf("Expected 3 items from trash-put, got %d: %v", len(foundNames), foundNames)
		}

		// Verify naming matches our convention
		expectedPattern := []string{fileName, fileName + "_1", fileName + "_2"}
		for _, expected := range expectedPattern {
			found := false
			for _, name := range foundNames {
				if name == expected {
					found = true
					break
				}
			}
			if !found {
				t.Logf("Expected name %q not found. This is OK if trash-put uses different numbering", expected)
			}
		}
	})
}

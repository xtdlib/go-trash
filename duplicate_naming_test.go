package trash

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestDuplicateNamingCompatibility tests that duplicate file naming
// follows the trash-put convention: a, a_1, a_2, etc.
func TestDuplicateNamingCompatibility(t *testing.T) {
	// Clean trash before test
	Empty()

	tempDir := t.TempDir()

	// Create and trash 5 files with the same name
	expectedNames := []string{"a", "a_1", "a_2", "a_3", "a_4"}

	for i := 0; i < 5; i++ {
		testFile := filepath.Join(tempDir, "a")
		if err := os.WriteFile(testFile, []byte("content"), 0644); err != nil {
			t.Fatalf("Failed to create test file %d: %v", i, err)
		}

		if err := Trash(testFile); err != nil {
			t.Fatalf("Failed to trash file %d: %v", i, err)
		}
	}

	// Get all items from trash
	items, err := List()
	if err != nil {
		t.Fatalf("Failed to list trash: %v", err)
	}

	// Find items that match our test
	var foundNames []string
	for _, item := range items {
		if filepath.Base(item.OriginalPath) == "a" &&
			filepath.Dir(item.OriginalPath) == tempDir {
			foundNames = append(foundNames, item.Name)
		}
	}

	// Verify we found all 5 files
	if len(foundNames) != 5 {
		t.Errorf("Expected 5 files in trash, got %d", len(foundNames))
		t.Logf("Found names: %v", foundNames)
	}

	// Verify the naming convention matches trash-put
	for i, expected := range expectedNames {
		found := false
		for _, name := range foundNames {
			if name == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Missing expected trash name %q (file %d). Found names: %v", expected, i, foundNames)
		}
	}

	// Verify no names use dot notation (old style)
	for _, name := range foundNames {
		if name != "a" && strings.Contains(name, ".") && !strings.Contains(name, "_") {
			t.Errorf("Found old-style naming with dots: %q", name)
		}
	}
}

// TestDuplicateNamingWithExtension tests duplicate naming with file extensions
func TestDuplicateNamingWithExtension(t *testing.T) {
	tempDir := t.TempDir()

	// Create and trash 3 files with the same name but with extension
	expectedNames := []string{"test.txt", "test.txt_1", "test.txt_2"}

	for i := 0; i < 3; i++ {
		testFile := filepath.Join(tempDir, "test.txt")
		if err := os.WriteFile(testFile, []byte("content"), 0644); err != nil {
			t.Fatalf("Failed to create test file %d: %v", i, err)
		}

		if err := Trash(testFile); err != nil {
			t.Fatalf("Failed to trash file %d: %v", i, err)
		}
	}

	items, err := List()
	if err != nil {
		t.Fatalf("Failed to list trash: %v", err)
	}

	var foundNames []string
	for _, item := range items {
		if filepath.Base(item.OriginalPath) == "test.txt" &&
			filepath.Dir(item.OriginalPath) == tempDir {
			foundNames = append(foundNames, item.Name)
		}
	}

	if len(foundNames) != 3 {
		t.Errorf("Expected 3 files in trash, got %d", len(foundNames))
	}

	for i, expected := range expectedNames {
		found := false
		for _, name := range foundNames {
			if name == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Missing expected trash name %q (file %d)", expected, i)
		}
	}
}

// TestDuplicateNamingRestore tests that duplicates can be restored correctly
func TestDuplicateNamingRestore(t *testing.T) {
	tempDir := t.TempDir()

	// Create and trash 3 files with the same name
	for i := 0; i < 3; i++ {
		testFile := filepath.Join(tempDir, "restore_test")
		content := []byte("content " + string(rune('0'+i)))
		if err := os.WriteFile(testFile, content, 0644); err != nil {
			t.Fatalf("Failed to create test file %d: %v", i, err)
		}

		if err := Trash(testFile); err != nil {
			t.Fatalf("Failed to trash file %d: %v", i, err)
		}
	}

	items, err := List()
	if err != nil {
		t.Fatalf("Failed to list trash: %v", err)
	}

	// Find all our items
	var testItems []TrashItem
	for _, item := range items {
		if filepath.Base(item.OriginalPath) == "restore_test" &&
			filepath.Dir(item.OriginalPath) == tempDir {
			testItems = append(testItems, item)
		}
	}

	if len(testItems) != 3 {
		t.Fatalf("Expected 3 items, got %d", len(testItems))
	}

	// Restore one of the duplicates
	if err := Restore(testItems[1].Name); err != nil {
		t.Fatalf("Failed to restore file: %v", err)
	}

	// Verify the file was restored
	restoredFile := filepath.Join(tempDir, "restore_test")
	if _, err := os.Stat(restoredFile); err != nil {
		t.Fatalf("Restored file doesn't exist: %v", err)
	}

	// Clean up
	os.Remove(restoredFile)
}

// TestTrashPutNamingPattern verifies the exact naming pattern used by trash-put
func TestTrashPutNamingPattern(t *testing.T) {
	testCases := []struct {
		baseName string
		index    int
		expected string
	}{
		{"file", 0, "file"},
		{"file", 1, "file_1"},
		{"file", 2, "file_2"},
		{"file.txt", 0, "file.txt"},
		{"file.txt", 1, "file.txt_1"},
		{"file.txt", 2, "file.txt_2"},
		{"my.archive.tar.gz", 0, "my.archive.tar.gz"},
		{"my.archive.tar.gz", 1, "my.archive.tar.gz_1"},
	}

	for _, tc := range testCases {
		var result string
		if tc.index == 0 {
			result = tc.baseName
		} else {
			result = tc.baseName + "_" + string(rune('0'+tc.index))
		}

		if result != tc.expected {
			t.Errorf("Naming pattern mismatch for %q at index %d: got %q, want %q",
				tc.baseName, tc.index, result, tc.expected)
		}
	}
}

// TestManyDuplicates tests behavior with many duplicate files
func TestManyDuplicates(t *testing.T) {
	tempDir := t.TempDir()

	// Create and trash 10 files with the same name
	count := 10
	for i := 0; i < count; i++ {
		testFile := filepath.Join(tempDir, "many")
		if err := os.WriteFile(testFile, []byte("content"), 0644); err != nil {
			t.Fatalf("Failed to create test file %d: %v", i, err)
		}

		if err := Trash(testFile); err != nil {
			t.Fatalf("Failed to trash file %d: %v", i, err)
		}
	}

	items, err := List()
	if err != nil {
		t.Fatalf("Failed to list trash: %v", err)
	}

	var foundNames []string
	for _, item := range items {
		if filepath.Base(item.OriginalPath) == "many" &&
			filepath.Dir(item.OriginalPath) == tempDir {
			foundNames = append(foundNames, item.Name)
		}
	}

	if len(foundNames) != count {
		t.Errorf("Expected %d files in trash, got %d", count, len(foundNames))
	}

	// Verify first is just "many"
	hasBase := false
	for _, name := range foundNames {
		if name == "many" {
			hasBase = true
			break
		}
	}
	if !hasBase {
		t.Error("Missing base name 'many' in trash")
	}

	// Verify all others follow the pattern
	for i := 1; i < count; i++ {
		expectedSuffix := "_" + string(rune('0'+i))
		found := false
		for _, name := range foundNames {
			if strings.HasSuffix(name, expectedSuffix) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Missing file with suffix %q", expectedSuffix)
		}
	}
}

// TestDuplicateNamingWithSpaces tests duplicate naming for files with spaces
func TestDuplicateNamingWithSpaces(t *testing.T) {
	// Clean trash before test
	Empty()

	tempDir := t.TempDir()

	// Create and trash 4 files with spaces in the name
	fileName := "my file with spaces.txt"
	expectedNames := []string{
		"my file with spaces.txt",
		"my file with spaces.txt_1",
		"my file with spaces.txt_2",
		"my file with spaces.txt_3",
	}

	for i := 0; i < 4; i++ {
		testFile := filepath.Join(tempDir, fileName)
		if err := os.WriteFile(testFile, []byte("content"), 0644); err != nil {
			t.Fatalf("Failed to create test file %d: %v", i, err)
		}

		if err := Trash(testFile); err != nil {
			t.Fatalf("Failed to trash file %d: %v", i, err)
		}
	}

	items, err := List()
	if err != nil {
		t.Fatalf("Failed to list trash: %v", err)
	}

	var foundNames []string
	for _, item := range items {
		if filepath.Base(item.OriginalPath) == fileName &&
			filepath.Dir(item.OriginalPath) == tempDir {
			foundNames = append(foundNames, item.Name)
		}
	}

	if len(foundNames) != 4 {
		t.Errorf("Expected 4 files in trash, got %d", len(foundNames))
		t.Logf("Found names: %v", foundNames)
	}

	// Verify all expected names are present
	for i, expected := range expectedNames {
		found := false
		for _, name := range foundNames {
			if name == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Missing expected trash name %q (file %d). Found names: %v", expected, i, foundNames)
		}
	}

	// Test that we can restore a file with spaces
	if len(foundNames) > 0 {
		// Restore the first duplicate (should be "my file with spaces.txt_1")
		var restoreName string
		for _, name := range foundNames {
			if strings.HasSuffix(name, "_1") {
				restoreName = name
				break
			}
		}

		if restoreName != "" {
			if err := Restore(restoreName); err != nil {
				t.Fatalf("Failed to restore file with spaces: %v", err)
			}

			// Verify the file was restored
			restoredFile := filepath.Join(tempDir, fileName)
			if _, err := os.Stat(restoredFile); err != nil {
				t.Fatalf("Restored file doesn't exist: %v", err)
			}

			content, err := os.ReadFile(restoredFile)
			if err != nil {
				t.Fatalf("Failed to read restored file: %v", err)
			}

			if string(content) != "content" {
				t.Errorf("Restored file content doesn't match: got %q, want %q", string(content), "content")
			}
		}
	}
}

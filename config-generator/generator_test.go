package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

func setupTestEnv(t *testing.T) (string, string, func()) {
	t.Helper()
	baseTmpDir, err := os.MkdirTemp("", "sogo-test-")
	if err != nil {
		t.Fatalf("Failed to create base temp dir: %v", err)
	}
	tmpConfigFolder := filepath.Join(baseTmpDir, "sogo.conf.d")
	tmpConfigPath := filepath.Join(baseTmpDir, "sogo.conf")
	if err := os.MkdirAll(tmpConfigFolder, 0755); err != nil {
		os.RemoveAll(baseTmpDir) // Clean up if subdir creation fails
		t.Fatalf("Failed to create temp config folder %s: %v", tmpConfigFolder, err)
	}
	cleanup := func() {
		os.RemoveAll(baseTmpDir)
	}
	return tmpConfigFolder, tmpConfigPath, cleanup
}

func createYAMLFile(t *testing.T, dir, filename, content string) {
	t.Helper()
	filePath := filepath.Join(dir, filename)
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write temp YAML file %s: %v", filePath, err)
	}
}

func TestDeepMergeMaps(t *testing.T) {
	t.Parallel() // This test doesn't have side effects

	tests := []struct {
		name string
		dst  map[string]interface{}
		src  map[string]interface{}
		want map[string]interface{}
	}{
		{
			name: "simple merge",
			dst:  map[string]interface{}{"a": 1},
			src:  map[string]interface{}{"b": 2},
			want: map[string]interface{}{"a": 1, "b": 2},
		},
		{
			name: "overwrite value",
			dst:  map[string]interface{}{"a": 1},
			src:  map[string]interface{}{"a": 2},
			want: map[string]interface{}{"a": 2},
		},
		{
			name: "nested merge",
			dst:  map[string]interface{}{"a": map[string]interface{}{"x": 1}},
			src:  map[string]interface{}{"a": map[string]interface{}{"y": 2}},
			want: map[string]interface{}{"a": map[string]interface{}{"x": 1, "y": 2}},
		},
		{
			name: "nested overwrite",
			dst:  map[string]interface{}{"a": map[string]interface{}{"x": 1}},
			src:  map[string]interface{}{"a": map[string]interface{}{"x": 2}},
			want: map[string]interface{}{"a": map[string]interface{}{"x": 2}},
		},
		{
			name: "add nested",
			dst:  map[string]interface{}{"a": 1},
			src:  map[string]interface{}{"b": map[string]interface{}{"y": 2}},
			want: map[string]interface{}{"a": 1, "b": map[string]interface{}{"y": 2}},
		},
		{
			name: "merge into nil dst",
			dst:  nil,
			src:  map[string]interface{}{"a": 1},
			want: map[string]interface{}{"a": 1},
		},
		{
			name: "merge nil src",
			dst:  map[string]interface{}{"a": 1},
			src:  nil,
			want: map[string]interface{}{"a": 1},
		},
		{
			name: "merge empty src",
			dst:  map[string]interface{}{"a": 1},
			src:  map[string]interface{}{},
			want: map[string]interface{}{"a": 1},
		},
		{
			name: "merge into empty dst",
			dst:  map[string]interface{}{},
			src:  map[string]interface{}{"a": 1},
			want: map[string]interface{}{"a": 1},
		},
		{
			name: "type overwrite map->scalar",
			dst:  map[string]interface{}{"a": map[string]interface{}{"x": 1}},
			src:  map[string]interface{}{"a": "hello"},
			want: map[string]interface{}{"a": "hello"},
		},
		{
			name: "type overwrite scalar->map",
			dst:  map[string]interface{}{"a": "hello"},
			src:  map[string]interface{}{"a": map[string]interface{}{"x": 1}},
			want: map[string]interface{}{"a": map[string]interface{}{"x": 1}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := deepMergeMaps(tt.dst, tt.src)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("deepMergeMaps() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func runGenerateTest(t *testing.T, setupFunc func(t *testing.T, tmpConfigFolder string)) (string, error) {
	t.Helper()
	tmpConfFolder, tmpConfPath, cleanup := setupTestEnv(t)
	// Use t.Cleanup for reliable cleanup even on panic
	t.Cleanup(cleanup)

	// Run the setup function to create specific test files
	if setupFunc != nil {
		setupFunc(t, tmpConfFolder)
	}

	// Execute the function under test, passing the temporary paths
	err := generateConfigFile(tmpConfFolder, tmpConfPath)

	// Read the output file content
	var outputContent string
	contentBytes, readErr := os.ReadFile(tmpConfPath)
	if readErr != nil {
		// If the file doesn't exist, it might be expected in some error cases,
		// but usually generateConfigFile should create it (even if empty/error).
		// Let the test case decide if this is an error. Return the readErr.
		if !os.IsNotExist(readErr) {
			t.Logf("Warning: Failed to read output config file %s: %v", tmpConfPath, readErr)
		}
		// Return the original error from generateConfigFile if it occurred
		if err != nil {
			return "", err
		}
		// If generateConfigFile had no error, but reading failed, return readErr
		return "", fmt.Errorf("generateConfigFile succeeded but reading output failed: %w", readErr)

	}
	outputContent = string(contentBytes)

	return outputContent, err // Return content and the error from generateConfigFile
}

func containsPlistKeyValuePair(t *testing.T, output, key, expectedValue string) bool {
	t.Helper()
	// Allow the key to be quoted or not.
	keyPattern := fmt.Sprintf(`"?%s"?`, regexp.QuoteMeta(key))
	var valuePattern string
	if expectedValue == "YES" {
		valuePattern = `(?:YES|1|true|True|TRUE)`
	} else if expectedValue == "NO" {
		valuePattern = `(?:NO|0|false|False|FALSE)`
	} else {
		if _, err := strconv.ParseFloat(expectedValue, 64); err == nil {
			valuePattern = expectedValue
		} else {
			valuePattern = fmt.Sprintf(`(?:%s|"%s")`, regexp.QuoteMeta(expectedValue), regexp.QuoteMeta(expectedValue))
		}
	}
	// Remove the begin-of-line anchor so that matching is less strict.
	fullPattern := fmt.Sprintf(`(?m)\s*%s\s*=\s*%s\s*;`, keyPattern, valuePattern)
	matched, _ := regexp.MatchString(fullPattern, output)
	if !matched {
		t.Logf("Debug: Failed to match key='%s', value='%s' with pattern: %s", key, expectedValue, fullPattern)
	}
	return matched
}

func containsPlistArrayItems(t *testing.T, output, key string, items []string) bool {
	t.Helper()
	// Allow the key to be quoted or not.
	startPattern := fmt.Sprintf(`(?m)^\s*"?%s"?\s*=\s*\(\s*`, regexp.QuoteMeta(key))
	startRe := regexp.MustCompile(startPattern)
	startIndex := startRe.FindStringIndex(output)
	if startIndex == nil {
		t.Logf("Could not find start of array for key '%s' using pattern '%s'", key, startPattern)
		return false
	}
	contentStart := startIndex[1]
	balance := 1
	contentEnd := -1
	inQuote := false
	for i := contentStart; i < len(output); i++ {
		switch output[i] {
		case '"':
			inQuote = !inQuote
		case '(':
			if !inQuote {
				balance++
			}
		case ')':
			if !inQuote {
				balance--
				if balance == 0 {
					if i+1 < len(output) && output[i+1] == ';' {
						contentEnd = i
						goto endLoop
					}
				}
			}
		}
	}
endLoop:
	if contentEnd == -1 {
		t.Logf("Could not find balanced closing parenthesis for array key '%s'", key)
		return false
	}
	arrayContent := output[contentStart:contentEnd]
	t.Logf("Found array content for key '%s': %s", key, arrayContent)
	lastIndex := 0
	allItemsFound := true
	for _, item := range items {
		var valuePat string
		if item == "YES" {
			valuePat = `(?:YES|1|true|True|TRUE)`
		} else if item == "NO" {
			valuePat = `(?:NO|0|false|False|FALSE)`
		} else if _, err := strconv.ParseFloat(item, 64); err == nil {
			valuePat = item
		} else {
			valuePat = fmt.Sprintf(`(?:%s|"%s")`, regexp.QuoteMeta(item), regexp.QuoteMeta(item))
		}
		// Look for the item at the beginning of a line (optionally preceded by a comma)
		itemPattern := fmt.Sprintf(`(?m)^\s*(?:,?\s*)%s\s*(?:,|$)`, valuePat)
		itemRe := regexp.MustCompile(itemPattern)
		loc := itemRe.FindStringIndex(arrayContent[lastIndex:])
		if loc == nil {
			t.Logf("Item '%s' not found in remaining array content: '%s'", item, arrayContent[lastIndex:])
			allItemsFound = false
			break
		}
		lastIndex += loc[1]
	}
	return allItemsFound
}

func TestGenerateConfigFile_Basic(t *testing.T) {
	t.Parallel()
	setup := func(t *testing.T, tmpConfigFolder string) {
		createYAMLFile(t, tmpConfigFolder, "01-basic.yaml", `
SOGoUserSources:
  - type: ldap
    id: users # Keep this unquoted in YAML
    hostname: ldap://ldap.example.com
    baseDN: ou=users,dc=example,dc=com
    port: 636
    is_bool: true
`)
	}

	output, err := runGenerateTest(t, setup)
	if err != nil {
		t.Fatalf("generateConfigFile() failed: %v", err)
	}

	if !strings.Contains(output, disclaimerMessage) {
		t.Errorf("Output missing disclaimer message")
	}

	// Check for key elements within the plist structure, less strict on exact format/order
	if !strings.Contains(output, `SOGoUserSources = (`) {
		t.Errorf("Missing 'SOGoUserSources = ('")
	}
	if !strings.Contains(output, `{`) { // Check for dictionary start
		t.Errorf("Missing '{' for dictionary inside array")
	}
	// Check key-value pairs within the dictionary - allow unquoted simple strings
	if !containsPlistKeyValuePair(t, output, "type", "ldap") {
		t.Errorf("Missing or incorrect 'type = ldap;'")
	}
	if !containsPlistKeyValuePair(t, output, "id", "users") { // howett.net/plist might not quote simple strings
		t.Errorf("Missing or incorrect 'id = users;'")
	}
	if !containsPlistKeyValuePair(t, output, "hostname", "ldap://ldap.example.com") {
		t.Errorf("Missing or incorrect 'hostname = \"ldap://ldap.example.com\";'")
	}
	if !containsPlistKeyValuePair(t, output, "baseDN", "ou=users,dc=example,dc=com") {
		t.Errorf("Missing or incorrect 'baseDN = \"ou=users,dc=example,dc=com\";'")
	}
	if !containsPlistKeyValuePair(t, output, "port", "636") {
		t.Errorf("Missing or incorrect 'port = 636;'")
	}
	if !containsPlistKeyValuePair(t, output, "is_bool", "YES") {
		t.Errorf("Missing or incorrect 'is_bool = YES;'")
	}
	if !strings.Contains(output, `);`) { // Check for array end
		t.Errorf("Missing ');' for array end")
	}
}

func TestGenerateConfigFile_Merge(t *testing.T) {
	t.Parallel()
	setup := func(t *testing.T, tmpConfigFolder string) {
		createYAMLFile(t, tmpConfigFolder, "01-db.yaml", `
SOGoProfileURL: "mysql://sogo:sogo@db/sogo/sogo_user_profile"
OCSEMailAlarmsFolderURL: "mysql://sogo:sogo@db/sogo/sogo_alarms_folder"
`)
		createYAMLFile(t, tmpConfigFolder, "02-ldap.yaml", `
SOGoUserSources:
  - type: ldap
    id: users
    hostname: ldap://ldap.example.com
    baseDN: ou=users,dc=example,dc=com
OCSEMailAlarmsFolderURL: "override" # This should overwrite
`)
	}

	output, err := runGenerateTest(t, setup)
	if err != nil {
		t.Fatalf("generateConfigFile() failed: %v", err)
	}

	// Check merged and overwritten values using the helper
	if !containsPlistKeyValuePair(t, output, "SOGoProfileURL", "mysql://sogo:sogo@db/sogo/sogo_user_profile") {
		t.Errorf("Missing or incorrect SOGoProfileURL from first file")
	}
	if !containsPlistKeyValuePair(t, output, "OCSEMailAlarmsFolderURL", "override") {
		t.Errorf("OCSEMailAlarmsFolderURL was not overwritten correctly")
	}
	// Check presence of a key from the second file's structure
	if !strings.Contains(output, `SOGoUserSources = (`) {
		t.Errorf("Missing SOGoUserSources structure from second file")
	}
	if !containsPlistKeyValuePair(t, output, "hostname", "ldap://ldap.example.com") {
		t.Errorf("Missing hostname within SOGoUserSources from second file")
	}
}

func TestGenerateConfigFile_EmptyDir(t *testing.T) {
	// Capture log output
	var logBuf bytes.Buffer
	logBuf.Reset()
	log.SetOutput(&logBuf)
	t.Cleanup(func() { log.SetOutput(os.Stderr) }) // Restore default logger

	output, err := runGenerateTest(t, nil) // Pass nil setupFunc
	if err != nil {
		// It shouldn't fail, just produce the empty message
		t.Fatalf("generateConfigFile() failed unexpectedly for empty dir: %v", err)
	}

	// Expect specific content for empty config, including disclaimer
	expectedStart := disclaimerMessage + "\n\n"
	expectedEnd := "{\n  /* No configuration loaded from YAML files. */\n}\n"
	expected := expectedStart + expectedEnd

	// Trim trailing whitespace from output for comparison robustness
	actualOutput := strings.TrimSpace(output) + "\n"
	expected = strings.TrimSpace(expected) + "\n"

	if actualOutput != expected {
		t.Errorf("Output for empty dir mismatch.\nGot:\n---\n%s\n---\nWant:\n---\n%s\n---", actualOutput, expected)
	}

	// Check log for the correct warning
	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "Warning: No valid YAML data found or merged result is empty.") {
		t.Errorf("Expected log warning 'No valid YAML data found...' but not found in logs:\n%s", logOutput)
	}
}

func TestGenerateConfigFile_OnlyEmptyFiles(t *testing.T) {
	setup := func(t *testing.T, tmpConfigFolder string) {
		createYAMLFile(t, tmpConfigFolder, "empty1.yaml", ``)
		createYAMLFile(t, tmpConfigFolder, "empty2.yaml", `   # Only comments`)
	}

	// Capture log output
	var logBuf bytes.Buffer
	logBuf.Reset()
	log.SetOutput(&logBuf)
	t.Cleanup(func() { log.SetOutput(os.Stderr) }) // Restore default logger

	output, err := runGenerateTest(t, setup)
	if err != nil {
		t.Fatalf("generateConfigFile() failed unexpectedly for empty files: %v", err)
	}

	// Expect specific content for empty config, including disclaimer
	expectedStart := disclaimerMessage + "\n\n"
	expectedEnd := "{\n  /* No configuration loaded from YAML files. */\n}\n"
	expected := expectedStart + expectedEnd

	// Trim trailing whitespace from output for comparison robustness
	actualOutput := strings.TrimSpace(output) + "\n"
	expected = strings.TrimSpace(expected) + "\n"

	if actualOutput != expected {
		t.Errorf("Output for only empty files mismatch.\nGot:\n---\n%s\n---\nWant:\n---\n%s\n---", actualOutput, expected)
	}

	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "Info: Skipping empty YAML file") {
		t.Errorf("Expected log message about skipping files but not found in logs:\n%s", logOutput)
	}
}

func TestGenerateConfigFile_InvalidYAML(t *testing.T) {
	setup := func(t *testing.T, tmpConfigFolder string) {
		createYAMLFile(t, tmpConfigFolder, "valid.yaml", `valid_key: true`)
		createYAMLFile(t, tmpConfigFolder, "invalid.yaml", `invalid: yaml: here`) // Invalid YAML syntax
	}

	// Capture log output
	var logBuf bytes.Buffer
	logBuf.Reset()
	log.SetOutput(&logBuf)
	t.Cleanup(func() { log.SetOutput(os.Stderr) }) // Restore default logger

	output, err := runGenerateTest(t, setup)
	if err != nil {
		// Should not fail, just skip the invalid file
		t.Fatalf("generateConfigFile() failed unexpectedly with invalid YAML: %v", err)
	}

	// Check that the valid key is present
	if !containsPlistKeyValuePair(t, output, "valid_key", "YES") {
		t.Errorf("Valid key 'valid_key = YES;' missing from output when invalid file was present")
	}

	// Check that the invalid key is NOT present
	// A simple check is usually sufficient
	if strings.Contains(output, `invalid`) {
		// More specific check to avoid matching the log message if it were in output
		if containsPlistKeyValuePair(t, output, "invalid", "yaml") {
			t.Errorf("Invalid key 'invalid' unexpectedly found in output plist structure")
		}
	}

	// Check log for warning about parsing failure
	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "Warning: Failed to parse YAML file") || !strings.Contains(logOutput, "invalid.yaml") {
		t.Errorf("Expected log warning about invalid YAML file 'invalid.yaml', but not found in logs:\n%s", logOutput)
	}
}

func TestGenerateConfigFile_Disclaimer(t *testing.T) {
	t.Parallel()
	setup := func(t *testing.T, tmpConfigFolder string) {
		createYAMLFile(t, tmpConfigFolder, "01-basic.yaml", `key: value`)
	}

	output, err := runGenerateTest(t, setup)
	if err != nil {
		t.Fatalf("generateConfigFile() failed: %v", err)
	}

	// Check that the output *starts* with the disclaimer
	if !strings.HasPrefix(output, "/* *********************") {
		t.Errorf("Output does not start with the expected disclaimer comment.")
	}

	// Check a known part of the disclaimer
	if !strings.Contains(output, "AUTOGENERATED by the Docker container") {
		t.Errorf("Output missing key part of disclaimer message")
	}

	// Check that there's a blank line after the disclaimer before the plist starts
	// The plist encoder adds its own newline after encoding, so we look for the end comment -> blank line -> {
	expectedStructure := "**************************/\n\n{"
	if !strings.Contains(output, expectedStructure) {
		t.Errorf("Could not find expected structure (disclaimer end -> blank line -> opening brace) in output:\n%s", output)
	}
}

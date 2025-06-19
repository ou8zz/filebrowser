package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/filebrowser/filebrowser/v2/files"
)

// OnlyOfficeCallback represents the callback data from OnlyOffice Document Server
type OnlyOfficeCallback struct {
	Key    string   `json:"key"`
	Status int      `json:"status"`
	URL    string   `json:"url"`
	Users  []string `json:"users"`
}

// DocumentKeyMapping stores the mapping between OnlyOffice document keys and file paths
var documentKeyMapping = make(map[string]string)

// DocumentKeyMappingRequest represents the request to store document key mapping
type DocumentKeyMappingRequest struct {
	Key  string `json:"key"`
	Path string `json:"path"`
}

// onlyOfficeMappingHandler handles storing document key to file path mappings
func onlyOfficeMappingHandler(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	// Add CORS headers for OnlyOffice integration
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	// Handle preflight OPTIONS request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return http.StatusOK, nil
	}
	// Only accept POST requests
	if r.Method != http.MethodPost {
		return http.StatusMethodNotAllowed, fmt.Errorf("method not allowed")
	}

	// Parse the mapping request
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("failed to read request body: %v", err)
	}
	defer r.Body.Close()

	var mappingReq DocumentKeyMappingRequest
	if err := json.Unmarshal(body, &mappingReq); err != nil {
		return http.StatusBadRequest, fmt.Errorf("failed to parse mapping request: %v", err)
	}

	// Store the mapping
	documentKeyMapping[mappingReq.Key] = mappingReq.Path
	fmt.Printf("Stored document key mapping: %s -> %s\n", mappingReq.Key, mappingReq.Path)

	// Return success response
	response := map[string]interface{}{
		"success": true,
		"message": "Document key mapping stored successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	return 0, json.NewEncoder(w).Encode(response)
}

// onlyOfficeCallbackHandler handles callbacks from OnlyOffice Document Server
func onlyOfficeCallbackHandler(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	// Add CORS headers for OnlyOffice integration
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	// Handle preflight OPTIONS request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return http.StatusOK, nil
	}
	// Only accept POST requests
	if r.Method != http.MethodPost {
		return http.StatusMethodNotAllowed, fmt.Errorf("method not allowed")
	}

	// Get user ID from URL parameters
	userIDStr := r.URL.Query().Get("userId")
	if userIDStr != "" && userIDStr != "anonymous" {
		// Try to parse user ID as uint
		if userID, err := strconv.ParseUint(userIDStr, 10, 32); err == nil {
			// Get user from database
			if user, err := d.store.Users.Get(d.server.Root, uint(userID)); err == nil {
				d.user = user
				fmt.Printf("Found user for callback: %s (ID: %d)\n", user.Username, user.ID)
			} else {
				fmt.Printf("Failed to get user with ID %d: %v\n", userID, err)
			}
		} else {
			fmt.Printf("Failed to parse user ID %s: %v\n", userIDStr, err)
		}
	}

	// Parse the callback data
	body, err := io.ReadAll(r.Body)

	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("failed to read request body: %v", err)
	}
	defer r.Body.Close()

	var callback OnlyOfficeCallback
	if err := json.Unmarshal(body, &callback); err != nil {
		return http.StatusBadRequest, fmt.Errorf("failed to parse callback data: %v", err)
	}

	// Log the callback for debugging
	fmt.Printf("OnlyOffice callback received: %+v\n", callback)

	// Handle different status codes
	switch callback.Status {
	case 1:
		// Document is being edited
		fmt.Println("Document is being edited")
	case 2:
		// Document is ready for saving
		fmt.Println("Document is ready for saving")
		if callback.URL != "" {
			if err := downloadAndSaveDocument(callback, d); err != nil {
				return http.StatusInternalServerError, fmt.Errorf("failed to save document: %v", err)
			}
		}
	case 3:
		// Document saving error
		fmt.Println("Document saving error occurred")
	case 4:
		// Document is closed with no changes
		fmt.Println("Document is closed with no changes")
	case 6:
		// Document is being edited, but current state is saved (force save)
		fmt.Println("Document force save")
		if callback.URL != "" {
			if err := downloadAndSaveDocument(callback, d); err != nil {
				return http.StatusInternalServerError, fmt.Errorf("failed to save document: %v", err)
			}
		}
	case 7:
		// Error occurred while force saving
		fmt.Println("Error occurred while force saving")
	default:
		fmt.Printf("Unknown status: %d\n", callback.Status)
	}

	// Return success response
	response := map[string]interface{}{
		"error": 0,
	}

	w.Header().Set("Content-Type", "application/json")
	return 0, json.NewEncoder(w).Encode(response)
}

// downloadAndSaveDocument downloads the document from OnlyOffice and saves it to the file system
func downloadAndSaveDocument(callback OnlyOfficeCallback, d *data) error {
	// Check if data and user are valid
	if d == nil {
		return fmt.Errorf("data parameter is nil")
	}
	if d.user == nil {
		return fmt.Errorf("user data is nil")
	}

	// Get the file path from the stored mapping
	filePath, exists := documentKeyMapping[callback.Key]
	if !exists {
		return fmt.Errorf("no file path mapping found for document key: %s", callback.Key)
	}

	// Download the document from OnlyOffice
	fmt.Printf("callbackURL:%s\n", callback.URL)
	resp, err := http.Get(callback.URL)
	if err != nil {
		return fmt.Errorf("failed to download document: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download document, status: %d", resp.StatusCode)
	}

	// The filePath from frontend is already the absolute path within user's scope
	// So we use it directly instead of joining with user.Scope again
	absPath := filepath.Join(d.user.Scope, filePath)
	fmt.Printf("Final path: %s\n", absPath)

	openFile, err := d.user.Fs.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, files.PermFile)
	if err != nil {
		return fmt.Errorf("could not open file: %v", err)
	}
	defer openFile.Close()

	_, err = openFile.Seek(0, 0)
	if err != nil {
		return fmt.Errorf("could not seek file: %v", err)
	}

	_, err = io.Copy(openFile, resp.Body)
	if err != nil {
		return fmt.Errorf("could not write to file: %v", err)
	}

	fmt.Printf("Document saved successfully to: %s\n", absPath)
	return nil
}

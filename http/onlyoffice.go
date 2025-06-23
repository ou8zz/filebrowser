package http

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

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

// OnlyOfficeConfigRequest represents the request to get OnlyOffice editor configuration
type OnlyOfficeConfigRequest struct {
	FilePath     string `json:"filePath"`
	FileName     string `json:"fileName"`
	FileModified string `json:"fileModified"`
	UserID       int    `json:"userId"`
	Username     string `json:"username"`
	Auth         string `json:"auth"`
	Origin       string `json:"origin"`
}

// OnlyOfficeConfig represents the complete OnlyOffice editor configuration
type OnlyOfficeConfig struct {
	DocumentType string                 `json:"documentType"`
	Document     OnlyOfficeDocument     `json:"document"`
	EditorConfig OnlyOfficeEditorConfig `json:"editorConfig"`
	Host         string                 `json:"host"`
	Token        string                 `json:"token"`
}

type OnlyOfficeDocument struct {
	Key         string                   `json:"key"`
	Title       string                   `json:"title"`
	URL         string                   `json:"url"`
	FileType    string                   `json:"fileType"`
	Permissions OnlyOfficeDocPermissions `json:"permissions"`
}

type OnlyOfficeDocPermissions struct {
	Edit     bool `json:"edit"`
	Download bool `json:"download"`
	Print    bool `json:"print"`
	Review   bool `json:"review"`
	Comment  bool `json:"comment"`
}

type OnlyOfficeEditorConfig struct {
	Mode          string                  `json:"mode"`
	Lang          string                  `json:"lang"`
	User          OnlyOfficeUser          `json:"user"`
	Customization OnlyOfficeCustomization `json:"customization"`
	CallbackURL   string                  `json:"callbackUrl"`
}

type OnlyOfficeUser struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type OnlyOfficeCustomization struct {
	Autosave  bool `json:"autosave"`
	Forcesave bool `json:"forcesave"`
}

// onlyOfficeMappingHandler handles OnlyOffice editor configuration requests
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

	// Parse the configuration request
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("failed to read request body: %v", err)
	}
	defer r.Body.Close()

	var configReq OnlyOfficeConfigRequest
	if err := json.Unmarshal(body, &configReq); err != nil {
		return http.StatusBadRequest, fmt.Errorf("failed to parse config request: %v", err)
	}
	cookie, _ := r.Cookie("auth")
	if cookie != nil && strings.Count(cookie.Value, ".") == 2 {
		configReq.Auth = strings.Replace(cookie.Value, "auth: ", "", 1)
	}

	// Generate OnlyOffice configuration
	configReq.Origin = r.Header.Get("Origin")
	config, err := generateOnlyOfficeConfig(configReq, d)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to generate config: %v", err)
	}

	// Store the document key mapping
	documentKeyMapping[config.Document.Key] = configReq.FilePath
	fmt.Printf("Stored document key mapping host:%s key:%s path:%s\n", configReq.Origin, config.Document.Key, configReq.FilePath)

	w.Header().Set("Content-Type", "application/json")
	return 0, json.NewEncoder(w).Encode(config)
}

// generateOnlyOfficeConfig generates the complete OnlyOffice editor configuration
func generateOnlyOfficeConfig(req OnlyOfficeConfigRequest, d *data) (*OnlyOfficeConfig, error) {
	// Generate document key
	documentKey := generateDocumentKey(req.FilePath, req.FileModified)

	// Get file extension and document type
	fileExt := getFileExtension(req.FileName)
	documentType := getDocumentType(fileExt)

	if !d.settings.OnlyOffice.Enabled || d.settings.OnlyOffice.Host == "" {
		return nil, fmt.Errorf("OnlyOffice服务未配置")
	}
	// Generate URLs
	documentURL := fmt.Sprintf("%s/api/raw%s?auth=%s", req.Origin, req.FilePath, req.Auth)
	callbackURL := fmt.Sprintf("%s/api/onlyoffice/callback?userId=%d", req.Origin, req.UserID)
	fmt.Println("documentURL:", documentURL)
	fmt.Println("callbackURL:", callbackURL)

	// Create configuration
	config := &OnlyOfficeConfig{
		DocumentType: documentType,
		Document: OnlyOfficeDocument{
			Key:      documentKey,
			Title:    req.FileName,
			URL:      documentURL,
			FileType: fileExt,
			Permissions: OnlyOfficeDocPermissions{
				Edit:     true,
				Download: true,
				Print:    true,
				Review:   true,
				Comment:  true,
			},
		},
		EditorConfig: OnlyOfficeEditorConfig{
			Mode: "edit",
			Lang: "zh-CN",
			User: OnlyOfficeUser{
				ID:   fmt.Sprintf("user_%d", req.UserID),
				Name: req.Username,
			},
			Customization: OnlyOfficeCustomization{
				Autosave:  true,
				Forcesave: false,
			},
			CallbackURL: callbackURL,
		},
		Host: d.settings.OnlyOffice.Host,
	}

	// Generate JWT token
	token, err := generateJWTToken(config, d.settings.OnlyOffice.JwtSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to generate JWT token: %v", err)
	}
	config.Token = token

	return config, nil
}

// generateDocumentKey generates a unique document key
func generateDocumentKey(filePath, fileModified string) string {
	return fmt.Sprintf("%s", base64.URLEncoding.EncodeToString([]byte(filePath+fileModified)))
}

// getFileExtension extracts file extension from filename
func getFileExtension(filename string) string {
	ext := filepath.Ext(filename)
	if len(ext) > 0 {
		return ext[1:] // Remove the dot
	}
	return ""
}

// getDocumentType determines OnlyOffice document type based on file extension
func getDocumentType(fileExt string) string {
	fileExt = strings.ToLower(fileExt)

	// Word documents
	wordExts := []string{"doc", "docx", "docm", "dot", "dotx", "dotm", "odt", "fodt", "ott", "rtf", "txt"}
	for _, ext := range wordExts {
		if fileExt == ext {
			return "word"
		}
	}

	// Excel documents
	excelExts := []string{"xls", "xlsx", "xlsm", "xlt", "xltx", "xltm", "ods", "fods", "ots", "csv"}
	for _, ext := range excelExts {
		if fileExt == ext {
			return "cell"
		}
	}

	// PowerPoint documents
	pptExts := []string{"ppt", "pptx", "pptm", "pot", "potx", "potm", "odp", "fodp", "otp"}
	for _, ext := range pptExts {
		if fileExt == ext {
			return "slide"
		}
	}

	return "word"
}

// generateJWTToken generates JWT token for OnlyOffice configuration
func generateJWTToken(config *OnlyOfficeConfig, jwtSecret string) (string, error) {
	// Create payload
	payload, err := json.Marshal(config)
	if err != nil {
		return "", fmt.Errorf("failed to marshal config: %v", err)
	}

	// Create header
	header := map[string]interface{}{
		"alg": "HS256",
		"typ": "JWT",
	}
	headerBytes, err := json.Marshal(header)
	if err != nil {
		return "", fmt.Errorf("failed to marshal header: %v", err)
	}

	// Encode header and payload
	headerEncoded := base64.RawURLEncoding.EncodeToString(headerBytes)
	payloadEncoded := base64.RawURLEncoding.EncodeToString(payload)

	// Create signature
	message := headerEncoded + "." + payloadEncoded
	h := hmac.New(sha256.New, []byte(jwtSecret))
	h.Write([]byte(message))
	signature := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

	// Combine all parts
	token := message + "." + signature
	return token, nil
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

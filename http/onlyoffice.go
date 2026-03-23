package fbhttp

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// OnlyOfficeCallback represents the callback data from OnlyOffice Document Server
type OnlyOfficeCallback struct {
	Key    string   `json:"key"`
	Status int      `json:"status"`
	URL    string   `json:"url"`
	Users  []string `json:"users"`
	Token  string   `json:"token"`
}

type documentMapping struct {
	Path   string
	UserID uint
}

var (
	documentKeyMappingMu sync.RWMutex
	documentKeyMapping   = make(map[string]documentMapping)
)

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
	if configReq.Origin == "" {
		scheme := "http"
		if proto := strings.TrimSpace(strings.Split(r.Header.Get("X-Forwarded-Proto"), ",")[0]); proto != "" {
			scheme = proto
		} else if r.TLS != nil {
			scheme = "https"
		}

		host := r.Host
		if forwardedHost := strings.TrimSpace(strings.Split(r.Header.Get("X-Forwarded-Host"), ",")[0]); forwardedHost != "" {
			host = forwardedHost
		}

		configReq.Origin = scheme + "://" + host
	}
	config, err := generateOnlyOfficeConfig(configReq, d)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to generate config: %v", err)
	}

	// Store the document key mapping
	documentKeyMappingMu.Lock()
	documentKeyMapping[config.Document.Key] = documentMapping{Path: configReq.FilePath, UserID: uint(configReq.UserID)}
	documentKeyMappingMu.Unlock()
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

	baseURL := strings.TrimRight(req.Origin, "/")
	if strings.TrimSpace(d.settings.OnlyOffice.FileBrowserURL) != "" {
		baseURL = strings.TrimRight(strings.TrimSpace(d.settings.OnlyOffice.FileBrowserURL), "/")
	}

	// Generate URLs
	documentURL := fmt.Sprintf("%s/api/raw%s?auth=%s", baseURL, req.FilePath, req.Auth)
	callbackURL := fmt.Sprintf("%s/api/onlyoffice/callback?userId=%d", baseURL, req.UserID)

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
				Forcesave: d.settings.OnlyOffice.ForceSave,
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

func onlyOfficeCallbackHandler(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return http.StatusOK, nil
	}

	if r.Method != http.MethodPost {
		return http.StatusMethodNotAllowed, fmt.Errorf("method not allowed")
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 2<<20))
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("failed to read request body: %v", err)
	}
	defer r.Body.Close()

	var callback OnlyOfficeCallback
	if err := json.Unmarshal(body, &callback); err != nil {
		return http.StatusBadRequest, fmt.Errorf("failed to parse callback data: %v", err)
	}

	if strings.TrimSpace(d.settings.OnlyOffice.JwtSecret) != "" {
		if strings.TrimSpace(callback.Token) == "" {
			return http.StatusUnauthorized, fmt.Errorf("missing onlyoffice callback token")
		}
		parsed, err := jwt.Parse(callback.Token, func(t *jwt.Token) (interface{}, error) {
			if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
				return nil, fmt.Errorf("unexpected jwt alg: %s", t.Method.Alg())
			}
			return []byte(d.settings.OnlyOffice.JwtSecret), nil
		}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
		if err != nil || parsed == nil || !parsed.Valid {
			return http.StatusUnauthorized, fmt.Errorf("invalid onlyoffice callback token")
		}
	}

	documentKeyMappingMu.RLock()
	mapping, exists := documentKeyMapping[callback.Key]
	documentKeyMappingMu.RUnlock()
	if !exists {
		return http.StatusBadRequest, fmt.Errorf("no file mapping for document key: %s", callback.Key)
	}
	if mapping.UserID == 0 {
		return http.StatusBadRequest, fmt.Errorf("invalid user mapping for document key: %s", callback.Key)
	}

	user, err := d.store.Users.Get(d.server.Root, mapping.UserID)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	d.user = user

	switch callback.Status {
	case 2, 6:
		if callback.URL != "" {
			if err := downloadAndSaveDocument(callback, mapping.Path, d); err != nil {
				return http.StatusInternalServerError, fmt.Errorf("failed to save document: %v", err)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	return 0, json.NewEncoder(w).Encode(map[string]int{"error": 0})
}

func downloadAndSaveDocument(callback OnlyOfficeCallback, filePath string, d *data) error {
	if d == nil {
		return fmt.Errorf("data parameter is nil")
	}
	if d.user == nil {
		return fmt.Errorf("user data is nil")
	}

	if callback.URL == "" {
		return fmt.Errorf("empty download url")
	}

	cbURL, err := url.Parse(callback.URL)
	if err != nil || cbURL.Hostname() == "" {
		return fmt.Errorf("invalid download url")
	}

	onlyofficeHostStr := strings.TrimSpace(d.settings.OnlyOffice.Host)
	if onlyofficeHostStr == "" {
		return fmt.Errorf("onlyoffice host not configured")
	}
	if !strings.Contains(onlyofficeHostStr, "://") {
		onlyofficeHostStr = "http://" + onlyofficeHostStr
	}
	ooURL, err := url.Parse(onlyofficeHostStr)
	if err != nil || ooURL.Hostname() == "" {
		return fmt.Errorf("invalid onlyoffice host")
	}
	if !strings.EqualFold(cbURL.Hostname(), ooURL.Hostname()) {
		return fmt.Errorf("download host mismatch")
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Get(callback.URL)
	if err != nil {
		return fmt.Errorf("failed to download document: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download document, status: %d", resp.StatusCode)
	}

	openFile, err := d.user.Fs.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, d.settings.FileMode)
	if err != nil {
		return fmt.Errorf("could not open file: %v", err)
	}
	defer openFile.Close()

	_, err = io.Copy(openFile, io.LimitReader(resp.Body, 1<<30))
	if err != nil {
		return fmt.Errorf("could not write to file: %v", err)
	}

	return nil
}

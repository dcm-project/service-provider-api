package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/dcm-project/service-provider-api/pkg/registration/client"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// Simple File Provider - manages files as resources
func main() {
	log.Println("")
	log.Println("╔═══════════════════════════════════════════════════════════╗")
	log.Println("║        FILE STORAGE PROVIDER (Demo)                      ║")
	log.Println("╚═══════════════════════════════════════════════════════════╝")
	log.Println("")
	log.Println("📋 What this does:")
	log.Println("   • Provides file storage service (CREATE/READ/DELETE)")
	log.Println("   • Auto-registers with DCM on startup")
	log.Println("   • Fulfills MULTIPLE catalog items: 'file' + 'container'")
	log.Println("   • Auto-unregisters on shutdown")
	log.Println("")

	time.Sleep(3 * time.Second)
	log.Println("")

	serviceID := uuid.New().String()
	providerAddr := "localhost:8081"
	dcmURL := getEnvOrDefault("DCM_URL", "http://localhost:9090")
	zone := getEnvOrDefault("ZONE", "datacenter-east")
	region := getEnvOrDefault("REGION", "us-east")

	// Create storage directory
	storageDir := "/tmp/file-provider-storage"
	os.MkdirAll(storageDir, 0755)

	provider := &FileProvider{
		serviceID:  serviceID,
		storageDir: storageDir,
	}

	// Start provider API server
	router := mux.NewRouter()
	router.HandleFunc("/health", provider.Health).Methods("GET")
	router.HandleFunc("/api/file", provider.CreateFile).Methods("POST")
	router.HandleFunc("/api/file/{id}", provider.DeleteFile).Methods("DELETE")
	router.HandleFunc("/api/file/{id}", provider.GetFile).Methods("GET")

	server := &http.Server{
		Addr:    providerAddr,
		Handler: router,
	}

	log.Println("🚀 Starting provider service...")
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for server to be ready
	time.Sleep(1 * time.Second)
	log.Printf("   • Service running on: http://%s", providerAddr)
	log.Println("")

	// Register with DCM
	regClient := client.New(client.Config{
		BaseURL: dcmURL,
	})

	ctx := context.Background()
	log.Println("📝 Registering with DCM...")
	log.Printf("   • Service ID: %s", serviceID)
	log.Printf("   • DCM URL: %s", dcmURL)
	log.Printf("   • Zone/Region: %s/%s", zone, region)
	log.Println("")

	// === FIRST REGISTRATION: FILE ===
	log.Println("═══ Registration 1/2: FILE resource type ═══")
	log.Println("")

	fileReq := &client.RegistrationRequest{
		ServiceID: serviceID, // Same ID for both
		Endpoint:  fmt.Sprintf("http://%s/api/file", providerAddr),
		Metadata: client.Metadata{
			Zone:   zone,
			Region: region,
		},
		Operations: []string{"CREATE", "DELETE", "READ"},
	}

	log.Println("🔧 API Call #1:")
	log.Printf("   Method:   POST")
	log.Printf("   Endpoint: %s/resource/file/provider", dcmURL)
	log.Println("   Payload:")
	payloadJSON, _ := json.MarshalIndent(fileReq, "   ", "  ")
	log.Printf("%s", string(payloadJSON))
	log.Println("")
	log.Println("   ⏳ Sending registration request...")

	time.Sleep(3 * time.Second)
	log.Println("")

	fileResp, err := regClient.Register(ctx, "file", fileReq)
	if err != nil {
		log.Fatalf("❌ File registration failed: %v", err)
	}

	log.Println("📨 API Response #1:")
	log.Println("   HTTP Status: 200 OK")
	log.Println("   Response Body:")
	responseJSON, _ := json.MarshalIndent(fileResp, "   ", "  ")
	log.Printf("%s", string(responseJSON))
	log.Println("")
	log.Println("   ✅ Registered for 'file' resource type")
	log.Println("")

	time.Sleep(2 * time.Second)

	// === SECOND REGISTRATION: CONTAINER ===
	log.Println("═══ Registration 2/2: CONTAINER resource type ═══")
	log.Println("")

	containerReq := &client.RegistrationRequest{
		ServiceID: serviceID, // SAME ID - this updates the existing entry
		Endpoint:  fmt.Sprintf("http://%s/api/container", providerAddr),
		Metadata: client.Metadata{
			Zone:   zone,
			Region: region,
		},
		Operations: []string{"CREATE", "DELETE", "READ"},
	}

	log.Println("🔧 API Call #2:")
	log.Printf("   Method:   POST")
	log.Printf("   Endpoint: %s/resource/container/provider", dcmURL)
	log.Println("   Payload:")
	payloadJSON2, _ := json.MarshalIndent(containerReq, "   ", "  ")
	log.Printf("%s", string(payloadJSON2))
	log.Println("")
	log.Println("   ⏳ Sending registration request...")

	time.Sleep(3 * time.Second)
	log.Println("")

	containerResp, err := regClient.Register(ctx, "container", containerReq)
	if err != nil {
		log.Fatalf("❌ Container registration failed: %v", err)
	}

	log.Println("📨 API Response #2:")
	log.Println("   HTTP Status: 200 OK")
	log.Println("   Response Body:")
	responseJSON2, _ := json.MarshalIndent(containerResp, "   ", "  ")
	log.Printf("%s", string(responseJSON2))
	log.Println("")
	log.Println("   ✅ Registered for 'container' resource type")
	log.Println("")

	time.Sleep(2 * time.Second)

	log.Println("✅ ALL REGISTRATIONS SUCCESSFUL")
	log.Println("")
	log.Println("📊 Result:")
	log.Println("   • Service Registry: 1 provider entry with 2 resource type registrations")
	log.Println("     {")
	log.Printf("       \"service_id\": \"%s\",\n", serviceID)
	log.Println("       \"registrations\": [")
	log.Println("         { \"resource_kind\": \"file\", \"endpoint\": \".../api/file\" },")
	log.Println("         { \"resource_kind\": \"container\", \"endpoint\": \".../api/container\" }")
	log.Println("       ]")
	log.Println("     }")
	log.Println("")
	log.Println("   • Service Catalog: Provider appears in 2 catalog items (file, container)")
	log.Println("     Catalog item is derived from resource_kind (no redundancy)")
	log.Println("")
	log.Println("💡 Key: SAME service ID + separate registrations = ONE provider, multiple capabilities")
	log.Println("")
	log.Println("   View registry: curl http://localhost:9090/admin/registry | jq")
	log.Println("   View catalog:  curl http://localhost:9090/admin/catalog | jq")
	log.Println("")
	log.Println("Press Ctrl+C to unregister and stop")
	log.Println("")

	// Wait for shutdown signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	log.Println("")
	log.Println("🛑 Shutting down provider...")
	log.Println("")

	// Unregister from DCM
	log.Println("📝 Unregistering from DCM...")
	log.Println("   Unregistering from both resource types")
	log.Println("")

	// === FIRST UNREGISTRATION: FILE ===
	log.Println("═══ Unregistration 1/2: FILE resource type ═══")
	log.Println("")
	log.Println("🔧 API Call #1:")
	log.Printf("   Method:   DELETE")
	log.Printf("   Endpoint: %s/resource/file/provider/%s", dcmURL, serviceID)
	log.Println("")
	log.Println("   ⏳ Sending unregister request...")

	time.Sleep(2 * time.Second)
	log.Println("")

	if err := regClient.Unregister(ctx, "file", serviceID); err != nil {
		log.Printf("   ⚠️  Failed to unregister from file: %v", err)
	} else {
		log.Println("📨 API Response #1:")
		log.Println("   HTTP Status: 204 No Content")
		log.Println("")
		log.Println("   ✅ Unregistered from 'file' resource type")
	}
	log.Println("")

	time.Sleep(1 * time.Second)

	// === SECOND UNREGISTRATION: CONTAINER ===
	log.Println("═══ Unregistration 2/2: CONTAINER resource type ═══")
	log.Println("")
	log.Println("🔧 API Call #2:")
	log.Printf("   Method:   DELETE")
	log.Printf("   Endpoint: %s/resource/container/provider/%s", dcmURL, serviceID)
	log.Println("")
	log.Println("   ⏳ Sending unregister request...")

	time.Sleep(2 * time.Second)
	log.Println("")

	if err := regClient.Unregister(ctx, "container", serviceID); err != nil {
		log.Printf("   ⚠️  Failed to unregister from container: %v", err)
	} else {
		log.Println("📨 API Response #2:")
		log.Println("   HTTP Status: 204 No Content")
		log.Println("")
		log.Println("   ✅ Unregistered from 'container' resource type")
	}
	log.Println("")

	time.Sleep(1 * time.Second)

	log.Println("✅ ALL UNREGISTRATIONS SUCCESSFUL")
	log.Println("")
	log.Println("📊 Result:")
	log.Println("   • Provider removed from Service Registry")
	log.Println("   • Provider removed from 'file' catalog")
	log.Println("   • Provider removed from 'container' catalog")
	log.Println("")

	// Shutdown server
	log.Println("   • Stopping HTTP server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Shutdown(ctx)

	log.Println("")
	log.Println("👋 File Provider stopped cleanly")
	log.Println("")
}

type FileProvider struct {
	serviceID  string
	storageDir string
}

type CreateFileRequest struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

type FileResource struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Path    string `json:"path"`
	Created string `json:"created"`
}

func (p *FileProvider) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":     "healthy",
		"service_id": p.serviceID,
	})
}

func (p *FileProvider) CreateFile(w http.ResponseWriter, r *http.Request) {
	var req CreateFileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fileID := uuid.New().String()
	filePath := filepath.Join(p.storageDir, fileID)

	if err := os.WriteFile(filePath, []byte(req.Content), 0644); err != nil {
		http.Error(w, fmt.Sprintf("Failed to create file: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("📄 Created file: %s (name: %s)", fileID, req.Name)

	resource := FileResource{
		ID:      fileID,
		Name:    req.Name,
		Path:    filePath,
		Created: time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resource)
}

func (p *FileProvider) DeleteFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fileID := vars["id"]

	filePath := filepath.Join(p.storageDir, fileID)

	if err := os.Remove(filePath); err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "File not found", http.StatusNotFound)
		} else {
			http.Error(w, fmt.Sprintf("Failed to delete file: %v", err), http.StatusInternalServerError)
		}
		return
	}

	log.Printf("🗑️  Deleted file: %s", fileID)

	w.WriteHeader(http.StatusNoContent)
}

func (p *FileProvider) GetFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fileID := vars["id"]

	filePath := filepath.Join(p.storageDir, fileID)

	content, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "File not found", http.StatusNotFound)
		} else {
			http.Error(w, fmt.Sprintf("Failed to read file: %v", err), http.StatusInternalServerError)
		}
		return
	}

	info, _ := os.Stat(filePath)

	resource := FileResource{
		ID:      fileID,
		Name:    fileID,
		Path:    filePath,
		Created: info.ModTime().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"resource": resource,
		"content":  string(content),
	})
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

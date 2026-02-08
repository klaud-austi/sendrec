package main

import (
	"embed"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"regexp"
	"sync"
	"time"
)

//go:embed static/*
var staticFiles embed.FS

type WaitlistEntry struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type WaitlistRequest struct {
	Email string `json:"email"`
}

type WaitlistResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type Store struct {
	mu      sync.RWMutex
	entries []WaitlistEntry
	file    string
	nextID  int64
}

func NewStore(filename string) (*Store, error) {
	s := &Store{file: filename, nextID: 1}
	if _, err := os.Stat(filename); err == nil {
		data, err := os.ReadFile(filename)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(data, &s.entries); err != nil {
			return nil, err
		}
		// Find max ID
		for _, e := range s.entries {
			if e.ID >= s.nextID {
				s.nextID = e.ID + 1
			}
		}
	}
	return s, nil
}

func (s *Store) Save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	data, err := json.MarshalIndent(s.entries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.file, data, 0644)
}

func (s *Store) Add(email string) (*WaitlistEntry, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Check if email already exists
	for _, e := range s.entries {
		if e.Email == email {
			return nil, false
		}
	}
	
	entry := WaitlistEntry{
		ID:        s.nextID,
		Email:     email,
		CreatedAt: time.Now(),
	}
	s.nextID++
	s.entries = append(s.entries, entry)
	
	// Save asynchronously
	go s.Save()
	
	return &entry, true
}

func (s *Store) GetAll() []WaitlistEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	// Return reverse order (newest first)
	result := make([]WaitlistEntry, len(s.entries))
	for i, e := range s.entries {
		result[len(s.entries)-1-i] = e
	}
	return result
}

var store *Store

func isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

func joinWaitlist(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req WaitlistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(WaitlistResponse{Success: false, Message: "Invalid JSON"})
		return
	}

	if req.Email == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(WaitlistResponse{Success: false, Message: "Email is required"})
		return
	}

	if !isValidEmail(req.Email) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(WaitlistResponse{Success: false, Message: "Invalid email format"})
		return
	}

	entry, added := store.Add(req.Email)
	if !added {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(WaitlistResponse{Success: false, Message: "Email already registered"})
		return
	}

	log.Printf("New waitlist entry: ID=%d, Email=%s", entry.ID, req.Email)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(WaitlistResponse{Success: true, Message: "Successfully joined the waitlist!"})
}

func viewWaitlist(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	entries := store.GetAll()

	html := `<!DOCTYPE html>
<html>
<head>
    <title>SendRec Waitlist</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; max-width: 900px; margin: 50px auto; padding: 20px; }
        h1 { color: #333; }
        table { width: 100%; border-collapse: collapse; margin-top: 20px; }
        th, td { text-align: left; padding: 12px; border-bottom: 1px solid #ddd; }
        th { background-color: #f5f5f5; font-weight: 600; }
        tr:hover { background-color: #f9f9f9; }
        .count { color: #666; margin-top: 10px; }
    </style>
</head>
<body>
    <h1>📧 SendRec Waitlist</h1>
    <div class="count">Total subscribers: {{len .}}</div>
    <table>
        <thead>
            <tr>
                <th>ID</th>
                <th>Email</th>
                <th>Joined</th>
            </tr>
        </thead>
        <tbody>
            {{range .}}
            <tr>
                <td>{{.ID}}</td>
                <td>{{.Email}}</td>
                <td>{{.CreatedAt.Format "2006-01-02 15:04"}}</td>
            </tr>
            {{else}}
            <tr>
                <td colspan="3" style="text-align: center; color: #999;">No entries yet</td>
            </tr>
            {{end}}
        </tbody>
    </table>
</body>
</html>`

	tmpl, err := template.New("admin").Parse(html)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	tmpl.Execute(w, entries)
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func main() {
	var err error
	store, err = NewStore("./data/waitlist.json")
	if err != nil {
		log.Fatal("Failed to initialize store:", err)
	}

	// API routes
	http.HandleFunc("/waitlist", joinWaitlist)
	http.HandleFunc("/admin", viewWaitlist)
	http.HandleFunc("/health", healthCheck)

	// Static files - serve landing page and assets
	http.Handle("/", http.FileServer(http.FS(staticFiles)))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	
	log.Printf("SendRec server starting on http://localhost:%s", port)
	log.Printf("Endpoints:")
	log.Printf("  GET  /              - Landing page")
	log.Printf("  POST /waitlist      - Join waitlist")
	log.Printf("  GET  /admin         - View waitlist")
	log.Printf("  GET  /health        - Health check")

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("Server failed:", err)
	}
}

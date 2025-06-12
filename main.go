package main

import (
	"encoding/json"
	"go-log-view/pkg/sshclient"
	"go-log-view/pkg/websocket"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Servers []ServerConfig `yaml:"servers"`
}

type ServerConfig struct {
	Name     string    `yaml:"name"`
	Host     string    `yaml:"host"`
	Port     int       `yaml:"port"`
	Username string    `yaml:"username"`
	Password string    `yaml:"password"`
	KeyPath  string    `yaml:"key_path"`
	LogFiles []LogFile `yaml:"log_files"`
}

type LogFile struct {
	Path  string `yaml:"path"`
	Alias string `yaml:"alias"`
}

type LogStream struct {
	ServerName string
	FileAlias  string
	Reader     io.Reader
	Closer     io.Closer
}

var (
	config      Config
	wsServer    *websocket.WebSocketServer
	logStreams  = make(map[string]*LogStream)
	streamMutex sync.Mutex
)

func main() {
	// 加载配置文件
	loadConfig()

	// 初始化WebSocket服务器
	wsServer = websocket.NewWebSocketServer()
	go wsServer.Run()

	// 设置HTTP路由
	router := mux.NewRouter()
	router.HandleFunc("/ws", wsServer.HandleWebSocket)
	router.HandleFunc("/api/servers", getServersHandler).Methods("GET")
	router.HandleFunc("/api/log/start", startLogHandler).Methods("POST")
	router.HandleFunc("/api/log/stop", stopLogHandler).Methods("POST")
	router.HandleFunc("/api/command", executeCommandHandler).Methods("POST")

	// 静态文件服务（Vue前端）
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./web/dist")))

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}

func loadConfig() {
	configPath := filepath.Join("config", "config.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Fatalf("Error parsing config file: %v", err)
	}
}

func getServersHandler(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(config.Servers)
}

func startLogHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ServerName string `json:"serverName"`
		FileAlias  string `json:"fileAlias"`
		Lines      int    `json:"lines"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 查找服务器和日志文件配置
	var serverConfig *ServerConfig
	var logFile *LogFile

	for _, s := range config.Servers {
		if s.Name == req.ServerName {
			serverConfig = &s
			for _, lf := range s.LogFiles {
				if lf.Alias == req.FileAlias {
					logFile = &lf
					break
				}
			}
			break
		}
	}

	if serverConfig == nil || logFile == nil {
		http.Error(w, "Server or log file not found", http.StatusNotFound)
		return
	}

	// 建立SSH连接
	client, err := sshclient.NewSSHClient(
		serverConfig.Host,
		serverConfig.Port,
		serverConfig.Username,
		serverConfig.Password,
		serverConfig.KeyPath,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 创建管道用于实时传输日志
	reader, writer := io.Pipe()

	// 启动日志跟踪
	go func() {
		defer writer.Close()
		err := client.TailFile(logFile.Path, req.Lines, writer)
		if err != nil {
			log.Printf("Error tailing file: %v", err)
		}
	}()

	streamID := req.ServerName + ":" + req.FileAlias

	streamMutex.Lock()
	logStreams[streamID] = &LogStream{
		ServerName: req.ServerName,
		FileAlias:  req.FileAlias,
		Reader:     reader,
		Closer:     writer,
	}
	streamMutex.Unlock()

	// 启动goroutine读取日志并发送到WebSocket
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := reader.Read(buf)
			if err != nil {
				log.Printf("Error reading log stream: %v", err)
				streamMutex.Lock()
				delete(logStreams, streamID)
				streamMutex.Unlock()
				return
			}

			message := map[string]interface{}{
				"type":      "log",
				"server":    req.ServerName,
				"file":      req.FileAlias,
				"content":   string(buf[:n]),
				"timestamp": time.Now().Unix(),
			}

			jsonMessage, _ := json.Marshal(message)
			wsServer.Broadcast <- jsonMessage
		}
	}()

	w.WriteHeader(http.StatusOK)
}

func stopLogHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ServerName string `json:"serverName"`
		FileAlias  string `json:"fileAlias"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	streamID := req.ServerName + ":" + req.FileAlias

	streamMutex.Lock()
	if stream, ok := logStreams[streamID]; ok {
		if closer, ok := stream.Closer.(io.Closer); ok {
			closer.Close()
		}
		delete(logStreams, streamID)
	}
	streamMutex.Unlock()

	w.WriteHeader(http.StatusOK)
}

func executeCommandHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ServerName string `json:"serverName"`
		Command    string `json:"command"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 查找服务器配置
	var serverConfig *ServerConfig
	for _, s := range config.Servers {
		if s.Name == req.ServerName {
			serverConfig = &s
			break
		}
	}

	if serverConfig == nil {
		http.Error(w, "Server not found", http.StatusNotFound)
		return
	}

	// 建立SSH连接并执行命令
	client, err := sshclient.NewSSHClient(
		serverConfig.Host,
		serverConfig.Port,
		serverConfig.Username,
		serverConfig.Password,
		serverConfig.KeyPath,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer client.Close()

	output, err := client.ExecuteCommand(req.Command)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"output": output,
	}
	json.NewEncoder(w).Encode(response)
}

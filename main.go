package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
)

type Stats struct {
	CPU        float64       `json:"cpu"`
	RAM        float64       `json:"ram"`
	GPU        []GPUStats    `json:"gpu"`
	Users      []UserStats   `json:"users"`
	Processes  []ProcStats   `json:"processes"`
	Timestamp  time.Time     `json:"timestamp"`
}

type GPUStats struct {
	Name        string `json:"name"`
	Util        string `json:"util"`
	MemoryUsed  string `json:"memory_used"`
	MemoryTotal string `json:"memory_total"`
	Temp        string `json:"temp"`
}

type UserStats struct {
	User     string `json:"user"`
	Terminal string `json:"terminal"`
	Source   string `json:"source"`
	LoginAt  string `json:"login_at"`
}

type ProcStats struct {
	PID     int32   `json:"pid"`
	User    string  `json:"user"`
	CPU     float64 `json:"cpu"`
	Memory  float32 `json:"memory"`
	Command string  `json:"command"`
}

var (
	stats      Stats
	statsMutex sync.RWMutex
	upgrader   = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
)

func main() {
	go collectStats()

	r := gin.Default()

	// Serve static files (frontend)
	r.Static("/assets", "./dist/assets")
	r.StaticFile("/", "./dist/index.html")

	// WebSocket endpoint
	r.GET("/ws", func(c *gin.Context) {
		ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Println("upgrade error:", err)
			return
		}
		defer ws.Close()

		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				statsMutex.RLock()
				data, _ := json.Marshal(stats)
				statsMutex.RUnlock()
				if err := ws.WriteMessage(websocket.TextMessage, data); err != nil {
					return
				}
			}
		}
	})

	log.Println("Server starting on :2000")
	r.Run(":2000")
}

func collectStats() {
	for {
		newStats := Stats{
			Timestamp: time.Now(),
		}

		// CPU
		c, _ := cpu.Percent(0, false)
		if len(c) > 0 {
			newStats.CPU = c[0]
		}

		// RAM
		v, _ := mem.VirtualMemory()
		newStats.RAM = v.UsedPercent

		// GPU (NVIDIA)
		newStats.GPU = getGPUStats()

		// Users (SSH)
		newStats.Users = getUserStats()

		// Processes
		newStats.Processes = getProcessStats()

		statsMutex.Lock()
		stats = newStats
		statsMutex.Unlock()

		time.Sleep(3 * time.Second)
	}
}

func getGPUStats() []GPUStats {
	cmd := exec.Command("nvidia-smi", "--query-gpu=name,utilization.gpu,memory.used,memory.total,temperature.gpu", "--format=csv,noheader,nounits")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var gpus []GPUStats
	for _, line := range lines {
		fields := strings.Split(line, ", ")
		if len(fields) >= 5 {
			gpus = append(gpus, GPUStats{
				Name:        fields[0],
				Util:        fields[1] + "%",
				MemoryUsed:  fields[2] + "MiB",
				MemoryTotal: fields[3] + "MiB",
				Temp:        fields[4] + "C",
			})
		}
	}
	return gpus
}

func getUserStats() []UserStats {
	cmd := exec.Command("who", "-u")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var users []UserStats
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 5 {
			source := ""
			if len(fields) >= 6 {
				source = strings.Trim(fields[len(fields)-1], "()")
			}
			users = append(users, UserStats{
				User:     fields[0],
				Terminal: fields[1],
				LoginAt:  fields[2] + " " + fields[3],
				Source:   source,
			})
		}
	}
	return users
}

func getProcessStats() []ProcStats {
	procs, err := process.Processes()
	if err != nil {
		return nil
	}

	var stats []ProcStats
	for _, p := range procs {
		user, _ := p.Username()
		cpuPercent, _ := p.CPUPercent()
		memPercent, _ := p.MemoryPercent()
		cmd, _ := p.Cmdline()

		if cpuPercent > 0.1 || memPercent > 0.1 {
			stats = append(stats, ProcStats{
				PID:     p.Pid,
				User:    user,
				CPU:     cpuPercent,
				Memory:  memPercent,
				Command: cmd,
			})
		}
	}
	return stats
}

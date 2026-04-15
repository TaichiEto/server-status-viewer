package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
)

type Stats struct {
	System     HostInfo      `json:"system"`
	CPU        CPUInfo       `json:"cpu"`
	RAM        RAMInfo       `json:"ram"`
	Disk       []DiskInfo    `json:"disk"`
	GPU        []GPUStats    `json:"gpu"`
	Users      []UserStats   `json:"users"`
	Processes  []ProcStats   `json:"processes"`
	Timestamp  time.Time     `json:"timestamp"`
}

type HostInfo struct {
	Hostname string `json:"hostname"`
	OS       string `json:"os"`
	Kernel   string `json:"kernel"`
	Uptime   uint64 `json:"uptime"`
}

type CPUInfo struct {
	Model    string    `json:"model"`
	Cores    int       `json:"cores"` // Physical
	Logical  int       `json:"logical"`
	Usage    float64   `json:"usage"` // Total
	PerCore  []float64 `json:"per_core"`
}

type RAMInfo struct {
	Total     uint64  `json:"total"`
	Used      uint64  `json:"used"`
	Percent   float64 `json:"percent"`
}

type DiskInfo struct {
	Path    string  `json:"path"`
	Total   uint64  `json:"total"`
	Used    uint64  `json:"used"`
	Percent float64 `json:"percent"`
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

	// Serve static files from the dist directory
	r.Static("/assets", "./dist/assets")
	
	// Handle SPA routing
	r.NoRoute(func(c *gin.Context) {
		c.File("./dist/index.html")
	})

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
			Disk:      make([]DiskInfo, 0),
			GPU:       make([]GPUStats, 0),
			Users:     make([]UserStats, 0),
			Processes: make([]ProcStats, 0),
		}

		// Host Info
		h, _ := host.Info()
		newStats.System = HostInfo{
			Hostname: h.Hostname,
			OS:       h.OS,
			Kernel:   h.KernelVersion,
			Uptime:   h.Uptime,
		}

		// CPU
		cPercent, _ := cpu.Percent(0, false)
		perCore, _ := cpu.Percent(0, true)
		cInfo, _ := cpu.Info()
		logical, _ := cpu.Counts(true)
		physical, _ := cpu.Counts(false)
		
		model := "Unknown"
		if len(cInfo) > 0 {
			model = cInfo[0].ModelName
		}

		newStats.CPU = CPUInfo{
			Model:   model,
			Cores:   physical,
			Logical: logical,
			Usage:   0,
			PerCore: make([]float64, 0),
		}
		if len(cPercent) > 0 {
			newStats.CPU.Usage = cPercent[0]
		}
		if perCore != nil {
			newStats.CPU.PerCore = perCore
		}

		// RAM
		v, _ := mem.VirtualMemory()
		newStats.RAM = RAMInfo{
			Total:   v.Total,
			Used:    v.Used,
			Percent: v.UsedPercent,
		}

		// Disk
		partitions, _ := disk.Partitions(false)
		for _, p := range partitions {
			if strings.HasPrefix(p.Mountpoint, "/boot") || strings.Contains(p.Mountpoint, "loop") {
				continue
			}
			u, err := disk.Usage(p.Mountpoint)
			if err == nil {
				newStats.Disk = append(newStats.Disk, DiskInfo{
					Path:    p.Mountpoint,
					Total:   u.Total,
					Used:    u.Used,
					Percent: u.UsedPercent,
				})
			}
		}

		// GPU, Users, Processes
		newStats.GPU = getGPUStats()
		newStats.Users = getUserStats()
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
		return make([]GPUStats, 0)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	gpus := make([]GPUStats, 0)
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
		return make([]UserStats, 0)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	users := make([]UserStats, 0)
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
		return make([]ProcStats, 0)
	}

	stats := make([]ProcStats, 0)
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

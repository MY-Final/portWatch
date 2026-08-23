package service

import (
	"path/filepath"
	"strings"

	"github.com/MY-Final/portWatch/pkg/model"
)

type Info struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	Confidence int    `json:"confidence"`
}

type Detector interface {
	Detect(port model.PortInfo, process model.ProcessInfo) Info
}

type Rules struct{}

func (Rules) Detect(port model.PortInfo, process model.ProcessInfo) Info {
	name := strings.ToLower(filepath.Base(process.Executable) + " " + process.Name + " " + process.Command)
	workingDir := strings.ToLower(process.WorkingDir)
	switch {
	case strings.Contains(name, "uvicorn") || strings.Contains(name, "fastapi"):
		return Info{Name: "FastAPI / Uvicorn", Type: "Python", Confidence: 90}
	case strings.Contains(name, "nacos"):
		return Info{Name: "Nacos", Type: "Java", Confidence: 95}
	case strings.Contains(name, "spring") || strings.Contains(name, ".jar"):
		return Info{Name: "Spring Boot", Type: "Java", Confidence: 75}
	case strings.Contains(name, "vite"):
		return Info{Name: "Vite", Type: "Node.js", Confidence: 95}
	case port.Port == 5173 && strings.Contains(name, "node"):
		return Info{Name: "Vite", Type: "Node.js", Confidence: 80}
	case port.Port == 8848 && strings.Contains(name, "java"):
		return Info{Name: "Nacos", Type: "Java", Confidence: 70}
	case strings.Contains(name, "node") && (strings.Contains(workingDir, "frontend") || strings.Contains(workingDir, "web")):
		return Info{Name: "Node.js", Type: "Node.js", Confidence: 70}
	case strings.Contains(name, "node"):
		return Info{Name: "Node.js", Type: "Node.js", Confidence: 60}
	default:
		return Info{Name: "Unknown", Type: "Unknown", Confidence: 0}
	}
}

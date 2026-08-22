package service

import (
	"path/filepath"
	"strings"

	"github.com/portwatch/portwatch/pkg/model"
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

func (Rules) Detect(_ model.PortInfo, process model.ProcessInfo) Info {
	name := strings.ToLower(filepath.Base(process.Executable) + " " + process.Name + " " + process.Command)
	switch {
	case strings.Contains(name, "uvicorn") || strings.Contains(name, "fastapi"):
		return Info{Name: "FastAPI / Uvicorn", Type: "Python", Confidence: 90}
	case strings.Contains(name, "spring") || strings.Contains(name, ".jar"):
		return Info{Name: "Spring Boot", Type: "Java", Confidence: 75}
	case strings.Contains(name, "vite"):
		return Info{Name: "Vite", Type: "Node.js", Confidence: 95}
	case strings.Contains(name, "node"):
		return Info{Name: "Node.js", Type: "Node.js", Confidence: 60}
	default:
		return Info{Name: "Unknown", Type: "Unknown", Confidence: 0}
	}
}

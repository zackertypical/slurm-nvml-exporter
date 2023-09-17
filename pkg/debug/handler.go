package debug

import (
	"encoding/json"
	"net/http"

	"github.com/nvml-exporter/pkg/collector"
)

// todo: server
type DebugHandler struct {
	cache *collector.NVMLCache
}

func HandlerFor(cache *collector.NVMLCache) http.Handler {
	return DebugHandler{
		cache: cache,
	}
}

func (h DebugHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// 根据请求路径选择处理函数
	switch r.URL.Path {
	case "/debug/gpuinfo":
		h.handleGPUInfo(w, r)
	case "/debug/gpustat":
		h.handleGPUStat(w, r)
	case "/debug/process":
		h.handleProcess(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (h DebugHandler) handleGPUInfo(w http.ResponseWriter, r *http.Request) {
	// 处理 /debug/gpuinfo 请求
	info := h.cache.GetGPUInfos()
	jsonResponse(w, info)
}

func (h DebugHandler) handleGPUStat(w http.ResponseWriter, r *http.Request) {
	// 处理 /debug/gpustat 请求
	info := h.cache.GetGPUStats()
	jsonResponse(w, info)
}

func (h DebugHandler) handleProcess(w http.ResponseWriter, r *http.Request) {
	// 处理 /debug/process 请求
	info := h.cache.GetProcessStats()
	jsonResponse(w, info)
}

func jsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

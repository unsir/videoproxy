package main

import (
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"time"
	"videoproxy/route"
	"videoproxy/util"
)

var (
	startTime = time.Now()
)

var sysStatus struct {
	Uptime       string
	GoVersion    string
	MemAllocated uint64
	MemTotal     uint64
	MemSys       uint64
	NumGoroutine int
	CPUNum       int
	Pid          int
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	util.Log.Fatal(serve(port))
}

func serve(port string) error {
	http.HandleFunc("/", routeMatch)
	http.HandleFunc("/status", status)
	util.Log.Printf("Starting up on port %s", port)
	return http.ListenAndServe(":"+port, nil)
}

func status(w http.ResponseWriter, r *http.Request) {
	memStat := new(runtime.MemStats)
	runtime.ReadMemStats(memStat)
	sysStatus.Uptime = time.Since(startTime).String()
	sysStatus.NumGoroutine = runtime.NumGoroutine()
	sysStatus.MemAllocated = memStat.Alloc
	sysStatus.MemTotal = memStat.TotalAlloc
	sysStatus.MemSys = memStat.Sys
	sysStatus.CPUNum = runtime.NumCPU()
	sysStatus.GoVersion = runtime.Version()
	sysStatus.Pid = os.Getpid()
	util.JSONPut(w, sysStatus)
}

func routeMatch(w http.ResponseWriter, r *http.Request) {
	found := false
	for _, p := range route.RoutePath {
		if p.Reg.MatchString(r.URL.Path) {
			found = true
			if err := p.Handler(w, r, p.Reg.FindStringSubmatch(r.URL.Path)); err != nil {
				util.Log.Print(err)
			}
			break
		}
	}
	if !found {
		fallback(w, r)
	}
}

func fallback(w http.ResponseWriter, r *http.Request) {
	const index = "index.html"
	files := []string{index}
	if r.URL.Path != "/" {
		files = []string{r.URL.Path, path.Join(r.URL.Path, index)}
	}
	if !tryFiles(files, w, r) {
		if !tryFiles([]string{index}, w, r) {
			http.NotFound(w, r)
		}
	}
}

func tryFiles(files []string, w http.ResponseWriter, r *http.Request) bool {
	for _, file := range files {
		realpath := filepath.Join("./public", file)
		if f, err := os.Stat(realpath); err == nil {
			if f.Mode().IsRegular() {
				http.ServeFile(w, r, realpath)
				return true
			}
		}
	}
	return false
}

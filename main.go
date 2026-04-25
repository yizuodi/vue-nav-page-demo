package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"mime"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type appConfig struct {
	Server struct {
		Port int    `json:"port"`
		Host string `json:"host"`
	} `json:"server"`
}

func pause() {
	fmt.Println()
	fmt.Println("按回车键退出...")
	_, _ = fmt.Scanln()
}

func fail(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Println("启动失败：" + msg)
	log.Println("启动失败：" + msg)
	pause()
	os.Exit(1)
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	if err := cmd.Start(); err != nil {
		log.Println("自动打开浏览器失败：", err)
	}
}

func readPortFromConfig(configPath string) (host string, port int) {
	host = "127.0.0.1"
	port = 8000
	b, err := os.ReadFile(configPath)
	if err != nil {
		return host, port
	}
	var cfg appConfig
	if err := json.Unmarshal(b, &cfg); err != nil {
		log.Println("读取 config.json 中 server 配置失败，将使用默认端口：", err)
		return host, port
	}
	if cfg.Server.Host != "" {
		host = cfg.Server.Host
	}
	if cfg.Server.Port > 0 && cfg.Server.Port <= 65535 {
		port = cfg.Server.Port
	}
	return host, port
}

func safeJoin(root, urlPath string) (string, bool) {
	urlPath = strings.TrimPrefix(urlPath, "/")
	if urlPath == "" {
		urlPath = "index.html"
	}
	clean := filepath.Clean(urlPath)
	if clean == "." || strings.HasPrefix(clean, "..") || filepath.IsAbs(clean) {
		return "", false
	}
	full := filepath.Join(root, clean)
	rootAbs, _ := filepath.Abs(root)
	fullAbs, _ := filepath.Abs(full)
	if fullAbs != rootAbs && !strings.HasPrefix(fullAbs, rootAbs+string(os.PathSeparator)) {
		return "", false
	}
	return fullAbs, true
}

func serveStatic(root string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Method, r.URL.Path)
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path == "/health" {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			_, _ = w.Write([]byte(`{"ok":true,"app":"nav-server"}`))
			return
		}
		filePath, ok := safeJoin(root, r.URL.Path)
		if !ok {
			http.Error(w, "bad path", http.StatusBadRequest)
			return
		}
		info, err := os.Stat(filePath)
		if err != nil || info.IsDir() {
			http.NotFound(w, r)
			return
		}
		if ext := filepath.Ext(filePath); ext != "" {
			if ct := mime.TypeByExtension(ext); ct != "" {
				w.Header().Set("Content-Type", ct)
			}
		}
		w.Header().Set("Cache-Control", "no-store")
		f, err := os.Open(filePath)
		if err != nil {
			http.Error(w, "open file failed", http.StatusInternalServerError)
			return
		}
		defer f.Close()
		http.ServeContent(w, r, info.Name(), info.ModTime(), f)
	}
}

func main() {
	exe, err := os.Executable()
	if err != nil {
		fmt.Println("无法获取 exe 路径：", err)
		pause()
		os.Exit(1)
	}
	root := filepath.Dir(exe)

	logPath := filepath.Join(root, "nav-server.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err == nil {
		defer logFile.Close()
		log.SetOutput(logFile)
	}
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	_ = mime.AddExtensionType(".js", "text/javascript; charset=utf-8")
	_ = mime.AddExtensionType(".json", "application/json; charset=utf-8")

	indexPath := filepath.Join(root, "index.html")
	configPath := filepath.Join(root, "config.json")
	if _, err := os.Stat(indexPath); err != nil {
		fail("没有找到 index.html。请确认 nav-server.exe 和 index.html 在同一个目录。当前目录：%s，错误：%v", root, err)
	}
	if _, err := os.Stat(configPath); err != nil {
		fail("没有找到 config.json。请确认 nav-server.exe 和 config.json 在同一个目录。当前目录：%s，错误：%v", root, err)
	}

	configHost, configPort := readPortFromConfig(configPath)
	portFlag := flag.Int("port", 0, "监听端口，例如：nav-server.exe -port 18000")
	hostFlag := flag.String("host", "", "监听地址，例如：127.0.0.1 或 0.0.0.0")
	flag.Parse()

	host := configHost
	port := configPort
	if envHost := os.Getenv("NAV_HOST"); envHost != "" {
		host = envHost
	}
	if envPort := os.Getenv("NAV_PORT"); envPort != "" {
		p, err := strconv.Atoi(envPort)
		if err != nil || p <= 0 || p > 65535 {
			fail("环境变量 NAV_PORT 不是有效端口：%s", envPort)
		}
		port = p
	}
	if *hostFlag != "" {
		host = *hostFlag
	}
	if *portFlag != 0 {
		if *portFlag <= 0 || *portFlag > 65535 {
			fail("命令行端口无效：%d", *portFlag)
		}
		port = *portFlag
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		fail("端口 %s 监听失败：%v。可能已有程序占用了该端口。你可以修改 config.json 里的 server.port。", addr, err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", serveStatic(root))

	browserHost := host
	if browserHost == "0.0.0.0" || browserHost == "::" {
		browserHost = "127.0.0.1"
	}
	url := fmt.Sprintf("http://%s:%d/", browserHost, port)
	fmt.Println("站点导航已启动：", url)
	fmt.Println("监听地址：", addr)
	fmt.Println("网站目录：", root)
	fmt.Println("日志文件：", logPath)
	fmt.Println("版本：v4-no-redirect")
	fmt.Println("修改 config.json 后刷新浏览器即可看到导航配置变化。")
	fmt.Println("修改 server.port 后需要重启 exe 才生效。")
	fmt.Println("关闭此窗口即可停止服务。")
	log.Println("server started v4-no-redirect", url, "addr=", addr, "root=", root)

	go func() {
		time.Sleep(800 * time.Millisecond)
		openBrowser(url)
	}()

	if err := http.Serve(ln, mux); err != nil {
		fail("HTTP 服务异常退出：%v", err)
	}
}

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"
)

var version = "dev"

type apod struct {
	Date        string `json:"date"`
	Title       string `json:"title"`
	MediaType   string `json:"media_type"`
	URL         string `json:"url"`
	HDURL       string `json:"hdurl"`
	Explanation string `json:"explanation"`
}

func main() {
	open := flag.Bool("open", false, "open the image in your browser")
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "nasa-apod - fetch NASA's Astronomy Picture of the Day link\n\nUsage:\n  nasa-apod [--open]\n\nFlags:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if *showVersion {
		fmt.Println("nasa-apod", version)
		return
	}

	client := &http.Client{Timeout: 20 * time.Second}
	key := os.Getenv("NASA_API_KEY")
	if key == "" {
		key = "DEMO_KEY"
	}
	req, _ := http.NewRequest(http.MethodGet, "https://api.nasa.gov/planetary/apod?api_key="+key, nil)
	req.Header.Set("User-Agent", "nasa-apod-cli/"+version)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintln(os.Stderr, "request failed:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var e struct {
			Error struct {
				Message string `json:"message"`
			} `json:"error"`
			Msg string `json:"msg"`
		}
		json.NewDecoder(resp.Body).Decode(&e)
		msg := e.Error.Message
		if msg == "" {
			msg = e.Msg
		}
		fmt.Fprintf(os.Stderr, "NASA API error (HTTP %d): %s\n", resp.StatusCode, msg)
		if resp.StatusCode == http.StatusTooManyRequests {
			fmt.Fprintln(os.Stderr, "DEMO_KEY 限流（30 次/小时/IP）。稍后重试，或到 https://api.nasa.gov 免费申请 key 后设置环境变量 NASA_API_KEY。")
		}
		os.Exit(1)
	}

	var pic apod
	if err := json.NewDecoder(resp.Body).Decode(&pic); err != nil {
		fmt.Fprintln(os.Stderr, "bad response:", err)
		os.Exit(1)
	}

	link := pic.URL
	if pic.MediaType == "image" && pic.HDURL != "" {
		link = pic.HDURL
	}

	fmt.Printf("NASA Astronomy Picture of the Day (%s)\n", pic.Date)
	fmt.Printf("Title: %s\n", pic.Title)
	fmt.Printf("Link:  %s\n", link)

	if *open {
		var cmd *exec.Cmd
		switch runtime.GOOS {
		case "darwin":
			cmd = exec.Command("open", link)
		case "windows":
			cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", link)
		default:
			cmd = exec.Command("xdg-open", link)
		}
		if err := cmd.Start(); err != nil {
			fmt.Fprintln(os.Stderr, "could not open browser:", err)
			os.Exit(1)
		}
	}
}
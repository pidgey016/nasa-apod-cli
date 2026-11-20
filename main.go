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

func fetchPic(client *http.Client, key, date string) (*apod, int, error) {
	req, _ := http.NewRequest(http.MethodGet, "https://api.nasa.gov/planetary/apod?api_key="+key+"&date="+date, nil)
	req.Header.Set("User-Agent", "nasa-apod-cli/"+version)
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
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
		return nil, resp.StatusCode, fmt.Errorf("NASA API error (HTTP %d): %s", resp.StatusCode, msg)
	}
	var pic apod
	if err := json.NewDecoder(resp.Body).Decode(&pic); err != nil {
		return nil, resp.StatusCode, fmt.Errorf("bad response: %w", err)
	}
	return &pic, resp.StatusCode, nil
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

	// 无 date 参数的接口间歇性 500，显式日期稳定；当天未发布时回退昨天一期
	now := time.Now().UTC()
	dates := []string{now.Format("2006-01-02"), now.AddDate(0, 0, -1).Format("2006-01-02")}
	var pic *apod
	var lastErr error
	var lastCode int
	for _, d := range dates {
		for attempt := 0; attempt < 2; attempt++ {
			p, code, err := fetchPic(client, key, d)
			if err == nil {
				pic = p
				break
			}
			lastErr, lastCode = err, code
			if code != http.StatusInternalServerError {
				break
			}
			time.Sleep(time.Second)
		}
		if pic != nil || lastCode == http.StatusTooManyRequests {
			break
		}
	}
	if pic == nil {
		fmt.Fprintln(os.Stderr, lastErr)
		if lastCode == http.StatusTooManyRequests {
			fmt.Fprintln(os.Stderr, "DEMO_KEY 限流（30 次/小时/IP）。稍后重试，或到 https://api.nasa.gov 免费申请 key 后设置环境变量 NASA_API_KEY。")
		}
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
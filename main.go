package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"
	_ "time/tzdata"
)

var version = "dev"

type apod struct {
	Date        string `json:"date"`
	Title       string `json:"title"`
	Explanation string `json:"explanation"`
	MediaType   string `json:"media_type"`
	URL         string `json:"url"`
	HDURL       string `json:"hdurl"`
	Alt         string `json:"alt"`
	Permalink   string `json:"permalink"`
}

func fetchPic(client *http.Client, date string) (*apod, int, error) {
	d, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, 0, err
	}
	// APOD 已于 2026-11 从 apod.nasa.gov 迁移到 science.nasa.gov，旧 api.nasa.gov 接口只返回占位内容
	url := "https://science.nasa.gov/wp-json/wp/v2/apod-basic/" + d.Format("060102") + "?_fields=date,title,explanation,media_type,url,hdurl,alt,permalink"
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("User-Agent", "nasa-apod-cli/"+version)
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		var e struct {
			Message string `json:"message"`
		}
		json.NewDecoder(resp.Body).Decode(&e)
		return nil, resp.StatusCode, fmt.Errorf("NASA API error (HTTP %d): %s", resp.StatusCode, e.Message)
	}
	var pic apod
	if err := json.NewDecoder(resp.Body).Decode(&pic); err != nil {
		return nil, resp.StatusCode, fmt.Errorf("bad response: %w", err)
	}
	return &pic, resp.StatusCode, nil
}

func main() {
	open := flag.Bool("open", false, "open the image in your browser")
	dateFlag := flag.String("date", "", "fetch a specific date (YYYY-MM-DD)")
	jsonOut := flag.Bool("json", false, "output raw JSON (date, title, explanation, media_type, url, hdurl, alt, permalink)")
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "nasa-apod - fetch NASA's Astronomy Picture of the Day link\n\nUsage:\n  nasa-apod [--open] [--date YYYY-MM-DD] [--json]\n\nFlags:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if *showVersion {
		fmt.Println("nasa-apod", version)
		return
	}

	client := &http.Client{Timeout: 20 * time.Second}

	// APOD 按美东时间发布，"当天"以该时区为准
	et, _ := time.LoadLocation("America/New_York")
	date := time.Now().In(et).Format("2006-01-02")
	if *dateFlag != "" {
		if _, err := time.Parse("2006-01-02", *dateFlag); err != nil {
			fmt.Fprintln(os.Stderr, "invalid --date (want YYYY-MM-DD):", *dateFlag)
			os.Exit(1)
		}
		date = *dateFlag
	}
	pic, _, err := fetchPic(client, date)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	link := pic.URL
	if pic.MediaType == "image" && pic.HDURL != "" {
		link = pic.HDURL
	}

	if *jsonOut {
		var buf bytes.Buffer
		enc := json.NewEncoder(&buf)
		enc.SetEscapeHTML(false)
		if err := enc.Encode(pic); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Print(buf.String())
	} else {
		fmt.Printf("NASA Astronomy Picture of the Day (%s)\n", pic.Date)
		fmt.Printf("Title: %s\n", pic.Title)
		fmt.Printf("Link:  %s\n", link)
	}

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
package main

import (
	"io"
	"log"
	"os"
	"os/exec"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Torrent struct {
	Path   string `json:"path"`
	Magnet string `json:"magnet"`
}

var reqs chan Torrent

func main() {
	e := echo.New()

	// Middleware
	e.Use(middleware.CORS())

	// serve static files
	e.Static("/", "public")
	e.Static("/stream", "stream")
	// set up buffered channel
	reqs = make(chan Torrent, 1000)

	// create log file
	_, err := os.Create("log.log")
	if err != nil {
		log.Fatal(err)
	}

	// start monitoring requests
	go monitorRequests()

	// api routes
	e.POST("/api/v1/torrents", func(c echo.Context) error {
		log.Println("POST /api/v1/torrents")

		// get the body
		body, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return err
		}

		// Get Path
		Path := c.QueryParam("path")

		// add request to queue
		reqs <- Torrent{Path: Path, Magnet: string(body)}

		return c.String(200, "POST /api/v1/torrents")
	})

	e.GET("/api/v1/torrents/stream", func(c echo.Context) error {
		// read file
		f, err := os.Open("log.log")
		if err != nil {
			return err
		}
		defer f.Close()

		// stream file
		if _, err := io.Copy(c.Response().Writer, f); err != nil {
			return err
		}
		return nil
	})

	// start server
	e.Logger.Fatal(e.Start(":4040"))
}

func downloadTorrent(Path string, magnet string) {
	log.Println("Downloading torrent: "+magnet, " to path: "+Path)

	// download torrent
	// cmd := exec.Command("cd", Path, "&&", "torrent", "download", magnet)
	cmd := exec.Command("torrent", "download", magnet)
	cmd.Dir = Path

	// print output to file
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	// get output line by line
	go func() {
		for {
			buf := make([]byte, 1024)
			n, err := stdout.Read(buf)
			if err != nil {
				if err == io.EOF {
					break
				}
				log.Fatal(err)
			}
			saveToFile(string(buf[:n]))
		}
	}()

	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}

	// get output line by line
	go func() {
		for {
			buf := make([]byte, 1024)
			n, err := stderr.Read(buf)
			if err != nil {
				if err == io.EOF {
					break
				}
				log.Fatal(err)
			}
			saveToFile(string(buf[:n]))
		}
	}()

	// run command
	cmd.Run()
}

func monitorRequests() {
	for {
		req := <-reqs
		downloadTorrent(req.Path, req.Magnet)
	}
}

func saveToFile(logLine string) {
	f, err := os.OpenFile("log.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	// write to file
	if _, err := f.WriteString(logLine); err != nil {
		log.Fatal(err)
	}
}

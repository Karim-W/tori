package main

import (
	"io"
	"log"
	"os"
	"os/exec"

	"github.com/labstack/echo/v4"
)

type Torrent struct {
	Path   string `json:"path"`
	Magnet string `json:"magnet"`
}

var reqs chan Torrent

func main() {
	e := echo.New()

	// set up buffered channel
	reqs = make(chan Torrent, 1000)

	// serve static files
	e.Static("/", "public")

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

	// start server
	e.Logger.Fatal(e.Start(":4040"))
}

func downloadTorrent(Path string, magnet string) {
	log.Println("Downloading torrent: "+magnet, " to path: "+Path)

	// download torrent
	// cmd := exec.Command("cd", Path, "&&", "torrent", "download", magnet)
	cmd := exec.Command("torrent", "download", magnet)
	cmd.Dir = Path

	// print output to console
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// run command
	cmd.Run()
}

func monitorRequests() {
	for {
		req := <-reqs
		downloadTorrent(req.Path, req.Magnet)
	}
}

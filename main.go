package main

import (
	"io"
	"log"
	"os"
	"os/exec"

	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()

	// serve static files
	e.Static("/", "public")

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

		go downloadTorrent(Path, string(body))

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

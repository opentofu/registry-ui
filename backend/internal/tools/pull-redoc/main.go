package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func main() {
	version := ""
	flag.StringVar(&version, "version", "latest", "The version of ReDoc to download.")
	flag.Parse()

	// Pull Redoc script from Redocly CDN

	redocCDNURL := fmt.Sprintf("https://cdn.redoc.ly/redoc/%s/bundles/redoc.standalone.js", version)

	redocFileName := "redoc.standalone.js"
	redocFile, err := os.Create(fmt.Sprintf("server/%s", redocFileName))
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}
	defer redocFile.Close()

	resp, err := http.Get(redocCDNURL)
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		err = fmt.Errorf("err downloading ReDoc script file: %+v\n", resp)
		log.Print(err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if _, err = io.Copy(redocFile, resp.Body); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}

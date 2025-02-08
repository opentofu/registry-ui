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
	redocFile, err := os.Create("server/redoc.tmp")
	if err != nil {
		log.Fatalf("err creating ReDoc script tmp file: %v", err)
	}
	defer redocFile.Close()

	resp, err := http.Get(redocCDNURL)
	if err != nil {
		log.Fatalf("err downloading ReDoc script file: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		err = fmt.Errorf("err downloading ReDoc script file: %+v\n", resp)
		log.Fatal(err)
	}

	if _, err = io.Copy(redocFile, resp.Body); err != nil {
		log.Fatal(err)
	}

	err = os.Rename("server/redoc.tmp", fmt.Sprintf("server/%s", redocFileName))
	if err != nil {
		log.Fatalf("Error renaming ReDoc script file: %v", err)
	}

	// Pull ReDoc LICENSE file
	licenseFileURL := "https://raw.githubusercontent.com/Redocly/redoc/refs/heads/main/LICENSE"

	licenseFile, err := os.Create("server/license.tmp")
	if err != nil {
		log.Fatalf("err creating ReDoc license tmp file: %v", err)
	}
	resp1, err := http.Get(licenseFileURL)
	if err != nil {
		log.Fatalf("err downloading ReDoc license file: %v", err)
	}

	defer resp1.Body.Close()
	if resp1.StatusCode < 200 || resp1.StatusCode > 299 {
		err = fmt.Errorf("err downloading ReDoc license file: %+v\n", resp1)
		log.Fatal(err)
	}

	if _, err = io.Copy(licenseFile, resp1.Body); err != nil {
		log.Fatal(err)
	}

	err = os.Rename("server/license.tmp", "server/redoc.standalone.js.LICENSE.txt")
	if err != nil {
		log.Fatalf("Error renaming ReDoc license file: %v", err)
	}
}

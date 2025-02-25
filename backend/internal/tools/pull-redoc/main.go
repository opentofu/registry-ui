package main

import (
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func main() {
	version := ""
	checksum := ""
	flag.StringVar(&version, "version", "latest", "The version of ReDoc to download.")
	flag.StringVar(&checksum, "checksum", "", "The checksum of the ReDoc script file corresponding to the version  to be downloaded.")
	flag.Parse()

	fmt.Printf("checksum: %+v\n", checksum)

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

	// Copy downloaded contents to a local file
	if _, err = io.Copy(redocFile, resp.Body); err != nil {
		log.Fatal(err)
	}

	// Reset file pointer to the beginning before reading it into the hash
	if _, err = redocFile.Seek(0, 0); err != nil {
		log.Fatal(err)
	}
	hash := sha256.New()
	_, err = io.Copy(hash, redocFile)
	if err != nil {
		log.Fatalf("err hashing script file: %+v\n", err)
	}

	// compare hashed value with supplied checksum
	if checksum != hex.EncodeToString(hash.Sum(nil)) {
		log.Fatalf("could not verify checksum for the file : %+v\n", redocFileName)
	}

	err = os.Rename("server/redoc.tmp", fmt.Sprintf("server/%s", redocFileName))
	if err != nil {
		log.Fatalf("Error renaming ReDoc script file: %v", err)
	}

	// Pull ReDoc LICENSE file
	licenseFileURL := fmt.Sprintf("https://cdn.redoc.ly/redoc/%s/bundles/redoc.standalone.js.LICENSE.txt", version)

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

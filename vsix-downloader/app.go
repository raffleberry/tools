package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	target := flag.String("url", "", "marketplace url")
	ver := flag.String("ver", "latest", "version")

	flag.Parse()
	if *target == "" {
		flag.Usage()
		return
	}
	url, err := url.Parse(*target)
	if err != nil {
		log.Fatalf("Bad Url: %s err: %v", *target, err)
	}

	itemName := url.Query().Get("itemName")
	if itemName == "" {
		log.Fatal("invalid url format, no itemName found")
	}

	pub, name, ok := strings.Cut(itemName, ".")
	if !ok {
		log.Fatalf("invalid url format: got itemName:[%s] expected:['<publisher>.<name>']", itemName)
	}

	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("failed to get working directory: %v", err)
	}

	outUrl := fmt.Sprintf("https://%s.gallery.vsassets.io/_apis/public/gallery/publisher/%s/extension/%s/%s/assetbyname/Microsoft.VisualStudio.Services.VSIXPackage", pub, pub, name, *ver)

	outName := fmt.Sprintf("%s.vsix", itemName)
	outPath := filepath.Join(wd, outName)
	dst, err := os.Create(outPath)
	if err != nil {
		log.Fatalf("failed to create file: %v", err)
	}
	defer dst.Close()

	log.Printf("Downloading: %s\n", outName)

	resp, err := http.Get(outUrl)
	if err != nil {
		log.Fatalf("failed to make HTTP request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("bad status code: %s", resp.Status)
	}

	_, err = io.Copy(dst, resp.Body)
	if err != nil {
		log.Fatalf("failed to save file contents: %v", err)
	}

	log.Printf("Saved to: %s\n", outPath)
}

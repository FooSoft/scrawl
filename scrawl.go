package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

func scrape(url, css, attr string) ([]string, error) {
	doc, err := goquery.NewDocument(url)
	if err != nil {
		return nil, err
	}

	var assets []string
	doc.Find(css).Each(func(index int, sel *goquery.Selection) {
		asset := sel.Text()
		if len(attr) > 0 {
			asset, _ = sel.Attr(attr)
		}

		asset = strings.TrimSpace(asset)
		if len(asset) > 0 {
			assets = append(assets, asset)
		}
	})

	return assets, nil
}

func download(url string, w io.Writer) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	io.Copy(w, resp.Body)
	return nil
}

func export(path string, r io.Reader) error {
	out, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	io.Copy(out, r)
	return nil
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [options] url selector [path]\n", path.Base(os.Args[0]))
	fmt.Fprintf(os.Stderr, "https://foosoft.net/projects/scrawl/\n\n")
	fmt.Fprintf(os.Stderr, "Parameters:\n")
	flag.PrintDefaults()
}

func main() {
	var (
		attr    = flag.String("attr", "", "attribute to query")
		dir     = flag.String("dir", ".", "output directory")
		verbose = flag.Bool("verbose", false, "verbose output")
	)

	flag.Usage = usage
	flag.Parse()

	if flag.NArg() != 2 {
		flag.Usage()
		os.Exit(2)
	}

	var (
		baseRaw = flag.Arg(0)
		css     = flag.Arg(1)
	)

	base, err := url.Parse(baseRaw)
	if err != nil {
		log.Fatal(err)
	}

	if *verbose {
		log.Printf("scraping page '%s'", baseRaw)
	}

	assetsRaw, err := scrape(baseRaw, css, *attr)
	if err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup
	for _, assetRaw := range assetsRaw {
		wg.Add(1)
		go func(assetRaw string) {
			defer wg.Done()

			if *verbose {
				log.Printf("parsing asset string '%s'", assetRaw)
			}

			asset, err := url.Parse(assetRaw)
			if err != nil {
				log.Fatal(err)
			}

			if !asset.IsAbs() {
				asset = asset.ResolveReference(base)
			}

			if *verbose {
				log.Printf("downloading file '%s'", asset.String())
			}

			var buff bytes.Buffer
			if err := download(asset.String(), &buff); err != nil {
				log.Fatal(err)
			}

			path := filepath.Join(*dir, filepath.Base(asset.Path))

			if *verbose {
				log.Printf("writing file '%s'", path)
			}

			if err := export(path, &buff); err != nil {
				log.Fatal(err)
			}
		}(assetRaw)
	}

	wg.Wait()
}

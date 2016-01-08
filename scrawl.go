/*
 * Copyright (c) 2016 Alex Yatskov <alex@foosoft.net>
 * Author: Alex Yatskov <alex@foosoft.net>
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy of
 * this software and associated documentation files (the "Software"), to deal in
 * the Software without restriction, including without limitation the rights to
 * use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
 * the Software, and to permit persons to whom the Software is furnished to do so,
 * subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
 * FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
 * COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
 * IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
 * CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

package main

import (
	"bytes"
	"flag"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func scrape(url, css, attr string) (string, error) {
	doc, err := goquery.NewDocument(url)
	if err != nil {
		return "", err
	}

	sel := doc.Find(css)

	var res string
	if len(attr) == 0 {
		res = sel.Text()
	} else {
		res, _ = sel.Attr(attr)
	}

	return strings.TrimSpace(res), nil
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

func main() {
	if flag.NArg() < 2 {
		flag.Usage()
		os.Exit(2)
	}

	var (
		attr    = flag.String("attr", "", "attrerty to query")
		rawBase = flag.Arg(1)
		css     = flag.Arg(2)
	)

	base, err := url.Parse(rawBase)
	if err != nil {
		log.Fatal(err)
	}

	resRaw, err := scrape(rawBase, css, *attr)
	if err != nil {
		log.Fatal(err)
	}

	log.Print(resRaw)

	res, err := url.Parse(resRaw)
	if err != nil {
		log.Fatal(err)
	}

	if !res.IsAbs() {
		res = res.ResolveReference(base)
	}

	var buff bytes.Buffer
	if err := download(res.String(), &buff); err != nil {
		log.Fatal(err)
	}

	var path string
	if flag.NArg() > 2 {
		path = flag.Arg(3)
	} else {
		path = filepath.Base(res.Path)
	}

	if err := export(path, &buff); err != nil {
		log.Fatal(err)
	}
}

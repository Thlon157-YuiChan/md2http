// md2http is a simple HTTP server that renders Markdown files as HTML.
//
// Copyright (C) 2026 yuthlon157@gmail.com
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
package main

import (
	"flag"
	"fmt"
	"html"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

var md = goldmark.New(
	goldmark.WithExtensions(extension.GFM),
)

func main() {
	port := flag.Int("port", 8080, "listen port")
	dir := flag.String("dir", ".", "directory to serve")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [-port PORT] [-dir DIR] [PORT]\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	p := *port
	if flag.NArg() > 0 {
		if n, err := strconv.Atoi(flag.Arg(0)); err == nil {
			p = n
		}
	}
	if v := os.Getenv("PORT"); v != "" && *port == 8080 {
		// allow env override when default port unchanged
		if n, err := strconv.Atoi(v); err == nil {
			p = n
		}
	}

	root, err := filepath.Abs(*dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		serve(w, r, root)
	})
	addr := fmt.Sprintf(":%d", p)
	fmt.Printf("md2http: serving %s on http://localhost%s\n", root, addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func serve(w http.ResponseWriter, r *http.Request, root string) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	rel := strings.TrimPrefix(r.URL.Path, "/")
	if rel == "" {
		rel = "."
	}
	clean := filepath.Clean(rel)
	if clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	full := filepath.Join(root, clean)
	info, err := os.Stat(full)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if info.IsDir() {
		serveDir(w, r, full, clean)
		return
	}
	ext := strings.ToLower(filepath.Ext(full))
	if ext == ".md" || ext == ".markdown" {
		serveMarkdown(w, full)
		return
	}
	http.ServeFile(w, r, full)
}

func serveDir(w http.ResponseWriter, r *http.Request, full, clean string) {
	entries, err := os.ReadDir(full)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	sort.Strings(names)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, "<!DOCTYPE html><html><head><meta charset=utf-8><title>%s</title></head><body><h1>%s</h1><ul>",
		html.EscapeString(clean), html.EscapeString(clean))
	for _, n := range names {
		href := "/" + filepath.ToSlash(filepath.Join(clean, n))
		if clean == "." {
			href = "/" + n
		}
		href = strings.ReplaceAll(href, "//", "/")
		fmt.Fprintf(w, "<li><a href=%q>%s</a></li>", href, html.EscapeString(n))
	}
	fmt.Fprint(w, "</ul></body></html>")
}

func serveMarkdown(w http.ResponseWriter, full string) {
	data, err := os.ReadFile(full)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var buf strings.Builder
	if err := md.Convert(data, &buf); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	title := filepath.Base(full)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<!DOCTYPE html><html><head><meta charset="utf-8"><title>%s</title>
<style>body{max-width:800px;margin:2em auto;padding:0 1em;font-family:sans-serif;line-height:1.6}pre{background:#f4f4f4;padding:1em;overflow:auto}code{background:#f4f4f4;padding:.2em .4em;border-radius:3px}img{max-width:100%%}</style>
</head><body>%s</body></html>`, html.EscapeString(title), buf.String())
}

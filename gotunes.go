package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"text/template"
)

type Directory struct { // TODO should be lowercase because it doesn't have to be exported
	Name string // TODO should be lowercase
}

type File struct { //  TODO should be lowercase
	Name                 string
	NameWithoutExtension string
}

type Playlist struct { // TODO should be lowercase
	Directories []Directory
	Files       []File
}

func loggingHandler(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		log.Println(req.Method, req.RemoteAddr, req.URL.Path)
		handler.ServeHTTP(res, req)
	})
}

func interceptor(res http.ResponseWriter, req *http.Request) {
	// TODO do a containsDotDot check first: https://cs.opensource.google/go/go/+/master:src/net/http/fs.go;drc=2b794ed86cb1b718bc212ee90fecbb8f3b28a744;l=853?q=servefile&ss=go%2Fgo and return an error
	if strings.HasSuffix(req.URL.Path, "/index.m3u") {
		tmpl, err := template.New("index.m3u").ParseFiles("index.m3u") // TODO template can be parsed in main and passed to interceptor as parameter or use the receiver pattern see apiHandler in example of: https://pkg.go.dev/net/http#ServeMux.Handle
		if err != nil {
			log.Fatal(err)
		}
		dirEntries, err := os.ReadDir("/var/lib/gotunes") // TODO concatenate actual path from req
		if err != nil {
			log.Fatal(err)
		}
		directories := slices.Collect(func(yield func(Directory) bool) { // TODO implement inner function as generic filter and map | https://pkg.go.dev/slices#Values could be useful
			for _, dirEntry := range dirEntries {
				if dirEntry.IsDir() {
					directory := Directory{
						Name: dirEntry.Name(),
					}
					if !yield(directory) {
						return
					}
				}
			}
		})
		files := slices.Collect(func(yield func(File) bool) { // TODO implement inner function as generic filter and map | https://pkg.go.dev/slices#Values could be useful
			for _, dirEntry := range dirEntries {
				if !dirEntry.IsDir() { // TODO filter .mp3 files as well
					file := File{
						Name:                 dirEntry.Name(),
						NameWithoutExtension: strings.TrimSuffix(dirEntry.Name(), filepath.Ext(dirEntry.Name())),
					}
					if !yield(file) {
						return
					}
				}
			}
		})
		playlist := Playlist{
			Directories: directories,
			Files:       files,
		}
		err = tmpl.Execute(res, playlist)
		if err != nil {
			log.Fatal(err)
		}
	} else if strings.HasSuffix(req.URL.Path, ".m3u8") {
		// TODO concatenate actual path from req
		// TODO run ffmpeg command to convert mp3 to hls into /tmp/gotunes/<path>/<name>.m3u8 | use os.TempDir() to find the actual dir
		// ffmpeg -i input.mp3 -c:a copy -f hls -hls_time 10 -hls_list_size 0 -hls_segment_filename output.%03d.ts output.m3u8
		http.FileServer(http.Dir("/tmp/gotunes")).ServeHTTP(res, req) // TODO http.FileServer(http.Dir("/tmp/gotunes")) can be initialized in main and passed to interceptor as parameter or use the receiver pattern see apiHandler in example of https://pkg.go.dev/net/http#ServeMux.Handle | use os.TempDir() to find the actual dir
	} else if strings.HasSuffix(req.URL.Path, ".ts") {
		// TODO same as above but no ffmpeg
		http.FileServer(http.Dir("/tmp/gotunes")).ServeHTTP(res, req)
	} else {
		http.NotFound(res, req)
	}
}

func main() {
	mux := http.NewServeMux()
	mux.Handle("GET /", loggingHandler(http.HandlerFunc(interceptor)))
	log.Fatal(http.ListenAndServe(":8080", mux))
}

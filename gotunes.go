package main

import (
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"text/template"

	"github.com/bernhardfritz/gotunes/itertools"
)

type Directory struct {
	Name string
}

type File struct {
	Name string
}

type Playlist struct {
	Path        string
	Directories []Directory
	Files       []File
}

func loggingHandler(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		log.Println(req.Method, req.RemoteAddr, req.URL.Path)
		handler.ServeHTTP(res, req)
	})
}

func containsDotDot(v string) bool {
	if !strings.Contains(v, "..") {
		return false
	}
	for _, ent := range strings.FieldsFunc(v, isSlashRune) {
		if ent == ".." {
			return true
		}
	}
	return false
}

func isSlashRune(r rune) bool {
	return r == '/' || r == '\\'
}

func isDirectory(dirEntry os.DirEntry) bool {
	return dirEntry.IsDir()
}

func toDirectory(dirEntry os.DirEntry) Directory {
	return Directory{
		Name: dirEntry.Name(),
	}
}

func isFile(dirEntry os.DirEntry) bool {
	return !dirEntry.IsDir()
}

func toFile(dirEntry os.DirEntry) File {
	return File{
		Name: dirEntry.Name(),
	}
}

func interceptor(res http.ResponseWriter, req *http.Request) {
	if containsDotDot(req.URL.Path) {
		http.Error(res, "invalid URL path", http.StatusBadRequest)
		return
	}
	// TODO implement redirect from / to /index.m3u => see https://cs.opensource.google/go/go/+/master:src/net/http/fs.go;l=674?q=FileServer&ss=go%2Fgo
	// also applies to /MasterOfPuppets to /MasterOfPuppets/index.m3u
	if strings.HasSuffix(req.URL.Path, "/index.m3u") {
		// TODO refactor-extract into dedicated function e.g. servePlaylist
		funcMap := template.FuncMap{
			"PathJoin": path.Join,
		}
		tmpl, err := template.New("index.m3u").Funcs(funcMap).ParseFiles("index.m3u") // TODO template can be parsed in main and passed to interceptor as parameter or use the receiver pattern see apiHandler in example of: https://pkg.go.dev/net/http#ServeMux.Handle
		if err != nil {
			log.Fatal(err)
		}
		dirEntries, err := os.ReadDir(filepath.Join("/var/lib/gotunes", filepath.Dir(req.URL.Path)))
		if err != nil {
			log.Fatal(err)
		}
		playlist := Playlist{
			Path:        path.Dir(req.URL.Path),
			Directories: slices.Collect(itertools.Map(toDirectory, itertools.Filter(isDirectory, slices.Values(dirEntries)))),
			Files:       slices.Collect(itertools.Map(toFile, itertools.Filter(isFile, slices.Values(dirEntries)))),
		}
		err = tmpl.Execute(res, playlist)
		if err != nil {
			log.Fatal(err)
		}
	} else if strings.HasSuffix(req.URL.Path, ".m3u8") {
		// TODO refactor-extract into dedicated function e.g. serveStream
		relativePath := strings.TrimSuffix(req.URL.Path, ".m3u8")
		input := filepath.Join("/var/lib/gotunes", relativePath)
		tmpDir := filepath.Join(os.TempDir(), "gotunes")
		output := filepath.Join(tmpDir, relativePath)
		err := os.MkdirAll(filepath.Dir(output), 0660)
		if err != nil {
			log.Fatal(err)
		}
		ffmpeg := exec.Command("ffmpeg", "-i", input, "-c:a", "copy", "-f", "hls", "-hls_time", "10", "-hls_list_size", "0", "-hls_segment_filename", output+".%03d.ts", output+".m3u8")
		log.Println(strings.Join(ffmpeg.Args, " "))
		out, err := ffmpeg.CombinedOutput()
		if err != nil {
			log.Fatal(err)
		}
		log.Println(string(out))
		http.FileServer(http.Dir(tmpDir)).ServeHTTP(res, req) // TODO http.FileServer(http.Dir("/tmp/gotunes")) can be initialized in main and passed to interceptor as parameter or use the receiver pattern see apiHandler in example of https://pkg.go.dev/net/http#ServeMux.Handle | use os.TempDir() to find the actual dir
	} else if strings.HasSuffix(req.URL.Path, ".ts") {
		// TODO refactor-extract into dedicated function e.g. serveSegment
		tmpDir := filepath.Join(os.TempDir(), "gotunes")
		http.FileServer(http.Dir(tmpDir)).ServeHTTP(res, req)
	} else {
		http.NotFound(res, req)
	}
}

func main() {
	http.Handle("GET /", loggingHandler(http.HandlerFunc(interceptor)))
	log.Fatal(http.ListenAndServe(":8080", nil))
}

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

	"github.com/bernhardfritz/mytunes/itertools"
)

type Playlist struct {
	Path        string
	Directories []string
	Files       []string
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

// localRedirect gives a Moved Permanently response.
// It does not convert relative paths to absolute paths like Redirect does.
func localRedirect(w http.ResponseWriter, r *http.Request, newPath string) {
	if q := r.URL.RawQuery; q != "" {
		newPath += "?" + q
	}
	w.Header().Set("Location", newPath)
	w.WriteHeader(http.StatusMovedPermanently)
}

func interceptor(res http.ResponseWriter, req *http.Request) {
	if containsDotDot(req.URL.Path) {
		http.Error(res, "invalid URL path", http.StatusBadRequest)
		return
	}

	if req.URL.Path[len(req.URL.Path)-1] == '/' {
		req.URL.Path += "index.m3u"
	}

	if strings.HasSuffix(req.URL.Path, "/index.m3u") {
		// TODO refactor-extract into dedicated function e.g. serveWildcardSlashIndexDotM3u
		funcMap := template.FuncMap{
			"PathJoin": path.Join,
		}
		tmpl, err := template.New("index.m3u").Funcs(funcMap).ParseFiles("index.m3u") // TODO template can be parsed in main and passed to interceptor as parameter or use the receiver pattern see apiHandler in example of: https://pkg.go.dev/net/http#ServeMux.Handle
		if err != nil {
			log.Fatal(err)
		}
		dirEntries, err := os.ReadDir(filepath.Join("/var/lib/mytunes", filepath.Dir(req.URL.Path))) // TODO prefer http.Dir("/var/lib/mytunes").Open(filepath.Dir(req.URL.Path)).Readdir(-1) instead
		if err != nil {
			log.Fatal(err)
		}
		playlist := Playlist{
			Path:        path.Dir(req.URL.Path),
			Directories: slices.Collect(itertools.Map(os.DirEntry.Name, itertools.Filter(os.DirEntry.IsDir, slices.Values(dirEntries)))),
			Files:       slices.Collect(itertools.Filter(itertools.HasSuffix(".mp3"), itertools.Map(os.DirEntry.Name, itertools.Filter(itertools.Not(os.DirEntry.IsDir), slices.Values(dirEntries))))),
		}
		err = tmpl.Execute(res, playlist)
		if err != nil {
			log.Fatal(err)
		}
	} else if strings.HasSuffix(req.URL.Path, ".m3u8") {
		// TODO refactor-extract into dedicated function e.g. serveWildcardDotM3u8
		relativePath := strings.TrimSuffix(req.URL.Path, ".m3u8")
		// TODO validate that relativePath ends with .mp3
		// TODO validate that file actually exists
		input := filepath.Join("/var/lib/mytunes", relativePath)
		tmpDir := filepath.Join(os.TempDir(), "mytunes")
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
		http.FileServer(http.Dir(tmpDir)).ServeHTTP(res, req) // TODO http.FileServer(http.Dir("/tmp/mytunes")) can be initialized in main and passed to interceptor as parameter or use the receiver pattern see apiHandler in example of https://pkg.go.dev/net/http#ServeMux.Handle | use os.TempDir() to find the actual dir
	} else if strings.HasSuffix(req.URL.Path, ".ts") {
		// TODO refactor-extract into dedicated function e.g. serveWildcardDotTs
		tmpDir := filepath.Join(os.TempDir(), "mytunes")
		http.FileServer(http.Dir(tmpDir)).ServeHTTP(res, req)
	} else {
		fs := http.Dir("/var/lib/mytunes")
		f, err := fs.Open(req.URL.Path)
		if err != nil {
			http.NotFound(res, req)
			return
		}
		defer f.Close()

		d, err := f.Stat()
		if err != nil {
			http.NotFound(res, req)
			return
		}

		if d.IsDir() {
			url := req.URL.Path
			// redirect if the directory name doesn't end in a slash
			if url == "" || url[len(url)-1] != '/' {
				localRedirect(res, req, path.Base(url)+"/")
				return
			}
		}

		http.NotFound(res, req)
	}
}

func main() {
	http.Handle("GET /", loggingHandler(http.HandlerFunc(interceptor)))
	log.Fatal(http.ListenAndServe(":8080", nil))
}

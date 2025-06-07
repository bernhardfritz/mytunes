package main

import (
	_ "embed"
	"io/fs"
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

type Context struct {
	indexM3u *template.Template
	musicDir string
}

func NewContext(indexM3u *template.Template, musicDir string) *Context {
	return &Context{
		indexM3u: indexM3u,
		musicDir: musicDir,
	}
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

func (ctx *Context) handleRoot(res http.ResponseWriter, req *http.Request) {
	if containsDotDot(req.URL.Path) {
		http.Error(res, "invalid URL path", http.StatusBadRequest)
		return
	}

	if req.URL.Path[len(req.URL.Path)-1] == '/' {
		req.URL.Path += "index.m3u"
	}

	if strings.HasSuffix(req.URL.Path, "/index.m3u") {
		dir := http.Dir(ctx.musicDir)
		f, err := dir.Open(filepath.Dir(req.URL.Path))
		if err != nil {
			log.Println(err)
			http.NotFound(res, req)
			return
		}
		defer f.Close()
		fileInfos, err := f.Readdir(-1)
		if err != nil {
			log.Println(err)
			http.NotFound(res, req)
			return
		}
		playlist := Playlist{
			Path:        path.Dir(req.URL.Path),
			Directories: slices.Collect(itertools.Map(fs.FileInfo.Name, itertools.Filter(fs.FileInfo.IsDir, slices.Values(fileInfos)))),
			Files:       slices.Collect(itertools.Filter(itertools.HasSuffix(".mp3"), itertools.Map(fs.FileInfo.Name, itertools.Filter(itertools.Not(fs.FileInfo.IsDir), slices.Values(fileInfos))))),
		}
		err = ctx.indexM3u.Execute(res, playlist)
		if err != nil {
			log.Println(err)
			http.Error(res, "Internal server error", http.StatusInternalServerError)
		}
	} else if strings.HasSuffix(req.URL.Path, ".m3u8") {
		relativePath := strings.TrimSuffix(req.URL.Path, ".m3u8")
		if !strings.HasSuffix(relativePath, ".mp3") {
			http.NotFound(res, req)
			return
		}
		dir := http.Dir(ctx.musicDir)
		f, err := dir.Open(relativePath)
		if err != nil {
			log.Println(err)
			http.NotFound(res, req)
			return
		}
		defer f.Close()
		fileInfo, err := f.Stat() // check if file exists
		if err != nil {
			log.Println(err)
			http.NotFound(res, req)
			return
		}
		if fileInfo.IsDir() {
			http.NotFound(res, req)
			return
		}
		input := filepath.Join(ctx.musicDir, relativePath)
		tmpDir := filepath.Join(os.TempDir(), "mytunes")
		output := filepath.Join(tmpDir, relativePath)
		err = os.MkdirAll(filepath.Dir(output), 0660)
		if err != nil {
			log.Println(err)
			http.Error(res, "Internal server error", http.StatusInternalServerError)
			return
		}
		ffmpeg := exec.Command("ffmpeg", "-i", input, "-c:a", "copy", "-f", "hls", "-hls_time", "10", "-hls_list_size", "0", "-hls_segment_filename", output+".%03d.ts", output+".m3u8")
		log.Println(strings.Join(ffmpeg.Args, " "))
		out, err := ffmpeg.CombinedOutput()
		if err != nil {
			log.Println(err)
			http.Error(res, "Internal server error", http.StatusInternalServerError)
			return
		}
		log.Println(string(out))
		http.FileServer(http.Dir(tmpDir)).ServeHTTP(res, req)
	} else if strings.HasSuffix(req.URL.Path, ".ts") {
		tmpDir := filepath.Join(os.TempDir(), "mytunes")
		http.FileServer(http.Dir(tmpDir)).ServeHTTP(res, req)
	} else {
		dir := http.Dir(ctx.musicDir)
		f, err := dir.Open(req.URL.Path)
		if err != nil {
			log.Println(err)
			http.NotFound(res, req)
			return
		}
		defer f.Close()

		fileInfo, err := f.Stat() // check if directory exists
		if err != nil {
			log.Println(err)
			http.NotFound(res, req)
			return
		}

		if fileInfo.IsDir() {
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

//go:embed index.m3u
var indexM3u string

func main() {
	if len(os.Args) > 2 {
		log.Fatal("Usage: mytunes [PATH]")
	}
	var musicDir string
	if len(os.Args) == 2 {
		musicDir = os.Args[1]
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatal(err)
		}
		musicDir = filepath.Join(home, "Music")
	}

	funcMap := template.FuncMap{
		"PathJoin": path.Join,
	}
	tmpl, err := template.New("index.m3u").Funcs(funcMap).Parse(indexM3u)
	if err != nil {
		log.Fatal(err)
	}

	ctx := NewContext(tmpl, musicDir)
	http.Handle("GET /", loggingHandler(http.HandlerFunc(ctx.handleRoot)))
	log.Fatal(http.ListenAndServe(":8080", nil))
}

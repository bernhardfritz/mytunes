package main

import (
	"crypto/sha256"
	"crypto/subtle"
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

	"github.com/alexedwards/scs/v2"
	"github.com/bernhardfritz/mytunes/itertools"
)

type Playlist struct {
	Path        string
	Directories []string
	Files       []string
}

var sessionManager *scs.SessionManager

func logger(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		log.Println(req.Method, req.RemoteAddr, req.URL.Path)
		handler.ServeHTTP(res, req)
	})
}

func basicAuth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		username, password, ok := req.BasicAuth()
		if ok {
			usernameHash := sha256.Sum256([]byte(username))
			passwordHash := sha256.Sum256([]byte(password))
			expectedUsernameHash := sha256.Sum256([]byte("admin"))
			expectedPasswordHash := sha256.Sum256([]byte(os.Getenv("MYTUNES_PASSWORD")))
			usernameMatch := (subtle.ConstantTimeCompare(usernameHash[:], expectedUsernameHash[:]) == 1)
			passwordMatch := (subtle.ConstantTimeCompare(passwordHash[:], expectedPasswordHash[:]) == 1)
			if usernameMatch && passwordMatch {
				sessionManager.Put(req.Context(), "authenticated", true)
				next.ServeHTTP(res, req)
				return
			}
		}

		authenticated := sessionManager.GetBool(req.Context(), "authenticated")
		if authenticated {
			next.ServeHTTP(res, req)
			return
		}

		res.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
		http.Error(res, "Unauthorized", http.StatusUnauthorized)
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

func handleRoot(res http.ResponseWriter, req *http.Request) {
	if containsDotDot(req.URL.Path) {
		http.Error(res, "invalid URL path", http.StatusBadRequest)
		return
	}

	if req.URL.Path[len(req.URL.Path)-1] == '/' {
		req.URL.Path += "index.m3u"
	}

	if strings.HasSuffix(req.URL.Path, "/index.m3u") {
		funcMap := template.FuncMap{
			"PathJoin": path.Join,
		}
		tmpl, err := template.New("index.m3u").Funcs(funcMap).ParseFiles("index.m3u")
		if err != nil {
			log.Fatal(err)
		}
		dir := http.Dir("/var/lib/mytunes")
		f, err := dir.Open(filepath.Dir(req.URL.Path))
		if err != nil {
			http.NotFound(res, req)
			return
		}
		defer f.Close()
		fileInfos, err := f.Readdir(-1)
		if err != nil {
			http.NotFound(res, req)
			return
		}
		playlist := Playlist{
			Path:        path.Dir(req.URL.Path),
			Directories: slices.Collect(itertools.Map(fs.FileInfo.Name, itertools.Filter(fs.FileInfo.IsDir, slices.Values(fileInfos)))),
			Files:       slices.Collect(itertools.Filter(itertools.HasSuffix(".mp3"), itertools.Map(fs.FileInfo.Name, itertools.Filter(itertools.Not(fs.FileInfo.IsDir), slices.Values(fileInfos))))),
		}
		err = tmpl.Execute(res, playlist)
		if err != nil {
			log.Fatal(err)
		}
	} else if strings.HasSuffix(req.URL.Path, ".m3u8") {
		relativePath := strings.TrimSuffix(req.URL.Path, ".m3u8")
		if !strings.HasSuffix(relativePath, ".mp3") {
			http.NotFound(res, req)
			return
		}
		dir := http.Dir("/var/lib/mytunes")
		f, err := dir.Open(relativePath)
		if err != nil {
			http.NotFound(res, req)
			return
		}
		defer f.Close()
		_, err = f.Stat()
		if err != nil {
			http.NotFound(res, req)
			return
		}
		input := filepath.Join("/var/lib/mytunes", relativePath)
		tmpDir := filepath.Join(os.TempDir(), "mytunes")
		output := filepath.Join(tmpDir, relativePath)
		err = os.MkdirAll(filepath.Dir(output), 0660)
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
		http.FileServer(http.Dir(tmpDir)).ServeHTTP(res, req)
	} else if strings.HasSuffix(req.URL.Path, ".ts") {
		tmpDir := filepath.Join(os.TempDir(), "mytunes")
		http.FileServer(http.Dir(tmpDir)).ServeHTTP(res, req)
	} else {
		dir := http.Dir("/var/lib/mytunes")
		f, err := dir.Open(req.URL.Path)
		if err != nil {
			http.NotFound(res, req)
			return
		}
		defer f.Close()

		fileInfo, err := f.Stat()
		if err != nil {
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

func main() {
	mytunesPassword := os.Getenv("MYTUNES_PASSWORD")
	if mytunesPassword == "" {
		log.Fatal("The MYTUNES_PASSWORD environment variable is empty or not set.")
	}
	sessionManager = scs.New()
	http.Handle("GET /", logger(sessionManager.LoadAndSave(basicAuth(http.HandlerFunc(handleRoot)))))
	log.Fatal(http.ListenAndServe(":8080", nil))
}

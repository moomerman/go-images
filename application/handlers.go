package application

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

// Router sets up all the route -> handler mappings for the Application
func (app *Application) Router() http.Handler {
	r := mux.NewRouter()
	r.Handle("/ping", app.PingHandler()).Methods("GET", "HEAD")
	r.Handle("/{bucket}/upload", app.UploadHandler()).Methods("POST")
	r.Handle("/{bucket}/url", app.URLHandler()).Methods("POST")
	r.Handle("/{processors}/{dimensions}/{bucket}/{key:.*}", app.ImageHandler()).Methods("GET")
	r.Handle("/{bucket}/{key:.*}", app.FileHandler()).Methods("GET")
	return r
}

// ImageHandler is responsible for returning an image modified by the parameters
func (app *Application) ImageHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		vars := mux.Vars(request)
		storage := app.Config.Storages[vars["bucket"]]

		if storage == nil {
			app.logger.Println("Unable to storage for bucket", vars["bucket"])
			http.Error(w, "404 Not Found", http.StatusNotFound)
			return
		}

		dimensions := NewDimensions(vars["dimensions"])

		original, err := app.ImageForKey(storage, vars["key"])

		if err != nil {
			http.Error(w, "404 Not Found", http.StatusNotFound)
			return
		}

		thumbnail := app.ThumbnailForParams(original, vars["processors"], dimensions)

		f, err := os.Open(thumbnail.Filename)
		defer f.Close()

		if err != nil {
			http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Cache-Control", "public, max-age="+strconv.Itoa(storage.CacheAge))
		io.Copy(w, f)
	})
}

// FileHandler is responsible for returning the original uploaded file
func (app *Application) FileHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		vars := mux.Vars(request)
		storage := app.Config.Storages[vars["bucket"]]

		if storage == nil {
			app.logger.Println("Unable to storage for bucket", vars["bucket"])
			http.Error(w, "404 Not Found", http.StatusNotFound)
			return
		}

		original, err := app.ImageForKey(storage, vars["key"])

		if err != nil {
			http.Error(w, "404 Not Found", http.StatusNotFound)
			return
		}

		f, err := os.Open(original.Filename)
		defer f.Close()

		if err != nil {
			http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", original.MimeType)
		w.Header().Set("Cache-Control", "public, max-age="+strconv.Itoa(storage.CacheAge))
		io.Copy(w, f)
	})
}

// UploadHandler is responsible for receiving a direct image upload, storing
// it and returning some stats about the image
func (app *Application) UploadHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		vars := mux.Vars(request)
		storage := app.Config.Storages[vars["bucket"]]

		file, header, err := request.FormFile("file")

		if err != nil {
			app.logger.Println("[UploadHandler] error retrieveing uploaded file:", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		defer file.Close()

		out, err := ioutil.TempFile("", "uploaded_file_")
		if err != nil {
			app.logger.Println("[UploadHandler] error creating temporary file:", err)
			fmt.Fprintf(w, "Unable to create the file for writing. Check your write access privilege")
			return
		}

		defer out.Close()

		_, err = io.Copy(out, file)
		if err != nil {
			app.logger.Println("[UploadHandler] error copying to temporary file:", err)
			fmt.Fprintln(w, err)
		}

		upload, err := app.StoreFile(storage, out, out.Name(), header.Filename)

		if err != nil {
			app.logger.Println("[UploadHandler] error storing file", err)
			return
		}

		upload.Identify()

		content, err := json.Marshal(map[string]string{
			"id":       upload.Key,
			"filename": header.Filename,
			"format":   upload.MimeType,
			"size":     strconv.FormatInt(upload.Size, 10),
		})

		fmt.Println("[UploadHandler]", string(content))

		w.Header().Set("Content-Type", "application/json")
		w.Write(content)
	})
}

// URLHandler takes an image URL, downloads it and then performs the same
// operations as UploadHandler (storing it and returning stats)
func (app *Application) URLHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		vars := mux.Vars(request)
		storage := app.Config.Storages[vars["bucket"]]

		request.ParseForm()

		url := request.FormValue("url")
		parts := strings.Split(url, "/")
		filename := parts[len(parts)-1]

		fmt.Println("fetching", url)

		// fetch the image from the URL specified
		response, err := http.Get(url)
		if err != nil {
			fmt.Fprintf(w, "Unable to fetch the url. Check it exists.")
			return
		}
		defer response.Body.Close()
		fmt.Println(response.Status)

		// TODO: way too much code duplication from here on with UploadHandler

		out, err := ioutil.TempFile("", "fetched_image_")
		if err != nil {
			fmt.Fprintf(w, "Unable to create the file for writing. Check your write access privilege")
			return
		}

		defer out.Close()

		_, err = io.Copy(out, response.Body)
		if err != nil {
			fmt.Fprintln(w, err)
		}

		image, err := app.StoreFile(storage, out, out.Name(), filename)

		if err != nil {
			app.logger.Println("Error storing file", err)
			return
		}

		image.Identify()

		content, err := json.Marshal(map[string]string{
			"id":       image.Key,
			"filename": filename,
			"format":   image.MimeType,
			"size":     strconv.FormatInt(image.Size, 10),
		})

		w.Header().Set("Content-Type", "application/json")
		w.Write(content)
	})
}

// PingHandler is used for monitoring and provides some internal instance stats
func (app *Application) PingHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		// memStats := &runtime.MemStats{}
		// runtime.ReadMemStats(memStats)

		content, err := json.MarshalIndent(map[string]interface{}{
			"response":     "pong",
			"numGoroutine": runtime.NumGoroutine(),
			// "memStats":     memStats,
			"goVersion": runtime.Version(),
			"version":   app.Config.Version,
			"stats":     app.stats.Data(),
		}, "", "  ")

		if err != nil {
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(content)
	})
}

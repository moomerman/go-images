package application

import (
	"io"
	"io/ioutil"
	"log"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/thoas/gostorages"
	"github.com/thoas/stats"
)

// Application holds the configuration and logger throughout the application
type Application struct {
	stats  *stats.Stats
	logger *log.Logger
	Config *Config
}

// NewApplication creates a new Application struct with the given config and
// default logger
func NewApplication(config *Config) *Application {
	return &Application{
		stats:  stats.New(),
		logger: log.New(os.Stdout, "[Application] ", log.Ldate|log.Ltime|log.Lshortfile),
		Config: config,
	}
}

// ImageForKey fetches the original image file for the specified key from
// disk cache or s3
func (app *Application) ImageForKey(storage *Storage, key string) (*Image, error) {
	filename := storage.Root + key

	// fetch the file from s3 not in the cache
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		remote, err := storage.S3().Open(key)

		if err != nil {
			app.logger.Println("Error opening connection to S3 for", key, err)
			return nil, err
		}

		defer remote.Close()

		os.MkdirAll(filepath.Dir(filename), 0777)

		local, err := os.Create(filename)

		if err != nil {
			app.logger.Println("Error creating file", filename, err)
			return nil, err
		}

		defer local.Close()

		app.logger.Println("downloading: ", key, filename)
		io.Copy(local, remote)
	}

	extension := filepath.Ext(filename)

	return &Image{
		Key:       key,
		Filename:  filename,
		Extension: extension,
		MimeType:  mime.TypeByExtension(extension),
	}, nil
}

// ThumbnailForParams takes a local image and applies transform operations
// returning the result
func (app *Application) ThumbnailForParams(original *Image, processors string, dimensions *Dimensions) *Image {
	filename := original.Filename + "+" + processors + "_" + dimensions.String

	thumbnail := &Image{
		Key:      original.Key,
		Filename: filename,
	}

	// generate the thumbnail if not in the cache
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		if strings.Contains(processors, "resize") {
			original.Resize(dimensions, filename)
		}

		if strings.Contains(processors, "crop") || strings.Contains(processors, "fill") {
			original.Fill(dimensions, filename)
		}

		thumbnail.Optimize()
	}

	return thumbnail
}

// StoreFile takes a recently uploaded file and stores it in the storage
func (app *Application) StoreFile(storage *Storage, file io.Reader, path string, filename string) (*Image, error) {

	// TODO: check the filesize and mime type / format before accepting
	upload := &Image{Filename: path, OriginalFilename: filename}
	upload.Identify()

	app.logger.Println("[StoreFile] temporary file", upload)

	md5, err := ComputeFileMd5(path)
	if err != nil {
		return nil, err
	}
	key := "upload/" + md5 + upload.Extension
	newFilename := storage.Root + key

	os.MkdirAll(filepath.Dir(newFilename), 0777)
	os.Rename(path, newFilename)

	app.logger.Println("[StoreFile] key:", key, newFilename)

	go app.Upload(storage, newFilename, key)

	upload.Filename = newFilename
	upload.Key = storage.Bucket + "/" + key

	return upload, nil
}

// Upload optimizes the file and uploads to S3 (to be called async)
func (app *Application) Upload(storage *Storage, filename string, key string) {

	upload := &Image{Filename: filename, Key: key}

	if storage.S3().Exists(key) {
		app.logger.Println("[Upload] already exists:", key)
		return
	}

	upload.Optimize()

	f, err := os.Open(filename)

	if err != nil {
		app.logger.Println("[Upload] error opening new file:", err)
		return
	}

	defer f.Close()

	content, err := ioutil.ReadAll(f)

	if err != nil {
		app.logger.Println("[Upload] Error reading new file:", err)
		return
	}

	app.logger.Println("[Upload] uploading: ", filename, key)
	err = storage.S3().Save(key, gostorages.NewContentFile(content))
	app.logger.Println("[Upload] uploaded: ", filename, key)

	if err != nil {
		app.logger.Println("[Upload] error uploading to S3:", err)
		return
	}
}

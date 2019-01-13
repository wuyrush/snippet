package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"time"
	"unicode/utf8"

	log "github.com/Sirupsen/logrus"
	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
	"github.com/kelseyhightower/envconfig"
	uuid "github.com/satori/go.uuid"
)

const (
	MULTIPART_FORM_BUFFER_SIZE_BYTES = 1 << 12
)

var (
	SnippetRetentionTime time.Duration
	store                SnippetStore
	supportedModes       = map[string]bool{
		"python":     true,
		"golang":     true,
		"rust":       true,
		"javascript": true,
		"text":       true,
	}
)

func SaveSnippetHandler(w http.ResponseWriter, req *http.Request) {
	if err := req.ParseMultipartForm(MULTIPART_FORM_BUFFER_SIZE_BYTES); err != nil {
		log.WithError(err).Error("SaveSnippetHandler: Error when parsing form")
		writeError(w, 400, "Got malformed form data")
		return
	}
	// validate client input, then generate a snippet object from the input if it is valid
	snippet, err := createSnippet(req.MultipartForm)
	if err != nil {
		log.WithError(err).Error("SaveSnippetHandler: Failed to create snippet.")
		if e, ok := err.(*Error); ok {
			writeError(w, e.Code, e.Error())
			return
		}
		w.WriteHeader(500)
		return
	}
	log.WithField("snippetId", snippet.Id).Info("SaveSnippetHandler: Snippet created")
	// save to storage backend
	err = store.Save(snippet)
	if err != nil {
		log.WithError(err).Error("SaveSnippetHandler: Failed to store snippet data")
		w.WriteHeader(500)
		return
	}
	log.WithField("snippetId", snippet.Id).Info("SaveSnippetHandler: Snippet saved to storage")
	// respond with snippet id so that client is able to view the saved snippet
	resp := struct {
		Id string `json:"snippetId"`
	}{snippet.Id}
	if jsonBlob, err := json.Marshal(resp); err != nil {
		writeError(w, 500, "Failed to generated response data")
	} else if _, err = w.Write(jsonBlob); err != nil {
		log.WithError(err).Error("SaveSnippetHandler: Failed to write response to client")
	}
}

func createSnippet(form *multipart.Form) (snippet *Snippet, err error) {
	values := form.Value
	if len(values["snippetName"]) == 0 || len(values["snippetText"]) == 0 || len(values["mode"]) == 0 {
		return nil, &Error{Message: "Missing form field", Code: 400}
	}
	snippet = &Snippet{
		Name: values["snippetName"][0],
		Body: values["snippetText"][0],
		Mode: values["mode"][0],
	}

	if utf8.RuneCountInString(snippet.Body) == 0 {
		return nil, &Error{Message: "Snippet body is empty", Code: 400}
	}

	if _, ok := supportedModes[snippet.Mode]; !ok {
		return nil, &Error{Message: "Unsupported mode", Code: 400}
	}

	// generate snippet id
	snippetUUID, e := uuid.NewV4()
	if e != nil {
		log.WithError(e).Error("createSnippet: Failed to generate snippet id.")
		return nil, &Error{Message: "Failed to generate ID for the snippet", Code: 500, Cause: e}
	}
	// discard dashes since it complicates id validation
	snippet.Id = hex.EncodeToString(snippetUUID.Bytes())
	// set creation and expiration time
	timeCreated := time.Now().UTC()
	snippet.TimeCreated = timeCreated.Unix()
	if utf8.RuneCountInString(snippet.Name) == 0 {
		// fall back to default
		snippet.Name = fmt.Sprintf("Snippet created at %s", timeCreated.Format("Mon Jan _2 15:04:05 MST 2006"))
	}
	snippet.TimeExpired = snippet.TimeCreated + int64(SnippetRetentionTime.Seconds())

	return snippet, nil
}

func ViewSnippetHandler(w http.ResponseWriter, req *http.Request) {
	snippetId := mux.Vars(req)["id"]
	// retrieve snippet data from storage backend
	snippet, err := store.Get(snippetId)
	if err != nil {
		log.WithError(err).Errorf("Failed to retrieve data of snippet with id %s", snippetId)
		if e, ok := err.(*Error); ok {
			writeError(w, e.Code, e.Error())
			return
		}
		w.WriteHeader(500)
		return
	}
	// Got the snippet with expected id. Return it in form of JSON
	log.WithField("snippetData", *snippet).Info("Retrieved snippet data successfully")
	jsonBlob, err := json.Marshal(snippet)
	if err != nil {
		log.WithError(err).Errorf("Failed to jsonify data of snippet with id %s", snippetId)
		writeError(w, 500, "Got malformed snippet data")
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	// TODO: maybe we should use closures to handle such boilerplate error handling logic
	if _, err = w.Write(jsonBlob); err != nil {
		log.WithError(err).Errorf("Failed to write response data back to client for snippet with id %s",
			snippetId)
	}
}

func writeError(w http.ResponseWriter, statusCode int, message string) {
	w.WriteHeader(statusCode)
	if _, err := w.Write([]byte(message)); err != nil {
		log.WithError(err).Error("writeError: Got error when writing data to client", err)
	}
}

func setupLogging(verbose bool) {
	// use JSON format so that the resulting log entries are easy to be consumed by log analyzers
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
	if verbose {
		log.SetLevel(log.DebugLevel)
	}
}

func setupRedis(config *Config) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:       config.RedisUrl,
		Password:   config.RedisPasswd,
		DB:         0, // what's the difference between default db and customized one?
		MaxRetries: config.RedisMaxRetries,
		MaxConnAge: config.RedisMaxConnAge,
		PoolSize:   config.RedisMaxConnPoolSize,
	})
}

func setupStore(config *Config) {
	redisDB := setupRedis(config)
	store = &RedisSnippetStore{
		db:            redisDB,
		retentionTime: config.SnippetRetentionTime,
	}
}

func main() {
	var config Config
	err := envconfig.Process("", &config)
	if err != nil {
		log.WithError(err).Fatal("main: Error when reading application configurations.")
	}

	setupLogging(config.Verbose)
	setupStore(&config)
	SnippetRetentionTime = config.SnippetRetentionTime

	// setup router
	r := mux.NewRouter()
	r.HandleFunc("/view/{id:[0-9a-f]{32}}", ViewSnippetHandler).Methods("GET")
	r.HandleFunc("/save", SaveSnippetHandler).Methods("POST")
	// hook the router to standard server mux
	http.Handle("/", r)
	// start the server
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	log.WithField("addressToListen", addr).Debug("main: Starting server")

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.WithError(err).Fatal("main: Error when listening or serving requests")
	}
}

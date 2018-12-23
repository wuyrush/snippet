package main

import (
    "fmt"
    "net/http"
    "mime/multipart"
    "unicode/utf8"
    "time"
    "os"
    "encoding/hex"
    "encoding/json"

    "github.com/satori/go.uuid"
    "github.com/go-redis/redis"
    log "github.com/Sirupsen/logrus"
    "github.com/kelseyhightower/envconfig"
    "github.com/gorilla/mux"
)

const (
    MULTIPART_FORM_BUFFER_SIZE_BYTES = 1 << 12
)

var supportedModes = map[string]bool{
    "python": true,
    "golang": true,
    "rust":   true,
    "javascript": true,
    "text": true,
}

// interface to data storage layer
var store SnippetStore

func SaveSnippetHandler(w http.ResponseWriter, req *http.Request) {
    if err := req.ParseMultipartForm(MULTIPART_FORM_BUFFER_SIZE_BYTES); err != nil {
        log.WithError(err).Error("SaveSnippetHandler: Error when parsing form")
    }
    // validate client input, then generate a snippet object from the input if it is valid
    snippet, err := createSnippet(req.MultipartForm)
    if err != nil {
        // TODO: notify the user about the error, in form of error notifications
        log.WithError(err).Error("SaveSnippetHandler: Failed to create snippet.")
        return
    }
    log.WithField("snippetId", snippet.Id).Info("SaveSnippetHandler: Snippet created")
    // save to storage backend
    err = store.Save(snippet)
    if err != nil {
        // TODO: notify the user about the error, in form of error notifications
        log.WithError(err).Error("SaveSnippetHandler: Failed to store snippet data")
        return
    }
    log.WithField("snippetId", snippet.Id).Info("SaveSnippetHandler: Snippet saved to data backend")
    w.WriteHeader(200)
}

func createSnippet(form *multipart.Form) (snippet *Snippet, err error) {
    defer func() {
        if e := recover(); e != nil {
            log.WithField("recovered", e).Error("createSnippet: Failed to create snippet from user input.")
            err = &Error{Message: "Unknown service error occurred.", Code: 500, Cause: e}
        }
    }()
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
    snippetUUID, e:= uuid.NewV4()
    if e != nil {
        errorMsg := "createSnippet: Failed to generate snippet id."
        log.WithError(e).Error(errorMsg)
        return nil, &Error{Message: errorMsg, Code: 500, Cause: e}
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
    // expire 5 min after saving
    snippet.TimeExpired = snippet.TimeCreated + 300

    return snippet, nil
}

func ViewSnippetHandler(w http.ResponseWriter, req *http.Request) {
    snippetId := mux.Vars(req)["id"]
    // retrieve snippet data from storage backend
    snippet, err := store.Get(snippetId)
    if err != nil {
        log.WithError(err).Errorf("Failed to retrieve data of snippet with id %s", snippetId)
        switch e := err.(type) {
        case *Error:
            http.NotFound(w, req)
            return
        default:
            // TODO: service-side error. Notify user
            _ = e
        }
    }
    // Got the snippet with expected id. Return it in form of JSON
    log.WithField("snippetData", *snippet).Info("Retrieved snippet data successfully")
    jsonBlob, err := json.Marshal(snippet)
    if err != nil {
        log.WithError(err).Errorf("Failed to jsonify data of snippet with id %s", snippetId)
        return
    }

    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    // TODO: maybe we should use closures to handle such boilerplate error handling logic
    if _, err = w.Write(jsonBlob); err != nil {
        log.WithError(err).Errorf("Failed to write response data back to client for snippet with id %s",
            snippetId)
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
        Addr: config.RedisUrl,
        Password: config.RedisPasswd,
        DB: 0,   // what's the difference between default db and customized one?
        MaxRetries: config.RedisMaxRetries,
        MaxConnAge: config.RedisMaxConnAge,
        PoolSize: config.RedisMaxConnPoolSize,
    })
}

func setupStore(config *Config) {
    redisDB := setupRedis(config)
    store = &RedisSnippetStore{
        db: redisDB,
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


package main

import (
    "fmt"
    "net/http"
    "mime/multipart"
    "unicode/utf8"
    "time"
    "os"

    "github.com/satori/go.uuid"
    "github.com/go-redis/redis"
    log "github.com/Sirupsen/logrus"
    "github.com/kelseyhightower/envconfig"
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
    snippet.Id = snippetUUID.String()
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
    // TODO: returns snippet data from storage backend
}

func setupLogging(verbose bool) {
    log.SetFormatter(&log.TextFormatter{
        DisableColors: false,
    })
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
        retentionTimeSeconds: config.SnippetRetentionPeriodSeconds,
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
    http.HandleFunc("/save", SaveSnippetHandler)
    // start the server
    addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
    log.WithField("addressToListen", addr).Debug("main: Starting server")

    if err := http.ListenAndServe(addr, nil); err != nil {
        log.WithError(err).Fatal("main: Error when listening or serving requests")
    }
}


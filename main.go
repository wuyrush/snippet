package main

import (
    "fmt"
    "flag"
    "log"
    "net/http"
    "mime/multipart"
    "unicode/utf8"
    "time"

    "github.com/satori/go.uuid"
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

/*
Entities involved in our application:

1. Snippet
    1. snippet-content
        1. snippet-name - string
        1. snippet-body - string
        1. language of that snippet belongs to (default is text) - string
    1. TODO: snippet owner - need User entity - user id, string
    1. Is the snippet open to public? (default to true if not logged in and false if logged in)
    1. snippet id - string
    1. expiration time - use epoch time in seconds
    1. creation time - same as above

We design the snippet content to be immutable once it is saved into our application.

TODO
2. User
    1. username - string
    1. user email for registration and login - string, must be unique
    1. user password - store the string / byte of password hash

*/

type Snippet struct {
    Name            string
    Body            string
    Mode            string
    Id              string
    TimeExpired     int64
    TimeCreated     int64
    UserId          string
}

type User struct {
    id          string  // user email
    username    string
}

func SaveSnippetHandler(w http.ResponseWriter, req *http.Request) {
    log.Println("Got request with method", req.Method)
    if err := req.ParseMultipartForm(MULTIPART_FORM_BUFFER_SIZE_BYTES); err != nil {
        log.Fatal("Error when parsing form:", err)
    }
    // validate client input, then generate a snippet object from the input if it is valid
    snippet, err := createSnippet(req.MultipartForm)
    if err != nil {
        // notify the user about the error
        log.Println("SaveSnippetHandler: Failed to create snippet.", err)
        return
    }
    log.Println("Snippet created:", *snippet)
}

func createSnippet(form *multipart.Form) (snippet *Snippet, err error) {
    defer func() {
        if e := recover(); e != nil {
            log.Println("createSnippet: Failed to create snippet from user input.", e)
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
        errorMsg := "Failed to generate snippet id."
        log.Println(errorMsg, e)
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
    snippet.TimeExpired = snippet.TimeCreated + 1800

    return snippet, nil
}

type Error struct {
    Message string
    Code    int
    Cause   interface{}
}

func (e *Error) Error() string {
    // If not explicitly specified, all the error shall be deemed as service-side error since the
    // service failed to determine what it is.
    side := "Service"
    if e.Code >= 400 && e.Code < 500 {
        side = "Client"
    }
    errStr := fmt.Sprintf("%sError: %s", side, e.Message)
    if e.Cause != nil {
        errStr += fmt.Sprintf(" Cause: %v", e.Cause)
    }
    return errStr
}

func ViewSnippetHandler(w http.ResponseWriter, req *http.Request) {
    // TODO: returns snippet data from storage backend
}

func main() {
    // define application CLI
    addr := flag.String("addr", ":3000", "service address")
    flag.Parse()

    http.HandleFunc("/save", SaveSnippetHandler)
    // start the server
    err := http.ListenAndServe(*addr, nil)
    if err != nil {
        log.Fatal("http.ListenAndServe: ", err)
    }
}

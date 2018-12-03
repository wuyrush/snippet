package main

import (
    "flag"
    "log"
    "net/http"
)

var addr = flag.String("addr", ":3000", "service address")

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
    Lang            string
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
    log.Println("Got request: ", *req)
    if err := req.ParseForm(); err != nil {
        log.Fatal("Error when parsing form: ", err)
    }
    log.Println("Form value: ", req.PostForm)
}

func ViewSnippetHandler(w http.ResponseWriter, req *http.Request) {
    // TODO: returns snippet data from storage backend
}

func main() {
    flag.Parse()
    http.HandleFunc("/save", SaveSnippetHandler)
    // start the server
    err := http.ListenAndServe(*addr, nil)
    if err != nil {
        log.Fatal("http.ListenAndServe: ", err)
    }
}

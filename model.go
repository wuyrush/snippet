package main

import (
    "time"
    "fmt"
)

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

// models snippets
type Snippet struct {
    Name            string  `json:"snippetName"`
    Body            string  `json:"snippetText"`
    Mode            string  `json:"mode"`
    Id              string  `json:"-"` // snippet ID is of no use on client side for now
    TimeExpired     int64   `json:"timeExpired"`
    TimeCreated     int64   `json:"timeCreated"`
    UserId          string  `json:"-"`
}

type User struct {
    id          string  // user email
    username    string
}

// holds application config values
type Config struct {
    Host string `required:"true"`
    Port int `required:"true"`
    Verbose bool `default:"false"`
    SnippetRetentionTime time.Duration `split_words:"true" required:"true"`
    RedisUrl string `split_words:"true" required:"true"`
    RedisMaxConnPoolSize int `split_words:"true" default: 10`
    RedisMaxRetries int `split_words:"true" default: 2`
    RedisMaxConnAge time.Duration `split_words:"true"`
    RedisPasswd string `split_words:"true" required:"true"`
}

// represents application errors
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

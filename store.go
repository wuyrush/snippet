package main

import (
    "fmt"
    "time"
    "strconv"
    "github.com/go-redis/redis"
    log "github.com/Sirupsen/logrus"
)

type SnippetStore interface {
    Get(id string) (*Snippet, error)
    Save(snippet *Snippet) error
}

type RedisSnippetStore struct {
    db *redis.Client
    retentionTime time.Duration // default snippet retention time in seconds
}

func (store *RedisSnippetStore) Get(snippetId string) (*Snippet, error) {
    re, err := store.db.HGetAll(snippetId).Result()
    // note if the key had already expired then we will get an empty map and err will be set to nil
    log.WithFields(log.Fields{
        "mapFromRedius": re,
        "error": err,
    }).Debug("Got result from redis")

    if err != nil {
        log.WithError(err).Errorf("Failed to get data of snippet %s from Redis.", snippetId)
        return nil, &Error{
            Message: "Failed to retrieve snippet data from storage",
            Code: 500,
            Cause: err,
        }
    } else if len(re) == 0 {
        // snippet with specified id not in redis - already expired
        return nil, &Error{
            Message: fmt.Sprintf("Data for snippet %s not found", snippetId),
            Code: 404,
        }
    }
    // generate snippet object and return
    timeCreated, err := strconv.ParseInt(re["timeCreated"], 10, 64)
    if err != nil {
        log.WithError(err).Errorf("Got malformed creation timestamp %s for snippet %s from Redis.",
            re["timeCreated"], snippetId)
        return nil, &Error{
            Message: "Malformed created timestamp found in snippet data",
            Code: 500,
            Cause: err,
        }
    }
    timeExpired, err := strconv.ParseInt(re["timeExpired"], 10, 64)
    if err != nil {
        log.WithError(err).Errorf("Got malformed expiration timestamp %s for snippet %s from Redis.",
            re["timeExpired"], snippetId)
        return nil, &Error{
            Message: "Malformed expiration timestamp found in snippet data",
            Code: 500,
            Cause: err,
        }
    }
    return &Snippet{
        Name: re["name"],
        Body: re["body"],
        Mode: re["mode"],
        TimeCreated: timeCreated,
        TimeExpired: timeExpired,
        UserId: re["userId"],
    }, nil
}

func (store *RedisSnippetStore) Save(snippet *Snippet) error {
    if _, err := store.db.HMSet(snippet.Id, map[string]interface{}{
        "name": snippet.Name,
        "body": snippet.Body,
        "mode": snippet.Mode,
        "timeCreated": snippet.TimeCreated,
        "timeExpired": snippet.TimeExpired,
        "userId": snippet.UserId,
    }).Result(); err != nil {
        log.WithError(err).Errorf("Failed to save snippet %s to Redis", snippet.Id)
        return err
    }
    log.Debugf("Snippet %s saved to Redis successfully", snippet.Id)

    if _, err := store.db.Expire(snippet.Id, store.retentionTime).Result(); err != nil {
        log.WithError(err).Errorf("Failed to set expiration time on snippet %s", snippet.Id)
        return err
    }
    log.Debugf("Set expiration time of snippet %s to %g seconds", snippet.Id, store.retentionTime.Seconds())
    return nil
}

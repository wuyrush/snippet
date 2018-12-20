package main

import (
    "github.com/go-redis/redis"
    log "github.com/Sirupsen/logrus"
)

type SnippetStore interface {
    Get(id string) (*Snippet, error)
    Save(snippet *Snippet) error
}

type RedisSnippetStore struct {
    db *redis.Client
    retentionTimeSeconds int  // default snippet retention time in seconds
}

func (store *RedisSnippetStore) Get(id string) (*Snippet, error) {
    // TODO
    return nil, nil
}

func (store *RedisSnippetStore) Save(snippet *Snippet) error {
    re, err := store.db.HMSet(snippet.Id, map[string]interface{}{
        "name": snippet.Name,
        "body": snippet.Body,
        "mode": snippet.Mode,
        "timeCreated": snippet.TimeCreated,
        "timeExpired": snippet.TimeExpired,
        "userId": snippet.UserId,
    }).Result()
    log.WithFields(log.Fields{
        "result": re,
        "error": err,
    }).Debug("HMSET command executed")
    return err
}

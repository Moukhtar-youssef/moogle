package database

import(
    "log"
    "context"

    "github.com/redis/go-redis/v9"
)

type Database struct {
    Client      *redis.Client
    Context     context.Context
}

func (db *Database) ConnectToRedis() {
    log.Println("Connecting to Redis...")

    db.Client = redis.NewClient(&redis.Options{
        Addr:       "localhost:6379",
        Password:   "",
        DB:         0,
        Protocol:   2,
    })

    db.Context = context.Background()

    _, err := db.Client.Ping(db.Context).Result()
    if err != nil {
        log.Printf("Error connecting to redis: %v", err)
    }

    log.Println("Successfully connected to Redis!")
}


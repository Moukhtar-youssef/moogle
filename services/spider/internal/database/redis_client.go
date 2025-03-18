package database

import(
    "fmt"
    "log"
    "context"
    "strconv"

    "github.com/redis/go-redis/v9"

    "github.com/IonelPopJara/search-engine/services/spider/internal/utils"
)

type Database struct {
    Client      *redis.Client
    Context     context.Context
}

func (db *Database) ConnectToRedis(redisHost, redisPort, redisPassword, redisDB string) error {
    log.Println("JUST TESTING THIS!\n")
    log.Println("Connecting to Redis...")
    log.Printf("\tRedis Host: '%s'\n", redisHost+":"+redisPort)
    log.Printf("\tRedis Password: '%s'\n", redisPassword)
    log.Printf("\tRedis DB: '%s'\n", redisDB)

    dbIndex, err := strconv.Atoi(redisDB)
    if err != nil {
        return fmt.Errorf("Could not parse DB value: %v\n", err)
    }

    db.Client = redis.NewClient(&redis.Options{
        Addr:       redisHost+":"+redisPort,
        Password:   redisPassword,
        DB:         dbIndex,
    })

    db.Context = context.Background()

    _, err = db.Client.Ping(db.Context).Result()
    if err != nil {
        return fmt.Errorf("Couldn't connect to redis: %v", err)
    }

    log.Println("Successfully connected to Redis!")
    return nil
}

func (db *Database) PushURL(rawURL string, score float64) error {
    // Normalize URL
    normalizedURL, err := utils.NormalizeURL(rawURL)
    if err != nil {
        return fmt.Errorf("Could not normalize starting URL: %v\n", err)
    }

    // Add key-value pair {normalized_url:normalizedURL: rawURL}
    setKey := "normalized_url:"+normalizedURL
    err = db.Client.Set(db.Context, setKey, rawURL, 0).Err()
    if err != nil {
        return fmt.Errorf("Could not store key-value pair {%v: %v}: %v\n", setKey, rawURL, err)
    }

    // Add the normalized url to a sorted set with 0 as score
    err = db.Client.ZAdd(db.Context, utils.URLQueueKey, redis.Z{
        Score: score,
        Member: normalizedURL,
    }).Err()

    if err != nil {
        return fmt.Errorf("Could not add %v to '%s'\n", normalizedURL, utils.URLQueueKey)
    }

    return nil
}

func (db *Database) PopURL() (string, float64, error) {
    // Get the next normalized URL from the priority queue
    result, err := db.Client.BZPopMin(db.Context, utils.Timeout, utils.URLQueueKey).Result()
    if err != nil {
        return "", 0.0, fmt.Errorf("Could not pop URL from queue: %v\n", err)
    }

    // Format the proper Redis queue to fetch data
    normalized_url_key := fmt.Sprintf("%v:%v", utils.NormalizedURLPrefix, result.Z.Member)

    // Fetch the raw url from Redis
    raw_url, err := db.Client.Get(db.Context, normalized_url_key).Result()
    if err != nil {
        return "", 0.0, fmt.Errorf("Could not fetch raw URL from %v: %v\n", normalized_url_key, err)
    }

    return raw_url, result.Z.Score, nil
}

func (db *Database) ExistsInQueue(rawURL string) (float64, bool) {
    // Normalize URL
    normalizedURL, err := utils.NormalizeURL(rawURL)
    if err != nil {
        return 0.0, false
    }

    result, err := db.Client.ZScore(db.Context, utils.URLQueueKey, normalizedURL).Result()
    if err != nil {
        return 0.0, false
    }

    return result, true
}

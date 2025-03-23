package database

import(
    "fmt"
    "log"
    "context"
    "strconv"

    "github.com/redis/go-redis/v9"

    "github.com/IonelPopJara/search-engine/services/spider/internal/utils"
)

// ------------------- REDIS SETUP -------------------
type Database struct {
    Client      *redis.Client
    Context     context.Context
}

func (db *Database) ConnectToRedis(redisHost, redisPort, redisPassword, redisDB string) error {
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
        return fmt.Errorf("Couldn't connect to shit [%v, %v]: %v", redisHost, redisPassword, err)
    }

    log.Println("Successfully connected to Redis!")
    return nil
}
// ------------------- REDIS SETUP -------------------

// ------------------- CRAWL LINKS -------------------
func (db *Database) addURLLookup(rawURL, normalizedURL string) error {
    // Check if it exists
    urlKey := utils.NormalizedURLPrefix + ":" + normalizedURL

    exists, err := db.Client.Exists(db.Context, urlKey).Result()
    if err != nil {
        return fmt.Errorf("Key not found: %w", err)
    }

    if exists > 0 {
        return fmt.Errorf("Key exists")
    }

    // Add hash
    err = db.Client.HSet(db.Context, urlKey, map[string]interface{}{
        "raw_url": rawURL,
        "visited": 0,
    }).Err()

    if err != nil {
        return fmt.Errorf("Could not store data in Redis: %w", err)
    }

    return nil
}

func (db *Database) PushURL(rawURL string, score float64) error {
    // Remove fragments and queries from rawURL
    rawURL, err := utils.StripURL(rawURL)
    if err != nil {
        return fmt.Errorf("Could not strip URL: %w", err)
    }

    // Normalize URL
    normalizedURL, err := utils.NormalizeURL(rawURL)
    if err != nil {
        return fmt.Errorf("Could not normalize URL: %w", err)
    }

    err = db.addURLLookup(rawURL, normalizedURL)
    if err != nil {
        // If the url is already in the lookup, we don't add it
        return fmt.Errorf("URL already in queue: %w", err)
    }

    // Add the normalized url to a sorted set with 0 as score
    err = db.Client.ZAdd(db.Context, utils.SpiderQueueKey, redis.Z{
        Score: score,
        Member: normalizedURL,
    }).Err()

    if err != nil {
        return fmt.Errorf("Could not add URL to queue: %w", err)
    }

    return nil
}

func (db *Database) ExistsInQueue(rawURL string) (float64, bool) {
    // Normalize URL
    normalizedURL, err := utils.NormalizeURL(rawURL)
    if err != nil {
        return 0.0, false
    }

    result, err := db.Client.ZScore(db.Context, utils.SpiderQueueKey, normalizedURL).Result()
    if err != nil {
        return 0.0, false
    }

    return result, true
}
// ------------------- CRAWL LINKS -------------------

// ------------------- VISIT PAGE -------------------
func (db *Database) HasURLBeenVisited(normalizedURL string) (bool, error) {
    normalized_url_key := utils.NormalizedURLPrefix + ":" + normalizedURL
    result, err := db.Client.HGet(db.Context, normalized_url_key, "visited").Result()

    if err == redis.Nil {
        return false, nil
    } else if err != nil {
        return false, fmt.Errorf("Could not fetch %v from Redis: %v\n", normalized_url_key, err)
    }

    visited, err := strconv.Atoi(result)
    if err != nil {
        return false, fmt.Errorf("Could not parse 'visited' value: %v", err)
    }

    if visited == 0 {
        return false, nil
    }

    return true, nil
}

func (db *Database) VisitPage(normalizedURL string) error {
    normalized_url_key := utils.NormalizedURLPrefix + ":" + normalizedURL
    _, err := db.Client.HSet(db.Context, normalized_url_key, "visited", 1).Result()

    if err != nil {
        return fmt.Errorf("Could not update visit %v from Redis: %v\n", normalized_url_key, err)
    }

    return nil
}
// ------------------- VISIT PAGE -------------------

// ------------------- GET NEXT ENTRY -------------------
func (db *Database) PopURL() (string, float64, error) {
    // Get the next normalized URL from the priority queue
    result, err := db.Client.BZPopMin(db.Context, utils.Timeout, utils.SpiderQueueKey).Result()
    if err != nil {
        return "", 0.0, fmt.Errorf("Could not pop URL from queue: %v\n", err)
    }

    // Format the proper Redis queue to fetch data
    normalized_url_key := fmt.Sprintf("%v:%v", utils.NormalizedURLPrefix, result.Z.Member)

    // Fetch the raw url from Redis
    raw_url, err := db.Client.HGet(db.Context, normalized_url_key, "raw_url").Result()
    if err != nil {
        return "", 0.0, fmt.Errorf("Could not fetch raw URL from %v: %v\n", normalized_url_key, err)
    }

    return raw_url, result.Z.Score, nil
}

func (db *Database) PopSignalQueue() (string, error) {
    result, err := db.Client.BRPop(db.Context, 0, utils.SignalQueueKey).Result()
    if err != nil {
        return "", fmt.Errorf("Could not pop from signal queue: %v\n", err)
    }

    return result[1], nil
}

func (db *Database) GetIndexerQueueSize() (int64, error) {
    size, err := db.Client.LLen(db.Context, utils.IndexerQueueKey).Result()
    if err != nil {
        return -1, fmt.Errorf("Could not get %v size: %v\n", utils.IndexerQueueKey, err)
    }

    return size, nil
}
// ------------------- GET NEXT ENTRY -------------------

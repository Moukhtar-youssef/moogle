package main

import (
	"context"
	"fmt"
	"os"
	"sort"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return fallback
}

func main() {
	mongoHost := getEnv("MONGO_HOST", "localhost")
	mongoPassword := getEnv("MONGO_PASSWORD", "")
	mongoUsername := getEnv("MONGO_USERNAME", "")

	fmt.Println("Page Rank Service!")

	mongoURI := fmt.Sprintf("mongodb://%s:%s@%s:27017/", mongoUsername, mongoPassword, mongoHost)

	client, err := mongo.Connect(options.Client().ApplyURI(mongoURI))
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		panic(fmt.Sprintf("Could not ping MongoDB: %v", err))
	}

	fmt.Println("Successfully connected to MongoDB!")

	// Access the test database
	db := client.Database("test")

	// Access the outlinks collection
	coll := db.Collection("outlinks")

	// Get the count of documents in the collection
	count, err := coll.CountDocuments(context.TODO(), bson.D{})
	if err != nil {
		panic(fmt.Sprintf("Could not count documents: %v", err))
	}

	fmt.Printf("Number of documents in the collection: %d\n", count)

	// Iterate over the documents in the collection
	cursor, err := coll.Find(context.TODO(), bson.D{})
	if err != nil {
		panic(fmt.Sprintf("Could not find documents: %v", err))
	}

	defer cursor.Close(context.TODO())

	// Page rank setup
	// Create a hash map to hold the page_url and its corresponding page rank
	pageRank := make(map[string]float64)

	for cursor.Next(context.TODO()) {
		var result bson.M
		if err := cursor.Decode(&result); err != nil {
			panic(fmt.Sprintf("Could not decode document: %v", err))
		}

		// Get the _id field, this is the page_url
		url, ok := result["_id"].(string)
		if !ok {
			panic("Could not convert _id to string")
		}

		// Assign a starting page rank value
		pageRank[url] = 1.0 / float64(count)
	}

	if err := cursor.Err(); err != nil {
		panic(fmt.Sprintf("Cursor error: %v", err))
	}

	// Print the initial page rank values
	fmt.Println("Initial Page Rank values:")
	for url, rank := range pageRank {
		fmt.Printf("Page URL: %s, Page Rank: %f\n", url, rank)
	}

	// Page rank algorithm
	// Set the number of iterations
	iterations := 10
	for range iterations {
		// Create a temporary hash map to hold the new page rank values
		newPageRank := make(map[string]float64)

		// Calculate the new page rank values
		for url, rank := range pageRank {
			fmt.Printf("Calculating new page rank for URL: %s | Previous Rank: %v\n", url, rank)

			// Get the backlinks for the current URL
			var backlinksDoc struct {
				Links []string `bson:"links"`
			}

			// Get the backlinks for the current URL
			err := db.Collection("backlinks").FindOne(context.TODO(), bson.D{{Key: "_id", Value: url}}).Decode(&backlinksDoc)
			if err != nil {
				if err == mongo.ErrNoDocuments {
					// No backlinks found for this URL
					fmt.Printf("No backlinks found for URL %s\n", url)
				} else {
					panic(fmt.Sprintf("Could not find backlinks for URL %s: %v", url, err))
				}
				continue
			}

			// Get the count of backlinks
			backlinksCount := len(backlinksDoc.Links)
			fmt.Printf("\tFound %d backlinks for URL: %s\n", backlinksCount, url)

			newCumulativeRank := 0.0

			// Iterate over the backlinks and calculate the new page rank
			for _, backlink := range backlinksDoc.Links {
				// Get the outlink document for the specified backlink
				var outlinkDoc struct {
					Links []string `bson:"links"`
				}

				// Get the count of outlinks
				err := db.Collection("outlinks").FindOne(context.TODO(), bson.D{{Key: "_id", Value: backlink}}).Decode(&outlinkDoc)
				if err != nil {
					if err == mongo.ErrNoDocuments {
						// No outlinks found for this URL
						fmt.Printf("No outlinks found for URL %s\n", backlink)
					} else {
						panic(fmt.Sprintf("Could not find outlinks for URL %s: %v", backlink, err))
					}
					continue
				}
				outlinksCount := len(outlinkDoc.Links)
				// fmt.Printf("\t\tFound %d outlinks for URL: %s\n", outlinksCount, backlink)

				// Get the previous page rank value for the backlink
				backlinkRank, ok := pageRank[backlink]
				if !ok {
					// fmt.Printf("No page rank found for backlink %s\n", backlink)
					continue
				}
				// fmt.Printf("\t\t\tBacklink Page Rank: %f\n", backlinkRank)

				newCumulativeRank += backlinkRank / float64(outlinksCount)
			}

			damping := 0.85
			newPageRank[url] = (1-damping)/float64(count) + damping*newCumulativeRank
			fmt.Println()
		}

		// Update the page rank values
		pageRank = newPageRank

		// Print the new page rank values
		fmt.Println("New Page Rank values:")
		for url, rank := range pageRank {
			fmt.Printf("Page URL: %s, Page Rank: %f\n", url, rank)
		}
		fmt.Println("--------------------------------------------------")
	}

	// Sort the page rank values by rank
	// Create a slice to hold the page rank values
	type PageRank struct {
		URL  string
		Rank float64
	}
	var pageRanks []PageRank
	for url, rank := range pageRank {
		pageRanks = append(pageRanks, PageRank{URL: url, Rank: rank})
	}
	// Sort the page ranks by rank
	sort.Slice(pageRanks, func(i, j int) bool {
		return pageRanks[i].Rank > pageRanks[j].Rank
	})

	// Print the sorted page rank values
	fmt.Println("Sorted Page Rank values:")
	for _, pageRank := range pageRanks {
		fmt.Printf("Page URL: %s, Page Rank: %f\n", pageRank.URL, pageRank.Rank)
	}

	// Save the page rank values to the database
	for _, pageRank := range pageRanks {
		_, err := db.Collection("pagerank").InsertOne(context.TODO(), bson.D{
			{Key: "_id", Value: pageRank.URL},
			{Key: "rank", Value: pageRank.Rank},
		})
		if err != nil {
			panic(fmt.Sprintf("Could not insert page rank value: %v", err))
		}
	}
	fmt.Println("Page rank values saved to the database!")
}

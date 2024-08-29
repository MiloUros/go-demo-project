package main

import (
	"context"
	"database/sql"
	"github.com/MiloUros/go-demo-project/internal/database"
	"github.com/google/uuid"
	"log"
	"strings"
	"sync"
	"time"
)

func startScraping(
	db *database.Queries,
	concurrency int,
	timeBetweenRequests time.Duration,
) {
	log.Printf("Scraping on %v gorutines every %s durtion", concurrency, timeBetweenRequests)
	ticker := time.NewTicker(timeBetweenRequests)

	for ; ; <-ticker.C {
		feed, err := db.GetNextFeedsToFetch(
			context.Background(),
			int32(concurrency),
		)

		if err != nil {
			log.Println("error fetching feeds:", err)
			continue
		}

		wg := sync.WaitGroup{}
		for _, feed := range feed {
			wg.Add(1)

			go scrapeFeed(db, &wg, feed)
		}
		wg.Wait()
	}
}

func scrapeFeed(db *database.Queries, wg *sync.WaitGroup, feed database.Feed) {
	defer wg.Done()

	_, err := db.MarkFeedAsFetched(context.Background(), feed.ID)
	if err != nil {
		log.Println("error fetching feed as fetched:", err)
		return
	}

	rssFeed, err := urlToFeed(feed.Url)
	if err != nil {
		log.Println("error parsing feed", err)
		return
	}

	for _, item := range rssFeed.Chanel.Item {

		description := sql.NullString{}
		if item.Description != "" {
			description = sql.NullString{String: item.Description, Valid: true}
		}
		pubAt, err := time.Parse(time.RFC1123Z, item.PubDate)
		if err != nil {
			log.Printf("couldn't parse date %v with err %v\n", item.PubDate, err)
			continue
		}
		_, err = db.CreatePost(context.Background(), database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
			Title:       item.Title,
			Description: description,
			PublishedAt: pubAt,
			Url:         item.Link,
			FeedID:      feed.ID,
		})
		if err != nil {
			if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
				continue
			}
			log.Println("failed to create post with error:", err)
		}
	}
	log.Printf("Feed %s collected %v posts found\n", feed.Name, len(rssFeed.Chanel.Item))
}

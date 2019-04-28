package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/coreos/pkg/flagutil"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	_ "github.com/lib/pq"
)

func main() {
	flags := flag.NewFlagSet("user-auth", flag.ExitOnError)
	consumerKey := flags.String("consumer-key", "", "Twitter Consumer Key")
	consumerSecret := flags.String("consumer-secret", "", "Twitter Consumer Secret")
	accessToken := flags.String("access-token", "", "Twitter Access Token")
	accessSecret := flags.String("access-secret", "", "Twitter Access Secret")
	dbURL := flags.String("db-url", "", "PostgreSQL database URL")

	flags.Parse(os.Args[1:])
	flagutil.SetFlagsFromEnv(flags, "TWITTER")

	if *consumerKey == "" || *consumerSecret == "" || *accessToken == "" || *accessSecret == "" {
		log.Fatal("Consumer key/secret and Access token/secret required")
	}

	if *dbURL == "" {
		log.Fatal("Database URL required")
	}

	db, err := sql.Open("postgres", *dbURL)
	if err != nil {
		log.Fatal(err)
	}

	config := oauth1.NewConfig(*consumerKey, *consumerSecret)
	token := oauth1.NewToken(*accessToken, *accessSecret)
	// OAuth1 http.Client will automatically authorize Requests
	httpClient := config.Client(oauth1.NoContext, token)
	client := twitter.NewClient(httpClient)

	demux := twitter.NewSwitchDemux()
	demux.Tweet = func(tweet *twitter.Tweet) {
		var text string
		if tweet.ExtendedTweet != nil {
			text = tweet.ExtendedTweet.FullText
			fmt.Println(">>", tweet.ID, tweet.CreatedAt, tweet.ExtendedTweet.FullText)
		} else {
			text = tweet.Text
			fmt.Println(">", tweet.ID, tweet.CreatedAt, tweet.Text)
		}

		createdAt, err := tweet.CreatedAtTime()
		if err != nil {
			fmt.Println("Cannot parse", tweet.CreatedAt, "into Time object")
		}
		_, err = db.Exec(
			`INSERT INTO twitter_tweet(id, created_at, text, screename)
			VALUES($1, $2, $3, $4)`,
			tweet.ID, createdAt, text, tweet.User.ScreenName)

		if err != nil {
			fmt.Println("Failed to insert tweet:", err)
		}
	}

	filterParams := &twitter.StreamFilterParams{
		Track:         []string{"rustlang", "golang", "python", "cat", "dog"},
		StallWarnings: twitter.Bool(true),
	}

	stream, err := client.Streams.Filter(filterParams)
	if err != nil {
		log.Fatal(err)
	}

	go demux.HandleChan(stream.Messages)

	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch

	fmt.Println("\nStopping Stream...")
	stream.Stop()
}

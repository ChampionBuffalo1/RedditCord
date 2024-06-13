package reddit

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func FetchRedditPost(subreddit string, c chan *RedditFetchResult) {
	defer close(c)
	resp, err := http.Get(fmt.Sprintf("https://www.reddit.com/r/%s.json", subreddit))

	if err != nil {
		c <- &RedditFetchResult{Error: fmt.Errorf("HTTP GET failed: %v", err)}
		return
	}

	if resp.StatusCode != 200 {
		c <- &RedditFetchResult{Error: errors.New("non 200 http status code")}
		return
	}

	if !strings.HasPrefix(resp.Header.Get("Content-Type"), "application/json") {
		c <- &RedditFetchResult{Error: fmt.Errorf("invalid Content-Type Header: %s", resp.Header.Get("Content-Type"))}
		return
	}

	defer resp.Body.Close()
	var redditResp RedditResponse

	err = json.NewDecoder(resp.Body).Decode(&redditResp)
	if err != nil {
		var unmarshalTypeErr *json.UnmarshalTypeError
		if errors.As(err, &unmarshalTypeErr) {
			c <- &RedditFetchResult{Error: fmt.Errorf("JSON unmarshal type error: %v", unmarshalTypeErr)}
			return
		}
		c <- &RedditFetchResult{Error: fmt.Errorf("unknown error during json decode: %v", err)}
		return
	}
	c <- &RedditFetchResult{Data: redditResp}
}

func GetSubreddits(value string, c chan []*discordgo.ApplicationCommandOptionChoice) {
	defer close(c)
	resp, err := http.Get(fmt.Sprintf("https://www.reddit.com/subreddits/search.json?q=%s", value))
	emptyResponse := make([]*discordgo.ApplicationCommandOptionChoice, 0)

	if err != nil {
		log.Printf("HTTP GET /subreddits/search failed: %v", err)
		c <- emptyResponse
		return
	}
	if resp.StatusCode != 200 {
		// No logs for non-200 statusCode
		c <- emptyResponse
		return
	}
	if !strings.HasPrefix(resp.Header.Get("Content-Type"), "application/json") {
		log.Printf("invalid Content-Type Header: %s", resp.Header.Get("Content-Type"))
		c <- emptyResponse
		return
	}

	defer resp.Body.Close()
	var subredditResponse RedditSearchResponse

	err = json.NewDecoder(resp.Body).Decode(&subredditResponse)
	if err != nil {
		var unmarshalTypeErr *json.UnmarshalTypeError
		if errors.As(err, &unmarshalTypeErr) {
			log.Printf("JSON unmarshal type error: %v", unmarshalTypeErr)
			return
		}
		log.Printf("unknown error during json decode: %v", err)
		return
	}
	choices := make([]*discordgo.ApplicationCommandOptionChoice, len(subredditResponse.Data.Children))

	for i, subreddit := range subredditResponse.Data.Children {
		choices[i] = &discordgo.ApplicationCommandOptionChoice{
			Name:  subreddit.Data.DisplayNamePrefixed,
			Value: subreddit.Data.DisplayName,
		}
	}

	c <- choices
}

type RedditChildData struct {
	Kind string `json:"kind"`
	Data struct {
		AuthorFullname        string  `json:"author_fullname"`
		Title                 string  `json:"title"`
		PermaLink             string  `json:"permalink"`
		SubredditNamePrefixed string  `json:"subreddit_name_prefixed"`
		UpvoteCount           int32   `json:"ups"`
		ThumbnailHeight       int32   `json:"thumbnail_height"`
		Name                  string  `json:"name"`
		UpvoteRatio           float64 `json:"upvote_ratio"`
		ViewCount             any     `json:"view_count"`
		CreatedAt             float64 `json:"created_utc"`
		ID                    string  `json:"id"`
		Author                string  `json:"author"`
		NumComments           int32   `json:"num_comments"`
		URL                   string  `json:"url"`
		NumCrossposts         int32   `json:"num_crossposts"`
		IsVideo               bool    `json:"is_video"`
		Gallery               struct {
			Items []any `json:"items"`
		} `json:"gallery_data"`
	} `json:"data"`
}

type RedditResponse struct {
	Data struct {
		After    string            `json:"after"`
		Children []RedditChildData `json:"children"`
	} `json:"data"`
}

type RedditSearchResponse struct {
	Data struct {
		Children []struct {
			Data struct {
				DisplayNamePrefixed string `json:"display_name_prefixed"`
				DisplayName         string `json:"display_name"`
				Title               string `json:"title"`
			} `json:"data"`
		} `json:"children"`
	} `json:"data"`
}

type RedditFetchResult struct {
	Data  RedditResponse
	Error error
}

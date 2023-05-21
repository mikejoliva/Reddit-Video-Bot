package main

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	CLIENT_ID       string = ""
	CLIENT_SECRET   string = ""
	USER_AGENT_NAME string = "RedditVideoBot/0.0.1"
)

type PostCollection struct {
	Subreddit string
	Posts     []PostData
}

type PostData struct {
	Title    string
	IsImage  bool
	MediaURL string
}

func LoadSubredditList(file string) ([]string, error) {
	readFile, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer readFile.Close()

	var subreddits []string
	scanner := bufio.NewScanner(readFile)
	for scanner.Scan() {
		subreddits = append(subreddits, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return subreddits, nil
}

func GetSubredditPosts(token string, subreddit string, timeframe string, limit int) (*PostCollection, error) {
	type ResponseJsonItem struct {
		Title    string `json:"title"`
		NSFW     bool   `json:"over_18"`
		PostHint string `json:"post_hint"`
		MediaURL string `json:"url_overridden_by_dest"`
	}

	type ResponseJsonChildren struct {
		Kind string           `json:"kind"`
		Data ResponseJsonItem `json:"data"`
	}

	type ResponseJsonData struct {
		Dist     int                    `json:"dist"`
		Children []ResponseJsonChildren `json:"children"`
	}

	type ResponseJson struct {
		Kind string           `json:"kind"`
		Data ResponseJsonData `json:"data"`
	}

	address := fmt.Sprintf("https://oauth.reddit.com/r/%s/top/?t=%s&limit=%d", subreddit, timeframe, 50)
	request, err := http.NewRequest("GET", address, strings.NewReader(""))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("device_id", "DO_NOT_TRACK_THIS_DEVICE")
	request.Header.Set("User-Agent", USER_AGENT_NAME)
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	client := &http.Client{
		Timeout: time.Second * 10,
	}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("Unexpected return code: %d for address: %s", response.StatusCode, address))
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var res ResponseJson
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}

	if len(res.Data.Children) == 0 {
		return nil, errors.New(fmt.Sprintf("Found 0 posts for address: %s", address))
	}

	var posts []PostData
	for idx := 0; idx < len(res.Data.Children); idx++ {
		if res.Data.Children[idx].Data.NSFW || strings.Contains(res.Data.Children[idx].Data.PostHint, "video") {
			continue
		}

		// TODO: Update ffmpeg to work with gifs
		if strings.HasSuffix(res.Data.Children[idx].Data.MediaURL, ".gif") {
			continue
		}

		if strings.Contains(res.Data.Children[idx].Data.MediaURL, "gallery") {
			continue
		}

		if len(res.Data.Children[idx].Data.Title) > 120 {
			continue
		}

		posts = append(posts, PostData{
			Title:    res.Data.Children[idx].Data.Title,
			IsImage:  res.Data.Children[idx].Data.PostHint == "image",
			MediaURL: res.Data.Children[idx].Data.MediaURL,
		})

		if len(posts) >= limit {
			break
		}
	}

	if len(posts) == 0 {
		return nil, errors.New(fmt.Sprintf("Found 0 appropriate posts for address: %s", address))
	}

	postCollection := &PostCollection{
		Subreddit: subreddit,
		Posts:     posts,
	}

	return postCollection, nil
}

func GetClientAccessToken() (string, error) {
	type ResponseJson struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
		Scope       string `json:"scope"`
	}

	response, err := DoClientLogin()
	if err != nil {
		return "", err
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	var res ResponseJson
	err = json.Unmarshal([]byte(body), &res)
	if err != nil {
		return "", err
	}

	return res.AccessToken, nil
}

func DoClientLogin() (*http.Response, error) {
	authenticationURL := "https://www.reddit.com/api/v1/access_token"
	authenticationGrant := "grant_type=client_credentials"

	secret := []byte(fmt.Sprintf("%s:%s", CLIENT_ID, CLIENT_SECRET))
	request, err := http.NewRequest("POST", authenticationURL, strings.NewReader(authenticationGrant))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("device_id", "DO_NOT_TRACK_THIS_DEVICE")
	request.Header.Set("User-Agent", USER_AGENT_NAME)
	request.Header.Set("Authorization", fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString(secret)))

	client := &http.Client{
		Timeout: time.Second * 10,
	}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("Unexpected return code: %d for address %s", response.StatusCode, authenticationURL))
	}

	return response, nil
}

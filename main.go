package main

import (
	"fmt"
)

const (
	NUMBER_OF_POSTS = 8
)

func main() {
	token, err := GetClientAccessToken()
	if err != nil {
		fmt.Print(err)
		return
	}

	subreddits, err := LoadSubredditList("./subreddits.txt")
	if err != nil {
		fmt.Println("Failed to load subreddit list!")
		fmt.Print(err)
		return
	}

	fmt.Println(subreddits[0])

	postCollection, err := GetSubredditPosts(token, "memes", "day", NUMBER_OF_POSTS)
	if err != nil {
		fmt.Println(fmt.Sprintf("Failed to get posts for subreddit: %s!", subreddits[0]))
		fmt.Print(err)
		return
	}

	err = GenerateTTS(postCollection)
	if err != nil {
		fmt.Println("Failed to generate TTS!")
		fmt.Print(err)
		return
	}

	err = GenerateVideo(postCollection)
	if err != nil {
		fmt.Println("Failed to generate video!")
		fmt.Print(err)
		return
	}
}

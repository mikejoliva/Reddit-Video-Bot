package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"
)

const (
	TTS_SERVICE_ADDRESS = "http://localhost"
	TTS_SERVICE_PORT    = 5500
	TTS_API_ENDPOINT    = "/api/tts"
	VOICE               = "coqui-tts:en_vctk#100"
)

func GenerateTTS(postCollection *PostCollection) error {
	for i, post := range postCollection.Posts {
		sentence := post.Title
		query := fmt.Sprintf("voice=%s&text=%s&cache=false", url.QueryEscape(VOICE), url.QueryEscape(sentence))
		endpoint := fmt.Sprintf("%s:%d%s?%s", TTS_SERVICE_ADDRESS, TTS_SERVICE_PORT, TTS_API_ENDPOINT, query)
		request, err := http.NewRequest("GET", endpoint, strings.NewReader(""))
		if err != nil {
			return err
		}
		request.Header.Set("accept", "*/*")

		client := &http.Client{
			Timeout: time.Second * 10,
		}
		response, err := client.Do(request)
		if err != nil {
			return err
		}
		if response.StatusCode != 200 {
			return errors.New(fmt.Sprintf("Unexpected return code: %d for address %s", response.StatusCode, endpoint))
		}

		directory, err := GetAssetsDirectroryForSubreddit(postCollection.Subreddit)
		if err != nil {
			return err
		}
		filename := path.Join(directory, fmt.Sprintf("%s.wav", fmt.Sprint(i)))
		file, err := os.Create(filename)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(file, response.Body)
		if err != nil {
			return err
		}
	}

	return nil
}

func strip(input string) string {
	var result strings.Builder
	for i := 0; i < len(input); i++ {
		char := input[i]
		if ('a' <= char && char <= 'z') || ('A' <= char && char <= 'Z') || char == ' ' {
			result.WriteByte(char)
		}
	}
	return result.String()
}

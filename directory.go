package main

import (
	"os"
	"path"
)

const (
	ASSETS_FOLDER = "assets"
	VIDEOS_FOLDER = "videos"
)

func GetAssetsDirectrory() (string, error) {
	return getDirectory(ASSETS_FOLDER)
}

func GetAssetsDirectroryForSubreddit(subreddit string) (string, error) {
	return getDirectory(path.Join(ASSETS_FOLDER, subreddit))
}

func GetVideosDirectrory() (string, error) {
	return getDirectory(VIDEOS_FOLDER)
}

func GetVideosDirectroryForSubreddit(subreddit string) (string, error) {
	return getDirectory(path.Join(VIDEOS_FOLDER, subreddit))
}

func getDirectory(sub string) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	dir := path.Join(cwd, sub)
	err = os.MkdirAll(dir, 0700)
	if err != nil {
		return "", err
	}

	return dir, nil
}

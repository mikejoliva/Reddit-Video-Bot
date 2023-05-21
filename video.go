package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	ffprobe "github.com/vansante/go-ffprobe"
)

const (
	VIDEO_LENGTH int = 50
)

func GenerateVideo(postCollection *PostCollection) error {
	background, err := getRandomBackgroundCut(postCollection.Subreddit)
	if err != nil {
		return err
	}

	err = downloadImageCollection(postCollection)
	if err != nil {
		return err
	}

	err = addImagesToVideo(background, postCollection)
	if err != nil {
		return err
	}

	return nil
}

func addImagesToVideo(background string, postCollection *PostCollection) error {
	base := background
	duration := VIDEO_LENGTH / len(postCollection.Posts)

	videoDir, err := GetVideosDirectrory()
	if err != nil {
		return err
	}
	assetsForSubredditDir, err := GetAssetsDirectroryForSubreddit(postCollection.Subreddit)
	if err != nil {
		return err
	}

	for idx := 0; idx < len(postCollection.Posts); idx++ {
		image := path.Join(assetsForSubredditDir, fmt.Sprintf("%d.%s", idx, getFiletype(postCollection.Posts[idx].MediaURL)))

		output := path.Join(assetsForSubredditDir, fmt.Sprintf("%d.mp4", idx))
		if idx == len(postCollection.Posts)-1 {
			output = path.Join(videoDir, fmt.Sprintf("%s.mp4", postCollection.Subreddit))
		}

		offset := idx * duration
		end := offset + duration

		audio := path.Join(assetsForSubredditDir, fmt.Sprintf("%d.wav", idx))
		err = addOVerlayAndAudioToStream(base, image, audio, fmt.Sprint(offset), fmt.Sprint(end), output)
		if err != nil {
			return err
		}

		err = os.Remove(image)
		if err != nil {
			return err
		}
		err = os.Remove(base)
		if err != nil {
			return err
		}
		err = os.Remove(audio)
		if err != nil {
			return err
		}

		base = output
	}

	return nil
}

func getRandomBackgroundCut(name string) (string, error) {
	assetsDir, err := GetAssetsDirectrory()
	if err != nil {
		return "", err
	}
	background := path.Join(assetsDir, "background.mp4")
	data, err := ffprobe.GetProbeData(background, 5*time.Second)
	if err != nil {
		return "", err
	}

	fmt.Println()
	seed := rand.NewSource(time.Now().UnixNano())
	random := rand.New(seed)
	startTime := random.Intn(int(data.Format.DurationSeconds - float64(VIDEO_LENGTH+1)))

	videosDir, err := GetVideosDirectrory()
	if err != nil {
		return "", err
	}

	filename := path.Join(videosDir, fmt.Sprintf("%s.mp4", name))
	err = cutBackgroundVideoCommand(background, startTime, filename)
	if err != nil {
		return "", err
	}

	return filename, nil
}

func downloadImageCollection(postCollection *PostCollection) error {
	assetsForSubredditDir, err := GetAssetsDirectroryForSubreddit(postCollection.Subreddit)
	if err != nil {
		return err
	}

	for idx := 0; idx < len(postCollection.Posts); idx++ {
		if !postCollection.Posts[idx].IsImage {
			continue
		}

		filename := fmt.Sprintf("temp-%d.%s", idx, getFiletype(postCollection.Posts[idx].MediaURL))
		err = downloadImage(postCollection.Posts[idx].MediaURL, assetsForSubredditDir, filename)
		if err != nil {
			return err
		}
	}

	return nil
}

func getFiletype(file string) string {
	parts := strings.Split(file, ".")
	return parts[len(parts)-1]
}

func downloadImage(URL string, directory string, name string) error {
	response, err := http.Get(URL)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	filepath := path.Join(directory, name)
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, response.Body)
	if err != nil {
		return err
	}

	output := path.Join(directory, strings.TrimPrefix(name, "temp-"))
	err = resizeImageCommand(filepath, output)
	if err != nil {
		return err
	}

	err = os.Remove(filepath)
	if err != nil {
		return err
	}

	return nil
}

func resizeImageCommand(input string, output string) error {
	cmd := exec.Command(
		"ffmpeg", "-y",
		"-i", input,
		"-vf", "scale=1030:-2",
		output)
	return runCommand(cmd)
}

func cutBackgroundVideoCommand(background string, start int, output string) error {
	// Add silent audio to background for audio mixing with TTS
	cmd := exec.Command(
		"ffmpeg", "-y",
		"-f", "lavfi",
		"-i", "anullsrc=channel_layout=stereo:sample_rate=44100",
		"-i", background,
		"-ss", fmt.Sprint(start),
		"-t", fmt.Sprint(VIDEO_LENGTH),
		"-c:v", "copy",
		"-c:a", "aac",
		output)
	return runCommand(cmd)
}

func addOVerlayAndAudioToStream(input string, image string, audio string, start string, end string, output string) error {
	startInt, err := strconv.Atoi(start)
	if err != nil {
		return err
	}
	startMilliseconds := startInt * 1000

	cmd := exec.Command(
		"ffmpeg", "-y",
		"-i", input,
		"-i", image,
		"-i", audio,
		"-filter_complex", fmt.Sprintf("[0]overlay=25:25:enable='between(t,%s,%s)'[out];[2]adelay=%d|%d[aud];[0][aud]amix=inputs=2:normalize=0", start, end, startMilliseconds, startMilliseconds),
		"-map", "[out]",
		"-map", "2:a",
		output)

	return runCommand(cmd)
}

func runCommand(command *exec.Cmd) error {
	fmt.Println(fmt.Sprintf("Running command: %s", command.String()))
	var stdout = bytes.Buffer{}
	var stderr = bytes.Buffer{}
	command.Stdout = &stdout
	command.Stderr = &stderr
	err := command.Run()
	if err != nil {
		return errors.New(stderr.String())
	}
	return nil
}

# Reddit video bot for generating YouTube Shorts videos

>I don't know why I wrote this, please see it in action [here](https://youtube.com/shorts/yfFaYC9PdHA).

Automatically generate YouTube Short videos with text-to-speech (TTS) voiceover for a given subreddit.

This program does the following:
1. Fetches the top posts of the day for a given subreddit (ignoring any NSFW flagged posts).
1. Send the title of the post to a TTS server.
1. Overlay the post image with the TTS voiceover.
1. Steps `2` & `3` are repeated for a given number of posts.
1. Save the generated video under the `./videos` directory.
This program will **not** automatically upload the video to YouTube.

## Building this project

> You will need a [coqui-ai](https://github.com/coqui-ai/TTS) TTS server running, the connection endpoints can be configured `tts.go`.

1. Ensure you have a TTS sever running at the specified endpoint in `tts.go`.
1. Set your Reddit developer `CLIENT_ID` and `CLIENT_SECRET` in `reddit.go`.
1. Supply a background video to be drawn behind the Reddit images (might I recommend some Subway Surfers gameplay).
1. This program will run `ffmpeg` on the `shell` of the running device, ensure it is installed and accessible by the current user.
1. Compile and run by running: `make run`
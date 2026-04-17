package main

import (
	"fmt"
	"log"
	"os"

	"github.com/huginmost/bili-music-player/bili"
)

const (
	defaultBVID   = "BV1oU1jBXEN8"
	defaultPIPath = "pi.json"
	defaultISPath = "is.json"
	defaultAudio  = "test.m4a"
)

func main() {
	cookie := os.Getenv("BILI_COOKIE")

	client, err := bili.New(cookie)
	if err != nil {
		log.Fatalf("bili_init failed: %v", err)
	}

	ok := client.Try()
	fmt.Println(ok)
	if !ok {
		return
	}

	if _, err := client.GetPlayInfo(defaultBVID, defaultPIPath); err != nil {
		log.Fatalf("bili_get_pi failed: %v", err)
	}

	if _, err := client.GetInitialState(defaultBVID, defaultISPath); err != nil {
		log.Fatalf("bili_get_is failed: %v", err)
	}

	title, err := client.GetUGCSeasonTitle()
	if err != nil {
		log.Fatalf("get title failed: %v", err)
	}
	fmt.Println(title)

	if _, err := client.GetBMPInfo(); err != nil {
		log.Fatalf("bili_get_bmpinfo failed: %v", err)
	}

	audioURL, err := client.GetAudio()
	if err != nil {
		log.Fatalf("bili_get_audio failed: %v", err)
	}
	fmt.Println(audioURL)

	if err := client.AudioDownload(audioURL, defaultAudio); err != nil {
		log.Fatalf("bili_audio_download failed: %v", err)
	}
}

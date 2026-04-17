package main

import (
	"fmt"
	"log"
	"os"

	"github.com/huginmost/bili-music-player/bili"
)

const defaultBVID = "BV1oU1jBXEN8"

func main() {
	client, err := bili.New(os.Getenv("BILI_COOKIE"))
	if err != nil {
		log.Fatalf("bili_init failed: %v", err)
	}

	ok := client.Try()
	fmt.Println(ok)
	if !ok {
		return
	}

	if _, err := client.GetPlayInfo(defaultBVID, bili.PlayInfoPath); err != nil {
		log.Fatalf("bili_get_pi failed: %v", err)
	}

	if _, err := client.GetInitialState(defaultBVID, bili.InitialStatePath); err != nil {
		log.Fatalf("bili_get_is failed: %v", err)
	}

	title, err := client.GetUGCSeasonTitle()
	if err != nil {
		log.Fatalf("GetUGCSeasonTitle failed: %v", err)
	}
	fmt.Println(title)

	if _, err := client.GetBMPInfo(); err != nil {
		log.Fatalf("bili_get_bmpinfo failed: %v", err)
	}

	if err := client.FixBMPInfo(""); err != nil {
		log.Fatalf("bili_bmpinfo_fix failed: %v", err)
	}
}

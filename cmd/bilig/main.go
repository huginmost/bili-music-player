package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/huginmost/bili-music-player/bili"
)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		log.Fatal(err)
	}
}

var newClient = func() (*bili.Bili, error) {
	return bili.New(os.Getenv("BILI_COOKIE"))
}

func run(args []string, stdout io.Writer) error {
	client, err := newClient()
	if err != nil {
		return fmt.Errorf("bili_init failed: %w", err)
	}

	if len(args) == 0 {
		return nil
	}

	switch args[0] {
	case "-try":
		fmt.Fprintln(stdout, client.Try())
		return nil
	case "-get":
		if len(args) < 2 {
			return fmt.Errorf("usage: bilig -get [bv]")
		}
		if _, err := client.GetPlayInfo(args[1], bili.PlayInfoPath); err != nil {
			return fmt.Errorf("bili_get_pi failed: %w", err)
		}
		if _, err := client.GetInitialState(args[1], bili.InitialStatePath); err != nil {
			return fmt.Errorf("bili_get_is failed: %w", err)
		}
		if _, err := client.GetBMPInfo(); err != nil {
			return fmt.Errorf("bili_get_bmpinfo failed: %w", err)
		}
		return nil
	case "-lget":
		if len(args) < 2 {
			return fmt.Errorf("usage: bilig -lget [ml]")
		}
		if _, err := client.GetListInitialState(args[1], bili.InitialStatePath); err != nil {
			return fmt.Errorf("bili_lget_is failed: %w", err)
		}
		if _, err := client.GetListBMPInfo(); err != nil {
			return fmt.Errorf("bili_lget_bmpinfo failed: %w", err)
		}
		return nil
	case "--title":
		title, err := client.GetUGCSeasonTitle()
		if err != nil {
			return fmt.Errorf("GetUGCSeasonTitle failed: %w", err)
		}
		fmt.Fprintln(stdout, title)
		return nil
	case "-fix":
		if len(args) < 2 {
			return fmt.Errorf("usage: bilig -fix [bv]")
		}
		if err := client.FixBMPInfo(args[1]); err != nil {
			return fmt.Errorf("bili_bmpinfo_fix failed: %w", err)
		}
		return nil
	case "-fix--all":
		if err := client.FixBMPInfo(""); err != nil {
			return fmt.Errorf("bili_bmpinfo_fix failed: %w", err)
		}
		return nil
	case "-download":
		if len(args) < 3 {
			return fmt.Errorf("usage: bilig -download [url] [title]")
		}
		if err := client.AudioDownload(args[1], args[2]); err != nil {
			return fmt.Errorf("bili_audio_download failed: %w", err)
		}
		return nil
	case "-del":
		if len(args) < 2 {
			return fmt.Errorf("usage: bilig -del [title]")
		}
		if err := client.DeleteTitle(args[1]); err != nil {
			return fmt.Errorf("bili_del failed: %w", err)
		}
		return nil
	default:
		return fmt.Errorf("unknown command: %s", args[0])
	}
}

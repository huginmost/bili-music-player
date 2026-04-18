package bili

import (
	"fmt"
	"time"
)

func (b *Bili) fillBMPInfoAudio(item *BMPInfoItem) error {
	if _, err := b.GetPlayInfo(item.BVID, PlayInfoPath); err != nil {
		return err
	}

	audioURL, err := b.GetAudio()
	if err != nil {
		return err
	}

	item.Audio = audioURL
	return nil
}

// FixBMPInfo fills the audio field in bmpinfo.json for one matching BV or for all items.
func (b *Bili) FixBMPInfo(bv string) error {
	payload, err := b.ReadBMPInfo()
	if err != nil {
		return err
	}

	if bv != "" {
		for title := range payload {
			for i := range payload[title] {
				if payload[title][i].BVID != bv {
					continue
				}

				if err := b.fillBMPInfoAudio(&payload[title][i]); err != nil {
					return err
				}

				return b.writeBMPInfo(payload)
			}
		}

		return fmt.Errorf("bvid %s not found in bmpinfo.json", bv)
	}

	first := true
	for title := range payload {
		for i := range payload[title] {
			if !first {
				time.Sleep(2 * time.Second)
			}
			first = false

			if err := b.fillBMPInfoAudio(&payload[title][i]); err != nil {
				return err
			}
		}
	}

	return b.writeBMPInfo(payload)
}

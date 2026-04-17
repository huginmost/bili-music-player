package bili

import "time"

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
		for i := range payload.PLInfo {
			if payload.PLInfo[i].BVID != bv {
				continue
			}

			if err := b.fillBMPInfoAudio(&payload.PLInfo[i]); err != nil {
				return err
			}

			return b.writeBMPInfo(payload)
		}

		return nil
	}

	for i := range payload.PLInfo {
		if i > 0 {
			time.Sleep(2 * time.Second)
		}

		if err := b.fillBMPInfoAudio(&payload.PLInfo[i]); err != nil {
			return err
		}
	}

	return b.writeBMPInfo(payload)
}

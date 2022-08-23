package bluez

import (
	"github.com/mafik/pulseaudio"
	"github.com/pkg/errors"
)

// AudioProfile stores the audio profile information.
type AudioProfile struct {
	Name        string
	Description string
	Index       uint32
	Active      bool
}

// ListAudioProfiles lists audio profiles of a sound card.
func ListAudioProfiles(deviceAddress string) ([]AudioProfile, error) {
	var profiles []AudioProfile

	client, err := pulseaudio.NewClient()
	if err != nil {
		return nil, err
	}
	defer client.Close()

	cards, err := client.Cards()
	if err != nil {
		return nil, err
	}

	for _, card := range cards {
		if addr, ok := card.PropList["device.string"]; ok {
			if addr != deviceAddress {
				continue
			}

			for profileName, profile := range card.Profiles {
				if profile.Available != 1 {
					continue
				}

				profiles = append(profiles, AudioProfile{
					Index:       card.Index,
					Name:        profileName,
					Description: profile.Description,
					Active:      profile.Name == card.ActiveProfile.Name,
				})
			}

			return profiles, nil
		}
	}

	return nil, errors.New("No profiles found")
}

// SetAudioProfile sets an audio profile for a sound card.
func (a AudioProfile) SetAudioProfile() error {
	client, err := pulseaudio.NewClient()
	if err != nil {
		return err
	}
	defer client.Close()

	return client.SetCardProfile(a.Index, a.Name)
}

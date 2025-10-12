//go:build cli

package cli

import (
	"time"

	"github.com/spf13/cobra"
	"gosper/internal/adapter/outbound/audio"
	"gosper/internal/adapter/outbound/model"
	"gosper/internal/adapter/outbound/storage"
	"gosper/internal/adapter/outbound/whispercpp"
	"gosper/internal/infrastructure/config"
	"gosper/internal/usecase"
)

var recordFlags = struct {
	device   string
	duration time.Duration
	model    string
	lang     string
	out      string
	beep     bool
	outdev   string
	beepvol  float64
}{}

var recordCmd = &cobra.Command{
	Use:   "record",
	Short: "Record from microphone and transcribe",
	RunE: func(cmd *cobra.Command, args []string) error {
		// load config for defaults
		cfg := config.FromEnv()
		if p, err := config.LoadFile(config.DefaultPath()); err == nil {
			if recordFlags.device == "" && p.LastDeviceID != "" {
				recordFlags.device = p.LastDeviceID
			}
			if !recordFlags.beep {
				recordFlags.beep = p.AudioFeedback
			}
			if recordFlags.outdev == "" && p.OutputDeviceID != "" {
				recordFlags.outdev = p.OutputDeviceID
			}
			if recordFlags.beepvol == 0 && p.BeepVolume != 0 {
				recordFlags.beepvol = p.BeepVolume
			}
		}

		uc := &usecase.RecordAndTranscribe{
			Audio: audio.NewInput(),
			Repo:  &model.FSRepo{},
			Trans: &whispercpp.Transcriber{},
			Store: &storage.FS{},
		}
		if recordFlags.beep {
			audio.PlayBeepOptions(audio.BeepOptions{DeviceID: recordFlags.outdev, Volume: float32(recordFlags.beepvol)})
		}
		_, err := uc.Execute(cmd.Context(), usecase.RecordInput{
			DeviceID:  recordFlags.device,
			Duration:  recordFlags.duration,
			ModelName: recordFlags.model,
			Language:  recordFlags.lang,
			OutPath:   recordFlags.out,
		})
		if recordFlags.beep {
			audio.PlayBeepOptions(audio.BeepOptions{DeviceID: recordFlags.outdev, Volume: float32(recordFlags.beepvol)})
		}

		// Persist configuration if any settings were provided
		configNeedsSaving := recordFlags.device != "" || recordFlags.outdev != "" ||
			recordFlags.beepvol != 0 || recordFlags.beep

		if configNeedsSaving || err != nil {
			// Try to load existing config
			p, loadErr := config.LoadFile(config.DefaultPath())
			if loadErr != nil {
				// Config doesn't exist, start with env defaults
				p = cfg
			}

			// Update config with flag values
			if recordFlags.device != "" {
				p.LastDeviceID = recordFlags.device
			}
			p.AudioFeedback = recordFlags.beep
			if recordFlags.outdev != "" {
				p.OutputDeviceID = recordFlags.outdev
			}
			if recordFlags.beepvol != 0 {
				p.BeepVolume = recordFlags.beepvol
			}

			// Save only if we have changes to persist
			if configNeedsSaving {
				_ = config.SaveFile(config.DefaultPath(), p)
			}
		}

		return err
	},
}

func init() {
	rootCmd.AddCommand(recordCmd)
	recordCmd.Flags().StringVar(&recordFlags.device, "device", "", "Device ID")
	recordCmd.Flags().DurationVar(&recordFlags.duration, "duration", 0, "Record duration (e.g., 5s). 0 until Ctrl-C")
	recordCmd.Flags().StringVar(&recordFlags.model, "model", "", "Model name or local path")
	recordCmd.Flags().StringVar(&recordFlags.lang, "lang", "auto", "Language code or 'auto'")
	recordCmd.Flags().StringVarP(&recordFlags.out, "out", "o", "", "Output transcript path (.txt or .json)")
	recordCmd.Flags().BoolVar(&recordFlags.beep, "audio-feedback", false, "Beep on start/stop (console bell)")
	recordCmd.Flags().StringVar(&recordFlags.outdev, "output-device", "", "Output device ID or name for beep")
	recordCmd.Flags().Float64Var(&recordFlags.beepvol, "beep-volume", 0.2, "Beep volume 0..1 (malgo builds)")
}

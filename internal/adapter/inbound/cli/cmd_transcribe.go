//go:build cli

package cli

import (
    "context"
    "fmt"
    "github.com/spf13/cobra"
    "gosper/internal/adapter/outbound/model"
    "gosper/internal/adapter/outbound/storage"
    "gosper/internal/adapter/outbound/whispercpp"
    "gosper/internal/usecase"
)

var transcribeFlags = struct{
    model string
    lang string
    translate bool
    threads uint
    out string
    timestamps bool
    beam int
    maxtokens uint
    prompt string
}{}

var transcribeCmd = &cobra.Command{
    Use:   "transcribe <audiofile>",
    Short: "Transcribe an audio file",
    Args:  cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        ctx := cmd.Context()
        path := args[0]
        uc := &usecase.TranscribeFile{
            Repo: &model.FSRepo{},
            Trans: &whispercpp.Transcriber{},
            Store: storage.FS{},
            Factory: nil, // default decoder.New
        }
        _, err := uc.Execute(ctx, usecase.TranscribeInput{
            Path: path,
            OutPath: transcribeFlags.out,
            ModelName: transcribeFlags.model,
            Language: transcribeFlags.lang,
            Translate: transcribeFlags.translate,
            Threads: transcribeFlags.threads,
            Timestamps: transcribeFlags.timestamps,
            BeamSize: transcribeFlags.beam,
            MaxTokens: transcribeFlags.maxtokens,
            InitialPrompt: transcribeFlags.prompt,
        })
        if err != nil { return fmt.Errorf("transcription failed: %w", err) }
        return nil
    },
}

func init() {
    rootCmd.AddCommand(transcribeCmd)
    transcribeCmd.Flags().StringVar(&transcribeFlags.model, "model", "", "Model name or local path")
    transcribeCmd.Flags().StringVar(&transcribeFlags.lang, "lang", "auto", "Language code or 'auto'")
    transcribeCmd.Flags().BoolVar(&transcribeFlags.translate, "translate", false, "Translate to English")
    transcribeCmd.Flags().UintVar(&transcribeFlags.threads, "threads", 0, "Number of threads to use")
    transcribeCmd.Flags().BoolVar(&transcribeFlags.timestamps, "timestamps", false, "Emit token timestamps")
    transcribeCmd.Flags().IntVar(&transcribeFlags.beam, "beam", 0, "Beam size (0 default)")
    transcribeCmd.Flags().UintVar(&transcribeFlags.maxtokens, "max-tokens", 0, "Max tokens per segment (0 unlimited)")
    transcribeCmd.Flags().StringVar(&transcribeFlags.prompt, "prompt", "", "Initial prompt")
    transcribeCmd.Flags().StringVarP(&transcribeFlags.out, "out", "o", "", "Output transcript path (.txt or .json)")
}


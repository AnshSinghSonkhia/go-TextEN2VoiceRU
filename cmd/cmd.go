package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	speaker "github.com/AnshSinghSonkhia/go-TextEN2VoiceRU/speech"
)

const (
	ExitCodeOK              = 0
	ExitCodeParseFlagsError = 1
	ExitCodeValidateError   = 2
	ExitCodeInternalError   = 3
	ExitCodeOutputError     = 4
)

type CLI struct {
	ErrStream io.Writer
}

// TranslateText translates English text to Russian using the LibreTranslate API.
func TranslateText(text string) (string, error) {
	const translateAPI = "https://libretranslate.com/translate"

	type requestPayload struct {
		Q            string `json:"q"`
		Source       string `json:"source"`
		Target       string `json:"target"`
		Format       string `json:"format"`
		Alternatives int    `json:"alternatives"`
		APIKey       string `json:"api_key"`
	}

	type responsePayload struct {
		TranslatedText string `json:"translatedText"`
	}

	payload := requestPayload{
		Q:            text,
		Source:       "en",
		Target:       "ru",
		Format:       "text",
		Alternatives: 3,
		APIKey:       "", // Use your API key from LibreTranslate: https://portal.libretranslate.com
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(translateAPI, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to translate text, status code: %d", resp.StatusCode)
	}

	var result responsePayload
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.TranslatedText, nil
}

func (cli *CLI) Run(args []string) int {
	flags := flag.NewFlagSet("google-text-to-speech", flag.ContinueOnError)

	var (
		text, voice, out string
		rate, pitch      float64
	)

	flags.StringVar(&text, "text", "", "text to speech")
	flags.StringVar(&voice, "voice", "stand-a", "voice name")
	flags.Float64Var(&rate, "rate", 1.00, "speech rate (0.25 to 4.0)")
	flags.Float64Var(&pitch, "pitch", 0.00, "pitch adjustment (-20.0 to 20.0)")
	flags.StringVar(&out, "o", "", "output audio file (supports format of the audio: LINEAR16, MP3, OGG_OPUS)")

	if err := flags.Parse(args); err != nil {
		fmt.Fprint(cli.ErrStream, "Error parsing flags: ", err)
		return ExitCodeParseFlagsError
	}

	// Translate text from English to Russian
	russianText, err := TranslateText(text)
	if err != nil {
		fmt.Fprint(cli.ErrStream, "Error translating text: ", err)
		return ExitCodeInternalError
	}

	opt, err := makeSpeechOptions(russianText, voice, out, rate, pitch)
	if err != nil {
		fmt.Fprint(cli.ErrStream, "Error validating flags: ", err)
		return ExitCodeValidateError
	}

	ctx := context.Background()
	client, err := speaker.NewSpeechClient(ctx)
	if err != nil {
		fmt.Fprint(cli.ErrStream, "Error creating speech client: ", err)
		return ExitCodeInternalError
	}

	b, err := client.Run(ctx, speaker.NewRequest(russianText, opt))
	if err != nil {
		fmt.Fprint(cli.ErrStream, "Error synthesizing speech: ", err)
		return ExitCodeInternalError
	}

	if err = os.WriteFile(out, b, 0644); err != nil {
		fmt.Fprint(cli.ErrStream, "Error writing output file: ", err)
		return ExitCodeOutputError
	}

	fmt.Fprintln(os.Stdout, "Speech synthesis completed successfully.")
	return ExitCodeOK
}

func makeSpeechOptions(text, voice, out string, rate, pitch float64) (*speaker.SpeechOptions, error) {
	if text == "" {
		return nil, fmt.Errorf("text is required")
	}

	var voiceName string
	switch v := strings.ToLower(voice); v {
	case "stand-a":
		voiceName = speaker.VoiceStandardA
	case "stand-b":
		voiceName = speaker.VoiceStandardB
	case "stand-c":
		voiceName = speaker.VoiceStandardC
	case "stand-d":
		voiceName = speaker.VoiceStandardD
	case "stand-e":
		voiceName = speaker.VoiceStandardE
	case "wave-a":
		voiceName = speaker.VoiceWaveNetA
	case "wave-b":
		voiceName = speaker.VoiceWaveNetB
	case "wave-c":
		voiceName = speaker.VoiceWaveNetC
	case "wave-d":
		voiceName = speaker.VoiceWaveNetD
	case "wave-e":
		voiceName = speaker.VoiceWaveNetE
	default:
		return nil, fmt.Errorf("invalid voice name: %v", v)
	}

	if 0.25 > rate || rate > 4.0 {
		return nil, fmt.Errorf("speech rate must be between 0.25 and 4.0, got: %g", rate)
	}

	if -20.0 > pitch || pitch > 20.0 {
		return nil, fmt.Errorf("pitch adjustment must be between -20.0 and 20.0, got: %g", pitch)
	}

	switch ext := strings.ToLower(filepath.Ext(out)); ext {
	case ".wav":
		return &speaker.SpeechOptions{
			LanguageCode:    "ru-RU",
			VoiceName:       voiceName,
			AudioEncoding:   speaker.AudioEncoding_Linear16,
			AudioSpeechRate: rate,
			AudioPitch:      pitch,
		}, nil
	case ".mp3":
		return &speaker.SpeechOptions{
			LanguageCode:    "ru-RU",
			VoiceName:       voiceName,
			AudioEncoding:   speaker.AudioEncoding_MP3,
			AudioSpeechRate: rate,
			AudioPitch:      pitch,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported output format: %s", out)
	}
}

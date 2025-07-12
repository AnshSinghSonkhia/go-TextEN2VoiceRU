package speaker

import (
	"context"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	texttospeechpb "cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
)

var speaker *Speaker

const (
	// ref https://cloud.google.com/text-to-speech/docs/list-voices-and-types

	VoiceStandardA = "ru-RU-Standard-A"
	VoiceStandardB = "ru-RU-Standard-B"
	VoiceStandardC = "ru-RU-Standard-C"
	VoiceStandardD = "ru-RU-Standard-D"
	VoiceStandardE = "ru-RU-Standard-E"

	VoiceWaveNetA = "ru-RU-Wavenet-A"
	VoiceWaveNetB = "ru-RU-Wavenet-B"
	VoiceWaveNetC = "ru-RU-Wavenet-C"
	VoiceWaveNetD = "ru-RU-Wavenet-D"
	VoiceWaveNetE = "ru-RU-Wavenet-E"

	AudioEncoding_Linear16 = texttospeechpb.AudioEncoding_LINEAR16
	AudioEncoding_MP3      = texttospeechpb.AudioEncoding_MP3
	AudioEncoding_OGG      = texttospeechpb.AudioEncoding_OGG_OPUS
)

type SpeechOptions struct {
	LanguageCode    string
	VoiceName       string
	AudioEncoding   texttospeechpb.AudioEncoding
	AudioSpeechRate float64
	AudioPitch      float64
}

type AudioEncoding texttospeechpb.AudioEncoding

type Speaker struct {
	client *texttospeech.Client
}

func NewSpeechClient(ctx context.Context) (*Speaker, error) {
	if speaker != nil {
		return speaker, nil
	}

	client, err := texttospeech.NewClient(ctx)
	if err != nil {
		return nil, err
	}

	speaker = &Speaker{
		client: client,
	}

	return speaker, nil
}

func NewRequest(text string, opt *SpeechOptions) *texttospeechpb.SynthesizeSpeechRequest {
	if opt == nil {
		opt = &SpeechOptions{
			LanguageCode:  "ru-RU",
			VoiceName:     VoiceStandardA,
			AudioEncoding: AudioEncoding_MP3,
		}
	}

	return &texttospeechpb.SynthesizeSpeechRequest{
		Input: &texttospeechpb.SynthesisInput{
			InputSource: &texttospeechpb.SynthesisInput_Text{Text: text},
		},
		Voice: &texttospeechpb.VoiceSelectionParams{
			LanguageCode: opt.LanguageCode,
			Name:         opt.VoiceName,
			SsmlGender:   texttospeechpb.SsmlVoiceGender_NEUTRAL,
		},
		AudioConfig: &texttospeechpb.AudioConfig{
			AudioEncoding: opt.AudioEncoding,
			SpeakingRate:  opt.AudioSpeechRate,
			Pitch:         opt.AudioPitch,
		},
	}
}

func (s *Speaker) Run(ctx context.Context, req *texttospeechpb.SynthesizeSpeechRequest) ([]byte, error) {
	resp, err := s.client.SynthesizeSpeech(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.AudioContent, nil
}

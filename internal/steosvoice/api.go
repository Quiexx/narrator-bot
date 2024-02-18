package steosvoice

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

const (
	GET_VOICES_URL             = "https://api.voice.steos.io/v1/get/voices"
	GET_SYMBOLS_URL            = "https://api.voice.steos.io/v1/get/symbols"
	GET_TARIFFS_URL            = "https://api.voice.steos.io/v1/get/tariffs"
	GET_SYNTHESIZED_SPEECH_URL = "https://api.voice.steos.io/v1/get/tts"
	FORMAT                     = "mp3"
)

type Voice struct {
	Id          int64             `json:"voice_id,omitempty"`
	Name        map[string]string `json:"name,omitempty"`
	Description map[string]string `json:"description,omitempty"`
	LangId      int64             `json:"id_lang,omitempty"`
	Sex         string            `json:"sex,omitempty"`
}

type GetVoiceResp struct {
	Status  bool     `json:"status,omitempty"`
	Message string   `json:"message,omitempty"`
	Voices  []*Voice `json:"voices,omitempty"`
}

type GetSymbolsResp struct {
	Status  bool   `json:"status,omitempty"`
	Message string `json:"message,omitempty"`
	Symbols int64  `json:"symbols,omitempty"`
}

type GetSynthSpeechResp struct {
	Status   bool   `json:"status,omitempty"`
	Message  string `json:"message,omitempty"`
	VoiceId  int64  `json:"voice_id,omitempty"`
	AudioUrl string `json:"audio_url,omitempty"`
	Format   string `json:"format,omitempty"`
}

type GetSynthSpeechReq struct {
	VoiceId int64  `json:"voice_id,omitempty"`
	Text    string `json:"text,omitempty"`
	Format  string `json:"format,omitempty"`
}

type SteosVoiceAPI struct {
	client *http.Client
}

func NewSteosVoiceAPI(client *http.Client) *SteosVoiceAPI {
	return &SteosVoiceAPI{client: client}
}

func (s *SteosVoiceAPI) GetVoices(apiKey string) (*GetVoiceResp, error) {
	req, err := http.NewRequest("GET", GET_VOICES_URL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	gvrep := &GetVoiceResp{}
	err = json.Unmarshal(body, gvrep)
	if err != nil {
		return nil, err
	}

	return gvrep, nil
}

func (s *SteosVoiceAPI) GetSymbols(apiKey string) (*GetSymbolsResp, error) {
	req, err := http.NewRequest("GET", GET_SYMBOLS_URL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	gsrep := &GetSymbolsResp{}
	err = json.Unmarshal(body, gsrep)
	if err != nil {
		return nil, err
	}

	return gsrep, nil
}

func (s *SteosVoiceAPI) GetSynthesizedSpeech(apiKey string, text string, voiceId int64) (*GetSynthSpeechResp, error) {
	gssBody := &GetSynthSpeechReq{VoiceId: voiceId, Text: text, Format: FORMAT}
	reqBody, err := json.Marshal(gssBody)
	if err != nil {
		return nil, err
	}
	jsonBytes := bytes.NewBuffer(reqBody)

	req, err := http.NewRequest("POST", GET_SYNTHESIZED_SPEECH_URL, jsonBytes)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	gssrep := &GetSynthSpeechResp{}
	err = json.Unmarshal(body, gssrep)
	if err != nil {
		return nil, err
	}

	return gssrep, nil
}

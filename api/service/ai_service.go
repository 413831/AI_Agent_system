package service

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type AIService struct {
	BaseURL string
}

func NewAIService() *AIService {
	return &AIService{
		BaseURL: "http://ai-server:8000",
	}
}

func (a *AIService) Ask(prompt string) (string, error) {
	payload := map[string]string{"prompt": prompt}
	body, _ := json.Marshal(payload)

	resp, err := http.Post(a.BaseURL+"/ask", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var res map[string]string
	json.NewDecoder(resp.Body).Decode(&res)

	return res["result"], nil
}

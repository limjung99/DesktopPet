package main

// ───────────── 사내 LLM API 호출 ─────────────
// 요청: POST llmEndpoint, Content-Type: application/json, Authorization: Bearer {API 키}
//
//	{"model":"Qwen-Inference-Think","messages":[{"role":"user","content":"..."}],"stream":false}
//
// 응답은 OpenAI 호환(choices[0].message.content)과 Ollama(message.content)를 모두 받는다.
// API 키는 첫 실행 때 입력받아 암호화해 둔다(apikey_store.go, apikey_window.go).

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const (
	llmEndpoint = "https://ollama.ai.koscom.co.kr"
	llmModel    = "Qwen-Inference-Think"
	llmTimeout  = 60 * time.Second
)

// 저장된 API 키. 시작할 때 loadAPIKey로 읽고, 키 입력 창에서 저장하면 바뀐다 (UI 스레드에서만 쓴다)
var llmAPIKey string

type llmMessage struct {
	Role    string `json:"role"` // "user" | "assistant"
	Content string `json:"content"`
}

type llmRequest struct {
	Model    string       `json:"model"`
	Messages []llmMessage `json:"messages"`
	Stream   bool         `json:"stream"` // Ollama는 기본이 스트리밍이라 꺼 둔다
}

// OpenAI 호환 choices[0].message.content 또는 Ollama message.content
type llmReply struct {
	Choices []struct {
		Message llmMessage `json:"message"`
	} `json:"choices"`
	Message llmMessage `json:"message"`
}

// 추론(Think) 모델이 답변 앞에 붙이는 생각 과정
var thinkBlock = regexp.MustCompile(`(?s)<think>.*?</think>`)

// 대화 기록으로 LLM의 다음 답변을 받는다
func askLLM(history []chatMessage, apiKey string) (string, error) {
	if apiKey == "" {
		return "", errors.New("API 키가 없어요. 펫을 오른쪽 클릭해서 [API 키 설정]으로 넣어 주세요")
	}
	body, err := buildLLMRequest(history)
	if err != nil {
		return "", err
	}
	respBody, err := postToLLM(body, apiKey)
	if err != nil {
		return "", err
	}
	return parseLLMReply(respBody)
}

// 대화 기록 전체를 messages로 보낸다 (이전 답변은 role=assistant)
func buildLLMRequest(history []chatMessage) ([]byte, error) {
	msgs := make([]llmMessage, 0, len(history))
	for _, m := range history {
		role := "assistant"
		if m.fromUser {
			role = "user"
		}
		msgs = append(msgs, llmMessage{Role: role, Content: m.text})
	}
	return json.Marshal(llmRequest{Model: llmModel, Messages: msgs})
}

func parseLLMReply(body []byte) (string, error) {
	var r llmReply
	if err := json.Unmarshal(body, &r); err != nil {
		return "", fmt.Errorf("응답을 해석하지 못했어요: %w", err)
	}
	content := r.Message.Content
	if len(r.Choices) > 0 {
		content = r.Choices[0].Message.Content
	}
	if answer := stripThinking(content); answer != "" {
		return answer, nil
	}
	return "", errors.New("응답에 답변 내용이 없어요")
}

// <think>...</think> 부분을 지우고 실제 답변만 남긴다
func stripThinking(content string) string {
	content = thinkBlock.ReplaceAllString(content, "")
	if _, after, found := strings.Cut(content, "</think>"); found { // 여는 태그 없이 닫는 태그만 온 경우
		content = after
	}
	return strings.TrimSpace(content)
}

func postToLLM(body []byte, apiKey string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodPost, llmEndpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	client := &http.Client{Timeout: llmTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("서버 오류 (HTTP %d)", resp.StatusCode)
	}
	return data, nil
}

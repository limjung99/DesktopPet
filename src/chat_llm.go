package main

// ───────────── 사내 LLM API 호출 ─────────────
// 엔드포인트와 요청/응답 형식은 아직 확정되지 않았다.
// 스펙이 정해지면 buildLLMRequest, parseLLMReply 두 함수만 고치면 된다.

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

type llmMessage struct {
	Role    string `json:"role"` // "user" | "assistant"
	Content string `json:"content"`
}

// 임시 응답 형식: {"reply":"..."} / {"answer":"..."} / OpenAI 호환 choices[0].message.content
type llmReply struct {
	Reply   string `json:"reply"`
	Answer  string `json:"answer"`
	Choices []struct {
		Message llmMessage `json:"message"`
	} `json:"choices"`
}

// 대화 기록으로 LLM의 다음 답변을 받는다. 엔드포인트가 없으면 데모 응답
func askLLM(history []chatMessage) (string, error) {
	if cfgLLMEndpoint == "" {
		return demoReply(history), nil
	}
	body, err := buildLLMRequest(history)
	if err != nil {
		return "", err
	}
	respBody, err := postToLLM(body)
	if err != nil {
		return "", err
	}
	return parseLLMReply(respBody)
}

func demoReply(history []chatMessage) string {
	last := history[len(history)-1].text
	return "[LLM 미연결] chat.ini에 llm_endpoint를 설정하면 실제 답변을 받아요.\n입력: " + last
}

// 임시 요청 형식: {"messages":[{"role":"user","content":"..."}, ...]}
func buildLLMRequest(history []chatMessage) ([]byte, error) {
	msgs := make([]llmMessage, 0, len(history))
	for _, m := range history {
		role := "assistant"
		if m.fromUser {
			role = "user"
		}
		msgs = append(msgs, llmMessage{Role: role, Content: m.text})
	}
	return json.Marshal(map[string]any{"messages": msgs})
}

func parseLLMReply(body []byte) (string, error) {
	var r llmReply
	if err := json.Unmarshal(body, &r); err != nil {
		return "", fmt.Errorf("응답을 해석하지 못했어요: %w", err)
	}
	switch {
	case r.Reply != "":
		return r.Reply, nil
	case r.Answer != "":
		return r.Answer, nil
	case len(r.Choices) > 0 && r.Choices[0].Message.Content != "":
		return r.Choices[0].Message.Content, nil
	}
	return "", errors.New("응답에 답변 내용이 없어요")
}

func postToLLM(body []byte) ([]byte, error) {
	req, err := http.NewRequest(http.MethodPost, cfgLLMEndpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if cfgLLMAPIKey != "" {
		req.Header.Set("Authorization", "Bearer "+cfgLLMAPIKey)
	}
	client := &http.Client{Timeout: time.Duration(cfgLLMTimeoutSec * float64(time.Second))}
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

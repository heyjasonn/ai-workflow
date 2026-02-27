package llm

import (
    "bytes"
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "net/http"
    "os"
    "time"
)

type OpenAIClient struct {
    APIKey string
    BaseURL string
    Model  string
    HTTP   *http.Client
}

type chatMessage struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

type chatRequest struct {
    Model       string        `json:"model"`
    Messages    []chatMessage `json:"messages"`
    Temperature float64       `json:"temperature"`
}

type chatResponse struct {
    Choices []struct {
        Message chatMessage `json:"message"`
    } `json:"choices"`
}

func (c OpenAIClient) Complete(ctx context.Context, prompt string) (string, error) {
    apiKey := c.APIKey
    if apiKey == "" {
        apiKey = os.Getenv("OPENAI_API_KEY")
    }
    if apiKey == "" {
        return "", errors.New("missing OPENAI_API_KEY")
    }

    baseURL := c.BaseURL
    if baseURL == "" {
        baseURL = os.Getenv("OPENAI_API_BASE")
    }
    if baseURL == "" {
        baseURL = "https://api.openai.com/v1"
    }

    model := c.Model
    if model == "" {
        model = os.Getenv("OPENAI_MODEL")
    }
    if model == "" {
        model = "gpt-4o-mini"
    }

    reqBody := chatRequest{
        Model:       model,
        Temperature: 0,
        Messages: []chatMessage{
            {Role: "developer", Content: "You are a deterministic transformer. Return only valid JSON or a unified diff patch as requested."},
            {Role: "user", Content: prompt},
        },
    }

    payload, err := json.Marshal(reqBody)
    if err != nil {
        return "", err
    }

    httpClient := c.HTTP
    if httpClient == nil {
        httpClient = &http.Client{Timeout: 60 * time.Second}
    }

    req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/chat/completions", bytes.NewReader(payload))
    if err != nil {
        return "", err
    }
    req.Header.Set("Authorization", "Bearer "+apiKey)
    req.Header.Set("Content-Type", "application/json")

    resp, err := httpClient.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    if resp.StatusCode < 200 || resp.StatusCode >= 300 {
        return "", fmt.Errorf("openai API error: %s", resp.Status)
    }

    var parsed chatResponse
    if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
        return "", err
    }
    if len(parsed.Choices) == 0 {
        return "", errors.New("openai API returned no choices")
    }
    return parsed.Choices[0].Message.Content, nil
}

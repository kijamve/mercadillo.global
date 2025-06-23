package H

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// RequestBody representa el cuerpo de la solicitud a la API de Ollama.
type iaRequestBody struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

// ResponseBody representa la estructura de la respuesta esperada.
type iaResponseBody struct {
	Response string `json:"response"`
}

func PromptToIA(prompt string) (string, error) {
	// URL de la API
	url := "http://localhost:11434/api/generate"
	requestBody := iaRequestBody{
		Model:  os.Getenv("IA_MODEL"),
		Prompt: prompt,
		Stream: false,
	}
	body, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	// {"model":"llama3","prompt":"Translate the following text from en to es. Reply to me just the text, without comments or anything additional: The client's DNI is invalid, must be at least 3 characters","stream":false}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		if resp != nil && resp.Body != nil {
			responseData, err := io.ReadAll(resp.Body)
			defer resp.Body.Close()
			if err != nil && len(responseData) > 0 {
				return "", fmt.Errorf("invalid response: %d, response: %s", resp.StatusCode, string(responseData))
			}
		}
		return "", err
	}
	defer resp.Body.Close()

	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("invalid response: %d, response: %s", resp.StatusCode, string(responseData))
	}

	var responseBody iaResponseBody
	if err := json.Unmarshal(responseData, &responseBody); err != nil {
		return "", err
	}

	return responseBody.Response, nil
}

func TranslateTextWithIA(text string, fromLanguage string, toLanguage string) string {
	text = strings.ReplaceAll(text, "_", " ")
	if fromLanguage == toLanguage {
		return text
	}
	prompt := fmt.Sprintf("Translate from %s to %s, keep the upper and lower case of the original text and reply to me just the text without comments or anything additional. Translate the following text: %s", fromLanguage, toLanguage, text)

	result, err := PromptToIA(prompt)
	if err != nil {
		return text
	}
	return result
}

package H

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type SMS struct {
	PhoneNumber string `json:"phone_number"`
	Message     string `json:"message"`
	DeviceID    int    `json:"device_id"`
}

type ResponseMessage struct {
	ID          int          `json:"id"`
	DeviceID    int          `json:"device_id"`
	PhoneNumber string       `json:"phone_number"`
	Message     string       `json:"message"`
	Status      string       `json:"status"`
	Log         []MessageLog `json:"log"`
	CreatedAt   string       `json:"created_at"`
}

type MessageLog struct {
	Status     string `json:"status"`
	OccurredAt string `json:"occurred_at"`
}

func SendSMS(smsList []SMS, token string) *[]ResponseMessage {

	// Codificar los datos de SMS a JSON
	smsData, err := json.Marshal(smsList)
	if err != nil {
		fmt.Println("Error al codificar los datos de SMS a JSON:", err)
		return nil
	}

	if IsEmpty(token) {
		fmt.Println("Falta el token de acceso")
		return nil
	}
	headers := http.Header{}
	headers.Set("Authorization", token)

	// Crear una solicitud HTTP POST con los datos codificados en JSON
	url := "https://smsgateway.me/api/v4/message/send"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(smsData))
	if err != nil {
		fmt.Println("Error al crear la solicitud HTTP POST:", err)
		return nil
	}
	req.Header = headers
	req.Header.Set("Content-Type", "application/json")

	// Enviar la solicitud HTTP y procesar la respuesta
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error al enviar la solicitud HTTP POST:", err)
		return nil
	}
	defer resp.Body.Close()

	var responseMessages []ResponseMessage
	if err := json.NewDecoder(resp.Body).Decode(&responseMessages); err != nil {
		return nil
	}

	return &responseMessages
}

func GetSMSStatus(id int, token string) *ResponseMessage {
	endpoint := fmt.Sprintf("https://smsgateway.me/api/v4/message/%d", id)

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		fmt.Println("Error al enviar la solicitud HTTP POST:", err)
		return nil
	}

	req.Header.Set("Authorization", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error al enviar la solicitud HTTP POST:", err)
		return nil
	}

	defer resp.Body.Close()

	var responseMessage ResponseMessage
	if err := json.NewDecoder(resp.Body).Decode(&responseMessage); err != nil {
		fmt.Println("Error al enviar la solicitud HTTP POST:", err)
		return nil
	}

	return &responseMessage
}

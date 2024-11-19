package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
)

var countCallCreateFunc = 0

func CreateAutoHost(c *fiber.Ctx) error {
	countCallCreateFunc += 1
	var request struct {
		HostCount int `json:"host_count"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid request body")
	}

	authToken, err := loginToZabbix()
	if err != nil {
		log.Printf("Login failed: %v", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Login failed")
	}

	ipAddress := os.Getenv("IP_ADDRESS")

	for i := 1; i <= request.HostCount; i++ {
		hostName := fmt.Sprintf("Havelsan-Host-AutoCreate--%d.%d", countCallCreateFunc, i)
		hostID, err := createHost(authToken, hostName, ipAddress, i)
		if err != nil {
			log.Printf("Create host failed for %s: %v", hostName, err)
			return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("Create host failed for %s", hostName))
		}
		log.Printf("Host created successfully: %s (ID: %s)", hostName, hostID)
	}

	return c.JSON(fiber.Map{"message": "Hosts created successfully"})
}

func loginToZabbix() (string, error) {
	Username := os.Getenv("USERNAME")
	Password := os.Getenv("PASSWORD")
	ipAddress := os.Getenv("IP_ADDRESS")
	zabbixURL := fmt.Sprintf("http://%s/zabbix/api_jsonrpc.php", ipAddress)

	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "user.login",
		"params": map[string]string{
			"username": Username,
			"password": Password,
		},
		"id": 1,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", zabbixURL, bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(responseBody, &result); err != nil {
		return "", err
	}

	if _, ok := result["error"]; ok {
		return "", fmt.Errorf("login error: %v", result["error"])
	}

	authToken, ok := result["result"].(string)
	if !ok {
		return "", fmt.Errorf("unexpected response format")
	}

	return authToken, nil
}

func createHost(authToken, hostName, ipAddress string, numberOfHost int) (string, error) {
	rand.Seed(time.Now().UnixNano())
	randomNumber := rand.Intn(50) + 1
	value := fmt.Sprintf("%s.%d", os.Getenv("MACROS_IP"), randomNumber)
	var macros []map[string]interface{}
	if numberOfHost <= 6 && numberOfHost >= 1 {
		macros = []map[string]interface{}{
			{
				"macro": "{$IP}",
				"value": value,
			},
		}
	} else {
		macros = []map[string]interface{}{}
	}

	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "host.create",
		"params": map[string]interface{}{
			"host": hostName,
			"interfaces": []map[string]interface{}{
				{
					"type":  1,
					"main":  1,
					"useip": 1,
					"ip":    ipAddress,
					"dns":   "",
					"port":  "10050",
				},
			},
			"groups": []map[string]interface{}{
				{
					"groupid": "2",
				},
			},
			"macros": macros,
		},
		"auth": authToken,
		"id":   2,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("http://%s/zabbix/api_jsonrpc.php", ipAddress), bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(responseBody, &result); err != nil {
		return "", err
	}

	if _, ok := result["error"]; ok {
		return "", fmt.Errorf("create host error: %v", result["error"])
	}

	hostIDs, ok := result["result"].(map[string]interface{})["hostids"].([]interface{})
	if !ok || len(hostIDs) == 0 {
		return "", fmt.Errorf("unexpected response format")
	}

	hostID, ok := hostIDs[0].(string)
	if !ok {
		return "", fmt.Errorf("unexpected response format")
	}

	return hostID, nil
}

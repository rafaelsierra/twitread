package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// ObtainBearerToken - Returns a token to be used with Twitter or an error
// Errors from net/http will bubble up
func ObtainBearerToken(apiKey, apiSecret string) (string, error) {
	form := url.Values{}
	form.Set("grant_type", "client_credentials")

	request, err := http.NewRequest("POST", "https://api.twitter.com/oauth2/token", strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	request.SetBasicAuth(apiKey, apiSecret)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(response.Body).Decode(&result)

	if apiErrors, ok := result["errors"]; ok {
		apiErrors := apiErrors.([]interface{})
		if len(apiErrors) == 1 {
			apiErrors := apiErrors[0].(map[string]interface{})
			if message, ok := apiErrors["message"]; ok {
				message := message.(string)
				return "", fmt.Errorf("Error obtaining token: %s", message)
			}
		}

		return "", fmt.Errorf("Failed to obtain token: %v", apiErrors)
	}

	if tokenType, ok := result["token_type"]; !ok || tokenType != "bearer" {
		return "", fmt.Errorf("Received invalid token type: %s", tokenType)
	}

	if accessToken, ok := result["access_token"]; ok {
		return accessToken.(string), nil
	}

	return "", fmt.Errorf("Access token not present in response: %v", result)
}

func main() {
	var apiKey, apiSecret string

	flag.StringVar(&apiKey, "apiKey", "", "Twitter API Key")
	flag.StringVar(&apiSecret, "apiSecret", "", "Twitter API Secret")
	flag.Parse()

	if apiKey == "" || apiSecret == "" {
		flag.Usage()
		return
	}

	bearer, err := ObtainBearerToken(apiKey, apiSecret)
	if err != nil {
		fmt.Println()
		fmt.Println(err.Error())
		return
	}

	fmt.Println("Token:", bearer)

}

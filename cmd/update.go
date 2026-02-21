package cmd

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

const generatedOutboundsPath = "generated-outbounds.json"

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Fetch xray outbounds from 3x-ui panel and print to stdout",
	RunE:  runUpdate,
}

func init() {
	rootCmd.AddCommand(updateCmd)
}

func getToken(baseURL, username, password string, allowInsecure bool) string {
	baseURL = strings.TrimRight(baseURL, "/")
	loginURL := baseURL + "/login"

	client := &http.Client{}
	if allowInsecure {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	form := url.Values{}
	form.Set("username", username)
	form.Set("password", password)

	req, err := http.NewRequest(http.MethodPost, loginURL, strings.NewReader(form.Encode()))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	for _, line := range resp.Header["Set-Cookie"] {
		if strings.HasPrefix(line, "3x-ui=") {
			val := strings.TrimPrefix(line, "3x-ui=")
			if i := strings.Index(val, ";"); i >= 0 {
				val = val[:i]
			}
			return strings.TrimSpace(val)
		}
	}

	fmt.Println("Headers:")
	for k, v := range resp.Header {
		fmt.Printf("  %s: %s\n", k, strings.Join(v, ", "))
	}
	fmt.Println("Body:")
	fmt.Println(string(body))
	panic("3x-ui cookie not found in login response")
}

func xrayClient(allowInsecure bool) *http.Client {
	client := &http.Client{}
	if allowInsecure {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}
	return client
}

func getPanelConfig(baseURL, token string, allowInsecure bool) string {
	baseURL = strings.TrimRight(baseURL, "/")
	panelURL := baseURL + "/panel/xray/"

	req, err := http.NewRequest(http.MethodPost, panelURL, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Cookie", "3x-ui="+token)

	resp, err := xrayClient(allowInsecure).Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var wrapper struct {
		Success bool   `json:"success"`
		Msg     string `json:"msg"`
		Obj     string `json:"obj"`
	}
	if err := json.Unmarshal(body, &wrapper); err != nil {
		fmt.Println("Body:", string(body))
		panic(err)
	}
	if !wrapper.Success {
		fmt.Println("Body:", string(body))
		panic("panel response success=false: " + wrapper.Msg)
	}
	return wrapper.Obj
}

func getOutbounds(baseURL, token string, allowInsecure bool) []interface{} {
	obj := getPanelConfig(baseURL, token, allowInsecure)
	var xray struct {
		XraySetting struct {
			Outbounds []interface{} `json:"outbounds"`
		} `json:"xraySetting"`
	}
	if err := json.Unmarshal([]byte(obj), &xray); err != nil {
		fmt.Println("Obj:", obj)
		panic(err)
	}
	return xray.XraySetting.Outbounds
}

func putPanelConfig(baseURL, token string, xraySetting map[string]interface{}, allowInsecure bool) {
	baseURL = strings.TrimRight(baseURL, "/")
	panelURL := baseURL + "/panel/xray/update"

	xrayJSON, err := json.Marshal(xraySetting)
	if err != nil {
		panic(err)
	}

	form := url.Values{}
	form.Set("xraySetting", string(xrayJSON))

	req, err := http.NewRequest(http.MethodPost, panelURL, strings.NewReader(form.Encode()))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Cookie", "3x-ui="+token)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")

	resp, err := xrayClient(allowInsecure).Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	bodyStr := strings.TrimSpace(string(body))
	if len(bodyStr) == 0 {
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return
		}
		panic(fmt.Sprintf("panel update status %d with empty body", resp.StatusCode))
	}
	var wrapper struct {
		Success bool   `json:"success"`
		Msg     string `json:"msg"`
	}
	if err := json.Unmarshal(body, &wrapper); err != nil {
		fmt.Println("Body:", string(body))
		panic(err)
	}
	if !wrapper.Success {
		fmt.Println("Body:", string(body))
		panic("panel update success=false: " + wrapper.Msg)
	}
}

func restartXrayService(baseURL, token string, allowInsecure bool) {
	baseURL = strings.TrimRight(baseURL, "/")
	restartURL := baseURL + "/panel/api/server/restartXrayService"

	req, err := http.NewRequest(http.MethodPost, restartURL, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Cookie", "3x-ui="+token)

	resp, err := xrayClient(allowInsecure).Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	bodyStr := strings.TrimSpace(string(body))
	if len(bodyStr) == 0 && resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return
	}
	if len(bodyStr) > 0 {
		var wrapper struct {
			Success bool   `json:"success"`
			Msg     string `json:"msg"`
		}
		if err := json.Unmarshal(body, &wrapper); err == nil && !wrapper.Success {
			panic("restartXrayService success=false: " + wrapper.Msg)
		}
	}
}

func runUpdate(cmd *cobra.Command, args []string) error {
	baseURL := os.Getenv("XUI_URL")
	username := os.Getenv("XUI_USERNAME")
	password := os.Getenv("XUI_PASSWORD")
	allowInsecure := os.Getenv("XUI_ALLOW_INSECURE") == "1" || strings.EqualFold(os.Getenv("XUI_ALLOW_INSECURE"), "true")
	outboundPrefix := os.Getenv("OUTBOUND_PREFIX")

	if baseURL == "" || username == "" || password == "" {
		return fmt.Errorf("XUI_URL, XUI_USERNAME, XUI_PASSWORD must be set")
	}

	token := getToken(baseURL, username, password, allowInsecure)
	obj := getPanelConfig(baseURL, token, allowInsecure)

	var config map[string]interface{}
	if err := json.Unmarshal([]byte(obj), &config); err != nil {
		fmt.Println("Obj:", obj)
		return err
	}

	xraySetting, _ := config["xraySetting"].(map[string]interface{})
	if xraySetting == nil {
		panic("xraySetting not found in panel config")
	}

	existing, _ := xraySetting["outbounds"].([]interface{})
	var outbounds []interface{}
	for _, o := range existing {
		ob, _ := o.(map[string]interface{})
		if ob == nil {
			outbounds = append(outbounds, o)
			continue
		}
		tag, _ := ob["tag"].(string)
		if outboundPrefix != "" && strings.HasPrefix(tag, outboundPrefix) {
			continue
		}
		outbounds = append(outbounds, o)
	}

	generated, err := os.ReadFile(generatedOutboundsPath)
	if err != nil {
		return err
	}
	var newOutbounds []interface{}
	if err := json.Unmarshal(generated, &newOutbounds); err != nil {
		return err
	}
	outbounds = append(outbounds, newOutbounds...)

	xraySetting["outbounds"] = outbounds
	putPanelConfig(baseURL, token, xraySetting, allowInsecure)
	restartXrayService(baseURL, token, allowInsecure)
	fmt.Println("[update] completed")
	return nil
}

package electricity

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

type mailgunConfiguration struct {
	toEmails, fromEmail []string
	mailgunApiKey       string
}

const (
	timestampFilePath = "/tmp/atys_last_call"
	tLayout           = time.RFC3339
)

func StartChecker() error {
	mailgunConfig, err := setEmailConfig()
	if err != nil {
		return fmt.Errorf("could not load email config: %v", err)
	}

	for {
		select {
		case <-time.After(5 * time.Second):
			if lastCallWasMoreThan30MinAgo() {
				sendAlertEmail(mailgunConfig)
			}
		}
	}
}

func sendAlertEmail(m mailgunConfiguration) {
	formData := make(url.Values)
	formData["from"] = m.fromEmail
	formData["to"] = m.toEmails
	formData["subject"] = []string{"NO tinc electricitat!"}
	formData["text"] = []string{fmt.Sprintf("Última trucada registrada: %v", getLastCallTime())}

	url := fmt.Sprintf("https://api:%v@api.mailgun.net/v3/xsimov.com/messages", m.mailgunApiKey)
	resp, _ := http.PostForm(url, formData)
	fmt.Println(resp.Status)
}

func lastCallWasMoreThan30MinAgo() bool {
	lastCall := getLastCallTime()
	return time.Since(lastCall) > 30*time.Minute
}

func setEmailConfig() (mailgunConfiguration, error) {
	var c mailgunConfiguration
	jsonConfig, err := ioutil.ReadFile("./mailgun_configuration.json")
	if err != nil {
		return mailgunConfiguration{}, fmt.Errorf("could not load mailgun configuration file: %v", err)
	}
	if err := json.Unmarshal(jsonConfig, &c); err != nil {
		return mailgunConfiguration{}, fmt.Errorf("could not unmarshal json %v: %v", jsonConfig, err)
	}
	return c, nil
}

func getLastCallTime() time.Time {
	strTime, err := ioutil.ReadFile(timestampFilePath)
	if err != nil {
		return time.Time{}
	}
	lastCallTime, err := time.Parse(tLayout, string(strTime))
	if err != nil {
		return time.Time{}
	}
	return lastCallTime
}

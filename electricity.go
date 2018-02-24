package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"
)

type mailgunConfiguration struct {
	ToEmails, FromEmail []string
	MailgunApiKey       string
}

const (
	heartbeatFilePath   = "/tmp/pdd_pulsometer_heartbeat_received_at"
	emailSentAtFilePath = "/tmp/pdd_pulsometer_email_sent_at"
	tLayout             = time.RFC3339
)

func HeartbeatReceived() {
	saveTimestampToDisk(heartbeatFilePath)
}

func StartPulsometer() {
	mailgunConfig, _ := setEmailConfig()
	// TODO: catch and log the possible errors when setting email config
	// if err != nil {
	// 	return fmt.Errorf("could not load email config: %v", err)
	// }

	var lastHeartbeat time.Time

	for {
		select {
		case <-time.After(5 * time.Second):
			if lastHeartbeatWasMoreThan30MinAgo(&lastHeartbeat) && notNotifiedToday() {
				sendAlertEmail(mailgunConfig, lastHeartbeat)
				saveTimestampToDisk(emailSentAtFilePath)
			}
		}
	}
}

func sendAlertEmail(m mailgunConfiguration, t time.Time) {
	formData := make(url.Values)
	formData["from"] = m.FromEmail
	formData["to"] = m.ToEmails
	formData["subject"] = []string{"NO tinc electricitat!"}
	formData["text"] = []string{fmt.Sprintf("Última trucada registrada: %v", t)}

	url := fmt.Sprintf("https://api:%v@api.mailgun.net/v3/xsimov.com/messages", m.MailgunApiKey)
	resp, _ := http.PostForm(url, formData)
	fmt.Println(resp.Status)
}

func lastHeartbeatWasMoreThan30MinAgo(t *time.Time) bool {
	*t = getTimestampFromFile(heartbeatFilePath)
	return time.Since(*t) > 30*time.Minute
}

func notNotifiedToday() bool {
	emailSentAt := getTimestampFromFile(emailSentAtFilePath)
	dayStart, dayEnd := dayBoundaries()
	return !(emailSentAt.After(dayStart) && emailSentAt.Before(dayEnd))
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

func dayBoundaries() (start time.Time, end time.Time) {
	t := time.Now()
	start, _ = time.Parse(tLayout, fmt.Sprintf("%04d-%02d-%02dT00:00:00Z", t.Year(), t.Month(), t.Day()))
	end = start.Add(24 * time.Hour)
	return
}

func getTimestampFromFile(path string) time.Time {
	strTime, err := ioutil.ReadFile(path)
	if err != nil {
		return time.Time{}
	}
	t, err := time.Parse(tLayout, string(strTime))
	if err != nil {
		return time.Time{}
	}
	return t
}

func saveTimestampToDisk(filePath string) {
	f, _ := os.Create(filePath)
	defer f.Close()
	// TODO: catch and log the possible errors opening the file
	// if err != nil {
	// 	return err
	// }
	fmt.Fprintf(f, "%v", time.Now().Format(tLayout))
}

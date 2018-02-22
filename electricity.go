package atys

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
	lastCallFilePath     = "/tmp/atys_last_call"
	lastNotifiedFilePath = "/tmp/atys_last_notified_at"
	tLayout              = time.RFC3339
)

func StartChecker() error {
	mailgunConfig, err := setEmailConfig()
	if err != nil {
		return fmt.Errorf("could not load email config: %v", err)
	}

	var lastCallWas time.Time

	for {
		select {
		case <-time.After(5 * time.Second):
			if lastCallWasMoreThan30MinAgo(&lastCallWas) && notNotifiedToday() {
				sendAlertEmail(mailgunConfig, lastCallWas)
				markAsAlreadyNotified()
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

func lastCallWasMoreThan30MinAgo(t *time.Time) bool {
	*t = getTimestampFromFile(lastCallFilePath)
	return time.Since(*t) > 30*time.Minute
}

func notNotifiedToday() bool {
	lastNotifiedAt := getTimestampFromFile(lastNotifiedFilePath)
	dayStart, dayEnd := dayBoundaries()
	return lastNotifiedAt.After(dayStart) && lastNotifiedAt.Before(dayEnd)
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
	start, _ = time.Parse(tLayout, fmt.Sprintf("%d%d%dT00:00:00Z", t.Year(), t.Month(), t.Day()))
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

func markAsAlreadyNotified() error {
	f, err := os.Create(lastNotifiedFilePath)
	defer f.Close()
	if err != nil {
		return err
	}
	fmt.Fprintf(f, "%v", time.Now())
	return nil
}

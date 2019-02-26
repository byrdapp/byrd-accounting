package slack

import (
	"os"

	"github.com/nlopes/slack"
)

// MsgBuilder msg builder for slack msgs
type MsgBuilder struct {
	TitleLink, Period, Text, Color, Footer, Pretext string
}

// NotifyPDFCreation notifies people when theres a newly generated PDF
func NotifyPDFCreation(s *MsgBuilder) error {
	att := []slack.Attachment{}
	a := slack.Attachment{
		Pretext:   s.Pretext,
		Title:     "Period: " + s.Period,
		TitleLink: s.TitleLink,
		Color:     s.Color,
		Fallback:  s.Text,
		Footer:    s.Footer,
	}
	att = append(att, a)
	msg := &slack.WebhookMessage{
		Text:        s.Text,
		Attachments: att,
	}

	if err := slack.PostWebhook(os.Getenv("SLACK_WEBHOOK"), msg); err != nil {
		return err
	}
	return nil
}

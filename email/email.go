package email

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/smtp2go-oss/smtp2go-go"
)

func SendEmail(recipient string, message string) (sent bool) {
	to := fmt.Sprintf("<%s>", recipient)

	email := smtp2go.Email{
		From: "Ask GPT <ai@jcwlkr.io>",
		To: []string{
			to,
		},
		Subject:  "New ticket from Ask GPT",
		TextBody: "This is ASK GPT, logging a new ticket.",
		HtmlBody: fmt.Sprintf("<h1>A message from Ask GPT</h1>\n<p>%s</p>\n<p>Kind regards,</p>\n<p>Ask GPT</p>", message),
	}
	_, err := smtp2go.Send(&email)
	if err != nil {
		log.Error().Msgf("An Error Occurred:", err)
		return false
	}
	log.Info().Msg("Sent Successfully")
	return true
}

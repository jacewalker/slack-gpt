package email

import (
	"fmt"

	"github.com/smtp2go-oss/smtp2go-go"
)

func SendEmail(recipient string) (sent bool) {
	to := fmt.Sprintf("<%s>", recipient)

	email := smtp2go.Email{
		From: "Ask GPT <ai@jcwlkr.io>",
		To: []string{
			to,
		},
		Subject:  "New ticket from Ask GPT",
		TextBody: "This is ASK GPT, logging a new ticket.",
		HtmlBody: `
		<h1>Ask GPT</h1>
		<p>I'm logging a new ticket!</p>
		<p>Kind regards,</p>
		<p>Ask GPT</p>
		`,
	}
	_, err := smtp2go.Send(&email)
	if err != nil {
		fmt.Println("An Error Occurred:", err)
		return false
	}
	fmt.Println("Sent Successfully")
	return true
}

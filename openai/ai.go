package openai

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/jacewalker/slack-gpt/email"
	"github.com/jacewalker/slack-gpt/slack"
	gogpt "github.com/sashabaranov/go-gpt3"
)

func MakePrompt(prompt string, apiKey *string, object *slack.SlackEvent) (response string) {
	c := gogpt.NewClient(*apiKey)
	ctx := context.Background()

	req := gogpt.CompletionRequest{
		Model:       gogpt.GPT3TextDavinci003,
		Prompt:      prompt,
		MaxTokens:   412,
		Temperature: 0.7,
		Stop: []string{
			// "Human:",
			"AI:",
			"Colleague:",
		},
		Suffix: "Act as a senior IT technical support agent speaking with a junior colleague over Slack. You will be asked IT technical support questions and are expected to answer them thoroughly.\nIf the response includes technical code or a script, it will be prefixed with three backticks.\nAI: Hi, how can I help you?\nColleague:",
	}

	resp, err := c.CreateCompletion(ctx, req)
	if err != nil {
		log.Println("Unable to complete GPT 3 request. Error:", err)
		return "Oops! There's been an error in my thinking. There may be an outage - I've let @jace know!\nFeel free to try me again."
	} else {
		log.Println("Received GPT Completion:\n\n", resp.Choices[0].Text)
	}

	return resp.Choices[0].Text
}

func MakeChatPrompt(prompt string, apiKey *string, history []gogpt.ChatCompletionMessage) (response string) {
	c := gogpt.NewClient(*apiKey)
	ctx := context.Background()

	message := []gogpt.ChatCompletionMessage{
		{Role: "system", Content: "You are a helpful assistant named Otto built to help Otto IT with technical support questions and script writing."},
	}
	message = append(history, gogpt.ChatCompletionMessage{Role: "user", Content: prompt})

	fmt.Println("FULL MESSEAGE: ", message)

	req := gogpt.ChatCompletionRequest{
		Model:    gogpt.GPT3Dot5Turbo,
		Messages: message,
	}

	resp, err := c.CreateChatCompletion(ctx, req)
	if err != nil {
		log.Println("Error:", err)
	} else {
		log.Println("Received Chat Completion Response: ", resp.Choices[0].Message.Content)
	}

	return resp.Choices[0].Message.Content
}

func CheckPromptType(prompt string, apiKey string) (response string) {
	c := gogpt.NewClient(apiKey)
	ctx := context.Background()

	req := gogpt.CompletionRequest{
		Model: gogpt.GPT3TextAda001,
		Prompt: `Text: "log a ticket for this"
		Label: "ticket"
		---
		Text: "can you put a ticket in"
		Label: "ticket"
		---
		Text: "can you help me with this problem"
		Label: "no ticket"
		---
		Text: "open a ticket for this issue"
		Label: "ticket"
		---
		Text: "this is not an issue, just a question"
		Label: "no ticket"
		---
		Text: ` + prompt + `
		Label:
		`,
		MaxTokens:   1,
		Temperature: 0,
		Stop: []string{
			"---",
		},
	}

	resp, err := c.CreateCompletion(ctx, req)
	if err != nil {
		log.Println("Unable to complete GPT 3 request. Error:", err)

		message := fmt.Sprintf("There's been an error:\n%s", err)
		email.SendEmail("jacewalker@me.com", message)
		return "Oops! There's been an error in my thinking. There may be an outage - I've let @jace know!\nFeel free to try me again."
	} else {
		log.Println("Received GPT Completion:\n\n", resp.Choices[0].Text)
	}

	if strings.Contains(resp.Choices[0].Text, "no ticket") {
		log.Println("No ticket to be logged.")
		return "not today sunnyboy"
	} else {
		log.Println("Logging a ticket.")
		return "its a day for happiness"
	}
}

func CreateHistoricPrompt(history map[string]string, newPrompt string) (prompt string) {
	var res string

	for key, value := range history {
		res += fmt.Sprintf("Colleague:%s\nAI:%s\n", key, value)
	}

	res += fmt.Sprintf("Colleague:%s\nAI:", newPrompt)
	return res
}

func CreateHistoricChatPrompt(history map[string]string, newPrompt string) (chatHistory []gogpt.ChatCompletionMessage) {
	var res []gogpt.ChatCompletionMessage

	for key, value := range history {
		res = append(res, gogpt.ChatCompletionMessage{Role: "assistant", Content: key})
		res = append(res, gogpt.ChatCompletionMessage{Role: "user", Content: value})
	}

	res = append(res, gogpt.ChatCompletionMessage{Role: "system", Content: "You are a helpful assistant named Otto built to help Otto IT with technical support questions and script writing."})

	return res
}

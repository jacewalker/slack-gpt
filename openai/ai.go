package openai

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/jacewalker/slack-gpt/slack"
	gogpt "github.com/sashabaranov/go-gpt3"
)

func MakePrompt(prompt string, apiKey string, object slack.SlackEvent) (response string) {
	c := gogpt.NewClient(apiKey)
	ctx := context.Background()

	req := gogpt.CompletionRequest{
		Model:       gogpt.GPT3TextDavinci003,
		Prompt:      prompt,
		MaxTokens:   412,
		Temperature: 0.8,
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
		return "Oops! There's been an error in my thinking. I've let <@U9WB4CL11 know!\nFeel free to try me again."
	} else {
		log.Println("Received GPT Completion:\n\n", resp.Choices[0].Text)
	}

	return resp.Choices[0].Text
}

func CheckPromptType(prompt string, apiKey string) (response string) {
	c := gogpt.NewClient(apiKey)
	ctx := context.Background()

	req := gogpt.CompletionRequest{
		Model:       gogpt.GPT3TextAda001,
		Prompt:      "Types of phrases:\n0 = \"Request for information\"\n1 = \"Request to open a support ticket\"\n\nThe following are phrases and their corresponding type:\n\nPhrase: \"can you log a ticket for this?\"\nType: 1\n\nPhrase: \"provide a powershell example that lists files in a directory\"\nType: 0\n\nPhrase: \"open a ticket for this\"\nType: 1\n\nPhrase: \"@askgpt open a case for this\"\nType: 1\n\nPhrase: \"What does RAM do?\"\nType: 0\n\nPhrase: \"@askgpt log a ticket\"\nType: 1\nPhrase: " + prompt + "\nType:",
		MaxTokens:   1,
		Temperature: 0.5,
		Stop: []string{
			"Phrase:",
			"Type:",
		},
		BestOf: 3,
	}

	resp, err := c.CreateCompletion(ctx, req)
	if err != nil {
		log.Println("Unable to complete GPT 3 request. Error:", err)
		return "Oops! There's been an error in my thinking. I've let <@U9WB4CL11 know!\nFeel free to try me again."
	} else {
		log.Println("Received GPT Completion:\n\n", resp.Choices[0].Text)
	}

	if strings.Contains(resp.Choices[0].Text, "0") {
		return "0"
	} else if strings.Contains(resp.Choices[0].Text, "1") {
		return "1"
	} else {
		return resp.Choices[0].Text
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

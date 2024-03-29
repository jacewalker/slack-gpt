package openai

import (
	"context"
	"strings"

	"github.com/rs/zerolog/log"
	gogpt "github.com/sashabaranov/go-openai"
)

func MakeChatPrompt(prompt string, apiKey *string, history []gogpt.ChatCompletionMessage) (response string) {
	c := gogpt.NewClient(*apiKey)
	ctx := context.Background()

	message := []gogpt.ChatCompletionMessage{}

	if history != nil {
		message = append(history, gogpt.ChatCompletionMessage{Role: "user", Content: prompt})
	}

	var req gogpt.ChatCompletionRequest

	if strings.Contains(prompt, "pretty please") {
		log.Info().Msg("Using GPT-4...")

		req = gogpt.ChatCompletionRequest{
			Model:    gogpt.GPT4,
			Messages: message,
		}
	} else {
		log.Info().Msg("Using GPT-3.5...")
		req = gogpt.ChatCompletionRequest{
			Model:    gogpt.GPT3Dot5Turbo,
			Messages: message,
		}
	}

	log.Info().Msg("Contacting OpenAI...")
	resp, err := c.CreateChatCompletion(ctx, req)
	if err != nil {
		log.Error().Msgf("Error:", err)
		return "I couldn't get a response from OpenAI, try sending that message again."
	} else {
		log.Info().Msgf("Received Chat Completion Response: ", resp.Choices[0].Message.Content)
		return resp.Choices[0].Message.Content
	}
}

func CheckPromptType(prompt string, apiKey *string) (response string) {
	c := gogpt.NewClient(*apiKey)
	ctx := context.Background()

	message := []gogpt.ChatCompletionMessage{
		{Role: "system", Content: "You are an AI system that interprets if a provided question is a request for a support ticket to be created or not.\nResponses should be a singular numeric character and not in sentence format.\nResponses should be a singular numeric digit.\nIf the request is of any other type, respond \"0\".\nIf the request is for a ticket to be created, respond \"1\"."},
		{Role: "user", Content: "Phrase: " + prompt},
	}

	req := gogpt.ChatCompletionRequest{
		Model:    gogpt.GPT3Dot5Turbo,
		Messages: message,
	}

	log.Info().Msg("Contacting OpenAI...")
	resp, err := c.CreateChatCompletion(ctx, req)
	if err != nil {
		log.Error().Msgf("Error:", err)
	} else {
		log.Info().Msgf("Received Chat Completion Response: ", resp.Choices[0].Message.Content)
	}

	return resp.Choices[0].Message.Content
}

func CreateHistoricChatPrompt(history map[string]string, newPrompt string) (chatHistory []gogpt.ChatCompletionMessage) {
	var res []gogpt.ChatCompletionMessage

	for key, value := range history {
		res = append(res, gogpt.ChatCompletionMessage{Role: "assistant", Content: key})
		res = append(res, gogpt.ChatCompletionMessage{Role: "user", Content: value})
	}

	res = append(res, gogpt.ChatCompletionMessage{Role: "system", Content: "Your name is Otto, you are designed to help with technical support questions and script writing. All requested steps should be broken down clearly."})

	return res
}

package ai

import (
	"fmt"
)

type Service struct {
	client       *Client
	chatID       string
	initialized  bool
	systemPrompt string
}

func NewService() (*Service, error) {
	c, err := NewClient()
	if err != nil {
		return nil, err
	}
	return &Service{
		client:       c,
		systemPrompt:  "Ты ZeroWorkflow AI - помощник разработчика. \nОтвечай кратко и по делу на русском языке.\nИспользуй markdown для форматирования.\nДля блоков кода используй тройные бэктики с указанием языка: " + "```язык\nкод\n```",
	}, nil
}

func (s *Service) ensureChat(firstUserMessage string) error {
	if s.chatID != "" {
		return nil
	}
	id, err := s.client.createNewChat(firstUserMessage)
	if err != nil {
		return fmt.Errorf("failed to init chat: %w", err)
	}
	s.chatID = id
	return nil
}

func (s *Service) Ask(question string) (string, error) {
	if err := s.ensureChat(question); err != nil {
		return "", err
	}
	msgs := []Message{{Role: "user", Content: question}}
	if !s.initialized {
		msgs = []Message{{Role: "system", Content: s.systemPrompt}, {Role: "user", Content: question}}
		s.initialized = true
	}
	return s.client.SendMessage(s.chatID, msgs)
}

func (s *Service) AskStream(question string, onDelta func(string)) (string, error) {
	if err := s.ensureChat(question); err != nil {
		return "", err
	}
	msgs := []Message{{Role: "user", Content: question}}
	if !s.initialized {
		msgs = []Message{{Role: "system", Content: s.systemPrompt}, {Role: "user", Content: question}}
		s.initialized = true
	}
	return s.client.SendMessageStream(s.chatID, msgs, onDelta)
}

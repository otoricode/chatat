package handler_test

import "context"

// mockSMSProvider implements service.SMSProvider for testing.
type mockSMSProvider struct {
	lastPhone   string
	lastMessage string
	sendError   error
}

func (m *mockSMSProvider) Send(phone string, message string) error {
	m.lastPhone = phone
	m.lastMessage = message
	return m.sendError
}

// mockWAProvider implements service.WhatsAppProvider for testing.
type mockWAProvider struct {
	businessNumber string
}

func (m *mockWAProvider) GetBusinessNumber() string {
	return m.businessNumber
}

func (m *mockWAProvider) SendMessage(_ context.Context, _ string, _ string) error {
	return nil
}

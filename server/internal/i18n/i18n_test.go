package i18n

import (
	"context"
	"fmt"
	"testing"
)

func TestGetLang_Default(t *testing.T) {
	ctx := context.Background()
	lang := GetLang(ctx)
	if lang != LangID {
		t.Errorf("expected %q, got %q", LangID, lang)
	}
}

func TestGetLang_FromContext(t *testing.T) {
	tests := []struct {
		set  Language
		want Language
	}{
		{LangID, LangID},
		{LangEN, LangEN},
		{LangAR, LangAR},
	}
	for _, tt := range tests {
		ctx := WithLang(context.Background(), tt.set)
		got := GetLang(ctx)
		if got != tt.want {
			t.Errorf("WithLang(%q): got %q, want %q", tt.set, got, tt.want)
		}
	}
}

func TestParseLang(t *testing.T) {
	tests := []struct {
		input string
		want  Language
	}{
		{"id", LangID},
		{"en", LangEN},
		{"ar", LangAR},
		{"ID", LangID},
		{"EN", LangEN},
		{"AR", LangAR},
		{"id-ID", LangID},
		{"en-US", LangEN},
		{"ar-SA", LangAR},
		{"id,en;q=0.9", LangID},
		{"en,id;q=0.8", LangEN},
		{"ar,en;q=0.7,id;q=0.5", LangAR},
		{"fr", LangID},
		{"", LangID},
		{"  en  ", LangEN},
	}
	for _, tt := range tests {
		got := ParseLang(tt.input)
		if got != tt.want {
			t.Errorf("ParseLang(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestT_ReturnsCorrectMessages(t *testing.T) {
	tests := []struct {
		lang Language
		want string
	}{
		{LangID, "Kode OTP telah dikirim"},
		{LangEN, "OTP code has been sent"},
	}
	for _, tt := range tests {
		ctx := WithLang(context.Background(), tt.lang)
		msgs := T(ctx)
		if msgs.Auth.OTPSent != tt.want {
			t.Errorf("T(%q).Auth.OTPSent = %q, want %q", tt.lang, msgs.Auth.OTPSent, tt.want)
		}
	}
}

func TestT_DefaultForUnknown(t *testing.T) {
	ctx := WithLang(context.Background(), Language("xx"))
	msgs := T(ctx)
	if msgs.Auth.OTPSent != "Kode OTP telah dikirim" {
		t.Error("expected Indonesian default for unknown language")
	}
}

func TestTLang(t *testing.T) {
	msgs := TLang(LangEN)
	if msgs.Error.NotFound != "Data not found" {
		t.Errorf("TLang(EN).Error.NotFound = %q, want %q", msgs.Error.NotFound, "Data not found")
	}

	msgs = TLang(Language("xx"))
	if msgs.Error.NotFound != "Data tidak ditemukan" {
		t.Error("expected Indonesian default for unknown language")
	}
}

func TestTf_FormatString(t *testing.T) {
	msgs := TLang(LangID)
	result := Tf(msgs.Chat.MemberAdded, "Ahmad")
	expected := "Ahmad telah ditambahkan"
	if result != expected {
		t.Errorf("Tf(MemberAdded, Ahmad) = %q, want %q", result, expected)
	}

	msgs = TLang(LangEN)
	result = Tf(msgs.Chat.MemberAdded, "Ahmad")
	expected = "Ahmad has been added"
	if result != expected {
		t.Errorf("Tf(MemberAdded, Ahmad) = %q, want %q", result, expected)
	}
}

func TestAllLanguagesHaveAllMessages(t *testing.T) {
	for _, lang := range []Language{LangID, LangEN, LangAR} {
		msgs := TLang(lang)
		if msgs == nil {
			t.Fatalf("TLang(%q) returned nil", lang)
		}
		if msgs.Auth.OTPSent == "" {
			t.Errorf("%s: Auth.OTPSent is empty", lang)
		}
		if msgs.Chat.MessageDeleted == "" {
			t.Errorf("%s: Chat.MessageDeleted is empty", lang)
		}
		if msgs.Document.Created == "" {
			t.Errorf("%s: Document.Created is empty", lang)
		}
		if msgs.Error.NotFound == "" {
			t.Errorf("%s: Error.NotFound is empty", lang)
		}
		if msgs.Search.MinQueryLength == "" {
			t.Errorf("%s: Search.MinQueryLength is empty", lang)
		}
		if fmt.Sprintf(msgs.Chat.MemberAdded, "X") == msgs.Chat.MemberAdded {
			t.Errorf("%s: Chat.MemberAdded missing format placeholder", lang)
		}
	}
}

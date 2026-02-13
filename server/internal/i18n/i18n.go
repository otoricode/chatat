package i18n

import (
	"context"
	"fmt"
	"strings"
)

// Language represents a supported language code.
type Language string

const (
	LangID Language = "id" // Indonesian (default)
	LangEN Language = "en" // English
	LangAR Language = "ar" // Arabic
)

type contextKey string

const langKey contextKey = "lang"

// WithLang adds the language to the context.
func WithLang(ctx context.Context, lang Language) context.Context {
	return context.WithValue(ctx, langKey, lang)
}

// GetLang extracts the language from context. Returns LangID as default.
func GetLang(ctx context.Context) Language {
	if lang, ok := ctx.Value(langKey).(Language); ok {
		return lang
	}
	return LangID
}

// ParseLang parses a language string. Returns LangID if not recognized.
func ParseLang(s string) Language {
	s = strings.ToLower(strings.TrimSpace(s))
	// Handle Accept-Language format: "id,en;q=0.9"
	if idx := strings.IndexByte(s, ','); idx >= 0 {
		s = s[:idx]
	}
	if idx := strings.IndexByte(s, ';'); idx >= 0 {
		s = s[:idx]
	}
	s = strings.TrimSpace(s)
	switch {
	case strings.HasPrefix(s, "id"):
		return LangID
	case strings.HasPrefix(s, "en"):
		return LangEN
	case strings.HasPrefix(s, "ar"):
		return LangAR
	default:
		return LangID
	}
}

// AuthMessages holds authentication-related messages.
type AuthMessages struct {
	OTPSent        string
	OTPInvalid     string
	OTPExpired     string
	PhoneRequired  string
	SessionExpired string
}

// ChatMessages holds chat-related messages.
type ChatMessages struct {
	MessageDeleted string
	GroupCreated   string
	MemberAdded    string // %s = member name
	MemberRemoved  string // %s = member name
	MemberLeft     string // %s = member name
}

// DocumentMessages holds document-related messages.
type DocumentMessages struct {
	Created            string
	Locked             string
	Unlocked           string
	SignatureRequested string
	Signed             string
	CannotEditLocked   string
}

// ErrorMessages holds error response messages.
type ErrorMessages struct {
	NotFound      string
	Unauthorized  string
	Forbidden     string
	BadRequest    string
	InternalError string
	RateLimited   string
}

// SearchMessages holds search-related messages.
type SearchMessages struct {
	MinQueryLength string
}

// Messages holds all localized message groups.
type Messages struct {
	Auth     AuthMessages
	Chat     ChatMessages
	Document DocumentMessages
	Error    ErrorMessages
	Search   SearchMessages
}

var translations = map[Language]*Messages{
	LangID: {
		Auth: AuthMessages{
			OTPSent:        "Kode OTP telah dikirim",
			OTPInvalid:     "Kode OTP tidak valid",
			OTPExpired:     "Kode OTP telah kedaluwarsa",
			PhoneRequired:  "Nomor telepon wajib diisi",
			SessionExpired: "Sesi telah berakhir, silakan login kembali",
		},
		Chat: ChatMessages{
			MessageDeleted: "Pesan telah dihapus",
			GroupCreated:   "Grup telah dibuat",
			MemberAdded:    "%s telah ditambahkan",
			MemberRemoved:  "%s telah dikeluarkan",
			MemberLeft:     "%s keluar dari grup",
		},
		Document: DocumentMessages{
			Created:            "Dokumen dibuat",
			Locked:             "Dokumen dikunci",
			Unlocked:           "Dokumen dibuka",
			SignatureRequested: "Permintaan tanda tangan dikirim",
			Signed:             "Dokumen ditandatangani",
			CannotEditLocked:   "Dokumen terkunci, tidak dapat diedit",
		},
		Error: ErrorMessages{
			NotFound:      "Data tidak ditemukan",
			Unauthorized:  "Tidak memiliki akses",
			Forbidden:     "Akses ditolak",
			BadRequest:    "Permintaan tidak valid",
			InternalError: "Terjadi kesalahan server",
			RateLimited:   "Terlalu banyak permintaan, coba lagi nanti",
		},
		Search: SearchMessages{
			MinQueryLength: "Kata kunci pencarian minimal 2 karakter",
		},
	},
	LangEN: {
		Auth: AuthMessages{
			OTPSent:        "OTP code has been sent",
			OTPInvalid:     "Invalid OTP code",
			OTPExpired:     "OTP code has expired",
			PhoneRequired:  "Phone number is required",
			SessionExpired: "Session has expired, please login again",
		},
		Chat: ChatMessages{
			MessageDeleted: "Message has been deleted",
			GroupCreated:   "Group has been created",
			MemberAdded:    "%s has been added",
			MemberRemoved:  "%s has been removed",
			MemberLeft:     "%s left the group",
		},
		Document: DocumentMessages{
			Created:            "Document created",
			Locked:             "Document locked",
			Unlocked:           "Document unlocked",
			SignatureRequested: "Signature request sent",
			Signed:             "Document signed",
			CannotEditLocked:   "Document is locked, cannot edit",
		},
		Error: ErrorMessages{
			NotFound:      "Data not found",
			Unauthorized:  "Unauthorized access",
			Forbidden:     "Access denied",
			BadRequest:    "Invalid request",
			InternalError: "Internal server error",
			RateLimited:   "Too many requests, try again later",
		},
		Search: SearchMessages{
			MinQueryLength: "Search query must be at least 2 characters",
		},
	},
	LangAR: {
		Auth: AuthMessages{
			OTPSent:        "\u062a\u0645 \u0625\u0631\u0633\u0627\u0644 \u0631\u0645\u0632 \u0627\u0644\u062a\u062d\u0642\u0642",
			OTPInvalid:     "\u0631\u0645\u0632 \u0627\u0644\u062a\u062d\u0642\u0642 \u063a\u064a\u0631 \u0635\u0627\u0644\u062d",
			OTPExpired:     "\u0627\u0646\u062a\u0647\u062a \u0635\u0644\u0627\u062d\u064a\u0629 \u0631\u0645\u0632 \u0627\u0644\u062a\u062d\u0642\u0642",
			PhoneRequired:  "\u0631\u0642\u0645 \u0627\u0644\u0647\u0627\u062a\u0641 \u0645\u0637\u0644\u0648\u0628",
			SessionExpired: "\u0627\u0646\u062a\u0647\u062a \u0627\u0644\u062c\u0644\u0633\u0629\u060c \u064a\u0631\u062c\u0649 \u062a\u0633\u062c\u064a\u0644 \u0627\u0644\u062f\u062e\u0648\u0644 \u0645\u0631\u0629 \u0623\u062e\u0631\u0649",
		},
		Chat: ChatMessages{
			MessageDeleted: "\u062a\u0645 \u062d\u0630\u0641 \u0627\u0644\u0631\u0633\u0627\u0644\u0629",
			GroupCreated:   "\u062a\u0645 \u0625\u0646\u0634\u0627\u0621 \u0627\u0644\u0645\u062c\u0645\u0648\u0639\u0629",
			MemberAdded:    "\u062a\u0645\u062a \u0625\u0636\u0627\u0641\u0629 %s",
			MemberRemoved:  "\u062a\u0645\u062a \u0625\u0632\u0627\u0644\u0629 %s",
			MemberLeft:     "\u063a\u0627\u062f\u0631 %s \u0627\u0644\u0645\u062c\u0645\u0648\u0639\u0629",
		},
		Document: DocumentMessages{
			Created:            "\u062a\u0645 \u0625\u0646\u0634\u0627\u0621 \u0627\u0644\u0645\u0633\u062a\u0646\u062f",
			Locked:             "\u062a\u0645 \u0642\u0641\u0644 \u0627\u0644\u0645\u0633\u062a\u0646\u062f",
			Unlocked:           "\u062a\u0645 \u0641\u062a\u062d \u0627\u0644\u0645\u0633\u062a\u0646\u062f",
			SignatureRequested: "\u062a\u0645 \u0625\u0631\u0633\u0627\u0644 \u0637\u0644\u0628 \u0627\u0644\u062a\u0648\u0642\u064a\u0639",
			Signed:             "\u062a\u0645 \u062a\u0648\u0642\u064a\u0639 \u0627\u0644\u0645\u0633\u062a\u0646\u062f",
			CannotEditLocked:   "\u0627\u0644\u0645\u0633\u062a\u0646\u062f \u0645\u0642\u0641\u0644\u060c \u0644\u0627 \u064a\u0645\u0643\u0646 \u0627\u0644\u062a\u0639\u062f\u064a\u0644",
		},
		Error: ErrorMessages{
			NotFound:      "\u0627\u0644\u0628\u064a\u0627\u0646\u0627\u062a \u063a\u064a\u0631 \u0645\u0648\u062c\u0648\u062f\u0629",
			Unauthorized:  "\u063a\u064a\u0631 \u0645\u0635\u0631\u062d \u0628\u0627\u0644\u0648\u0635\u0648\u0644",
			Forbidden:     "\u0627\u0644\u0648\u0635\u0648\u0644 \u0645\u0631\u0641\u0648\u0636",
			BadRequest:    "\u0637\u0644\u0628 \u063a\u064a\u0631 \u0635\u0627\u0644\u062d",
			InternalError: "\u062e\u0637\u0623 \u0641\u064a \u0627\u0644\u062e\u0627\u062f\u0645",
			RateLimited:   "\u0637\u0644\u0628\u0627\u062a \u0643\u062b\u064a\u0631\u0629 \u062c\u062f\u0627\u064b\u060c \u062d\u0627\u0648\u0644 \u0645\u0631\u0629 \u0623\u062e\u0631\u0649 \u0644\u0627\u062d\u0642\u0627\u064b",
		},
		Search: SearchMessages{
			MinQueryLength: "\u064a\u062c\u0628 \u0623\u0646 \u064a\u0643\u0648\u0646 \u0627\u0644\u0628\u062d\u062b \u062d\u0631\u0641\u064a\u0646 \u0639\u0644\u0649 \u0627\u0644\u0623\u0642\u0644",
		},
	},
}

// T returns the messages for the language in the given context.
func T(ctx context.Context) *Messages {
	lang := GetLang(ctx)
	if msgs, ok := translations[lang]; ok {
		return msgs
	}
	return translations[LangID]
}

// TLang returns the messages for the specified language.
func TLang(lang Language) *Messages {
	if msgs, ok := translations[lang]; ok {
		return msgs
	}
	return translations[LangID]
}

// Tf returns a formatted message using the specified format and arguments.
func Tf(format string, args ...interface{}) string {
	return fmt.Sprintf(format, args...)
}

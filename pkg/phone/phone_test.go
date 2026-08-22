package phone_test

import (
	"testing"

	"github.com/quixiq/polyglot/pkg/phone"
	"github.com/stretchr/testify/assert"
)

func TestNormalize(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "empty", input: "", want: ""},
		{name: "whitespace only", input: "   ", want: ""},
		{name: "no digits", input: "abc-@", want: ""},
		{name: "local 08 prefix", input: "081234567890", want: "6281234567890"},
		{name: "plus prefix", input: "+6281234567890", want: "6281234567890"},
		{name: "dashes and spaces", input: "0812-3456-7890", want: "6281234567890"},
		{name: "parentheses dots", input: "(021) 555.123", want: "6221555123"},
		{name: "already international", input: "6281234567890", want: "6281234567890"},
		{
			name:  "full JID with device suffix",
			input: "6281234567890:12@s.whatsapp.net",
			want:  "6281234567890",
		},
		{name: "bare JID without server", input: "6281234567890@s.whatsapp.net", want: "6281234567890"},
		{name: "group JID passthrough", input: "120363041234567890@g.us", want: "120363041234567890@g.us"},
		{name: "foreign number passthrough", input: "17289371293", want: "17289371293"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, phone.Normalize(tc.input))
		})
	}
}

func TestNormalize_Idempotent(t *testing.T) {
	samples := []string{"0812-3456-7890", "+62 812 3456 7890", "081234567890"}
	for _, s := range samples {
		once := phone.Normalize(s)
		assert.Equal(t, once, phone.Normalize(once))
	}
}

func TestToWhatsAppJID(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "local format", input: "081234567890", want: "6281234567890@s.whatsapp.net"},
		{name: "plus format", input: "+6281234567890", want: "6281234567890@s.whatsapp.net"},
		{name: "messy format", input: "0812-3456-7890", want: "6281234567890@s.whatsapp.net"},
		{name: "group passthrough", input: "120363041234567890@g.us", want: "120363041234567890@g.us"},
		{name: "invalid empty", input: "", want: ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, phone.ToWhatsAppJID(tc.input))
		})
	}
}

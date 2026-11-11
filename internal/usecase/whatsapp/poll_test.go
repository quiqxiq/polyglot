package whatsapp

import (
	"crypto/sha256"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsePollDirectiveNone(t *testing.T) {
	directive, prose, err := parsePollDirective("Status ONT normal, tidak ada redaman tinggi.")
	require.NoError(t, err)
	assert.Nil(t, directive)
	assert.Equal(t, "Status ONT normal, tidak ada redaman tinggi.", prose)
}

func TestParsePollDirectiveValid(t *testing.T) {
	input := "Ada beberapa perangkat aktif.\n```wa-poll\n" +
		`{"question":"Pilih perangkat yang ingin dicek","text":"Silakan tentukan perangkat:","select":1,"options":[{"label":"Router-Core","value":"router-core"},{"label":"OLT-POP1","value":"olt-pop1"}]}` +
		"\n```\nMohon pilih salah satu."

	directive, prose, err := parsePollDirective(input)
	require.NoError(t, err)
	require.NotNil(t, directive)
	assert.Equal(t, "Pilih perangkat yang ingin dicek", directive.Question)
	assert.Equal(t, "Silakan tentukan perangkat:", directive.Text)
	assert.Equal(t, 1, directive.Select)
	require.Len(t, directive.Options, 2)
	assert.Equal(t, "Router-Core", directive.Options[0].Label)
	assert.Equal(t, "router-core", directive.Options[0].Value)
	assert.Equal(t, "OLT-POP1", directive.Options[1].Label)
	assert.Equal(t, "olt-pop1", directive.Options[1].Value)

	assert.Contains(t, prose, "Ada beberapa perangkat aktif.")
	assert.Contains(t, prose, "Mohon pilih salah satu.")
	assert.NotContains(t, prose, "wa-poll")
	assert.NotContains(t, prose, "router-core")
}

func TestParsePollDirectiveStringOptions(t *testing.T) {
	input := "```wa-poll\n" +
		`{"question":"Pilih opsi","options":["Opsi 1","Opsi 2"]}` +
		"\n```"

	directive, _, err := parsePollDirective(input)
	require.NoError(t, err)
	require.NotNil(t, directive)
	assert.Equal(t, 1, directive.Select)
	require.Len(t, directive.Options, 2)
	assert.Equal(t, "Opsi 1", directive.Options[0].Label)
	assert.Equal(t, "Opsi 1", directive.Options[0].Value)
}

func TestParsePollDirectiveMultipleBlocksRejected(t *testing.T) {
	input := "```wa-poll\n{\"question\":\"A\",\"options\":[\"1\",\"2\"]}\n```\n" +
		"```wa-poll\n{\"question\":\"B\",\"options\":[\"1\",\"2\"]}\n```"

	directive, prose, err := parsePollDirective(input)
	require.Error(t, err)
	assert.Nil(t, directive)
	assert.Contains(t, err.Error(), "more than once")
	assert.NotContains(t, prose, "wa-poll")
}

func TestParsePollDirectiveInvalidJSON(t *testing.T) {
	input := "```wa-poll\n{invalid json}\n```"
	directive, prose, err := parsePollDirective(input)
	require.Error(t, err)
	assert.Nil(t, directive)
	assert.Equal(t, "", prose)
}

func TestParsePollDirectiveValidations(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		wantErr string
	}{
		{
			name:    "empty question",
			body:    `{"question":"","options":["1","2"]}`,
			wantErr: "poll question is empty",
		},
		{
			name:    "question exceeds limit",
			body:    `{"question":"` + strings.Repeat("q", maxPollQuestionRunes+1) + `","options":["1","2"]}`,
			wantErr: "poll question exceeds",
		},
		{
			name:    "less than 2 options",
			body:    `{"question":"Q","options":["1"]}`,
			wantErr: "requires at least two options",
		},
		{
			name:    "more than 12 options",
			body:    `{"question":"Q","options":["1","2","3","4","5","6","7","8","9","10","11","12","13"]}`,
			wantErr: "allows at most 12 options",
		},
		{
			name:    "duplicate option labels",
			body:    `{"question":"Q","options":["Router","router"]}`,
			wantErr: "duplicates another label",
		},
		{
			name:    "option label exceeds limit",
			body:    `{"question":"Q","options":["` + strings.Repeat("o", maxPollOptionRunes+1) + `","2"]}`,
			wantErr: "label exceeds",
		},
		{
			name:    "select out of bounds",
			body:    `{"question":"Q","select":5,"options":["1","2"]}`,
			wantErr: "poll select must be between 1 and 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := parsePollDirective("```wa-poll\n" + tt.body + "\n```")
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestMatchPollOptions(t *testing.T) {
	options := []pollOption{
		{Label: "Router Core", Value: "device-1"},
		{Label: "OLT POP 1", Value: "device-2"},
		{Label: "Switch Dist", Value: "device-3"},
	}
	pending := newPendingPoll(1, "chat@s.whatsapp.net", "poll-123", "cek status", options)

	hash1 := sha256.Sum256([]byte("Router Core"))
	hash3 := sha256.Sum256([]byte("Switch Dist"))
	unknownHash := sha256.Sum256([]byte("Tidak Ada"))

	selected := matchPollOptions(pending.options, [][]byte{hash3[:], hash1[:], unknownHash[:]})
	require.Len(t, selected, 2)
	assert.Equal(t, "Switch Dist", selected[0].Label)
	assert.Equal(t, "device-3", selected[0].Value)
	assert.Equal(t, "Router Core", selected[1].Label)
	assert.Equal(t, "device-1", selected[1].Value)
}

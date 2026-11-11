package whatsapp

import (
	"context"
	"errors"
	"strings"
	"testing"

	domainWA "github.com/quixiq/polyglot/internal/domain/whatsapp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeResolver struct {
	identity *domainWA.SenderIdentity
	err      error
	calls    int
	raw      string
}

func (f *fakeResolver) ResolveFresh(_ context.Context, rawSender string) (*domainWA.SenderIdentity, error) {
	f.calls++
	f.raw = rawSender
	return f.identity, f.err
}

type fakeHarness struct {
	reply  string
	err    error
	calls  int
	prompt string
}

func (f *fakeHarness) Ask(_ context.Context, prompt string) (string, error) {
	f.calls++
	f.prompt = prompt
	return f.reply, f.err
}

type sentMessage struct {
	sessionID uint
	to        string
	content   string
}

type fakeSender struct {
	messages []sentMessage
	polls    []sentPoll
	presence []sentPresence
	err      error
	pollID   string
	pollErr  error
}

type sentPoll struct {
	sessionID  uint
	to         string
	question   string
	options    []string
	selectable int
}

type sentPresence struct {
	sessionID uint
	to        string
	state     string
	media     string
}

func (f *fakeSender) SendMessageContext(_ context.Context, sessionID uint, to string, content string) error {
	f.messages = append(f.messages, sentMessage{sessionID: sessionID, to: to, content: content})
	return f.err
}

func (f *fakeSender) SendPoll(_ context.Context, sessionID uint, to string, question string, options []string, selectable int) (string, error) {
	f.polls = append(f.polls, sentPoll{sessionID: sessionID, to: to, question: question, options: options, selectable: selectable})
	if f.pollErr != nil {
		return "", f.pollErr
	}
	id := f.pollID
	if id == "" {
		id = "fake-poll-id"
	}
	return id, nil
}

func (f *fakeSender) SendChatPresence(_ context.Context, sessionID uint, to string, state string, media string) error {
	f.presence = append(f.presence, sentPresence{sessionID: sessionID, to: to, state: state, media: media})
	return nil
}

func adminIdentity() *domainWA.SenderIdentity {
	return &domainWA.SenderIdentity{
		Role:          domainWA.RoleAdmin,
		DisplayName:   "Budi",
		AllowedScopes: []string{domainWA.ScopeAll},
	}
}

func customerIdentity() *domainWA.SenderIdentity {
	return &domainWA.SenderIdentity{
		Role:          domainWA.RoleCustomer,
		DisplayName:   "Pelanggan",
		AllowedScopes: domainWA.DefaultScopesForRole(domainWA.RoleCustomer),
	}
}

func TestAssistantIgnoresOtherSessions(t *testing.T) {
	resolver := &fakeResolver{identity: adminIdentity()}
	harness := &fakeHarness{reply: "ok"}
	sender := &fakeSender{}
	assistant := NewAssistant(resolver, harness, sender, 7)

	require.NoError(t, assistant.HandleIncoming(context.Background(), 8, "msg-1", "628@s.whatsapp.net", "628", "halo"))
	assert.Zero(t, resolver.calls)
	assert.Zero(t, harness.calls)
	assert.Empty(t, sender.messages)
}

func TestAssistantDisabledWithoutSession(t *testing.T) {
	assistant := NewAssistant(&fakeResolver{}, &fakeHarness{}, &fakeSender{}, 0)
	assert.False(t, assistant.Enabled())
	require.NoError(t, assistant.HandleIncoming(context.Background(), 1, "msg-1", "628@s.whatsapp.net", "628", "halo"))
}

func TestAssistantRepliesAdminOnlyToNonAdmin(t *testing.T) {
	resolver := &fakeResolver{identity: customerIdentity()}
	harness := &fakeHarness{reply: "should not be used"}
	sender := &fakeSender{}
	assistant := NewAssistant(resolver, harness, sender, 7)

	require.NoError(t, assistant.HandleIncoming(context.Background(), 7, "msg-1", "628@s.whatsapp.net", "628", "halo"))
	assert.Equal(t, 1, resolver.calls)
	assert.Equal(t, "628", resolver.raw)
	assert.Zero(t, harness.calls)
	require.Len(t, sender.messages, 1)
	assert.Equal(t, NonAdminReply, sender.messages[0].content)
	assert.Equal(t, uint(7), sender.messages[0].sessionID)
	assert.Equal(t, "628@s.whatsapp.net", sender.messages[0].to)
}

func TestAssistantDeniesWhenResolutionFails(t *testing.T) {
	resolver := &fakeResolver{err: errors.New("database unavailable")}
	harness := &fakeHarness{reply: "should not be used"}
	sender := &fakeSender{}
	assistant := NewAssistant(resolver, harness, sender, 7)

	require.NoError(t, assistant.HandleIncoming(context.Background(), 7, "msg-1", "628@s.whatsapp.net", "628", "halo"))
	assert.Zero(t, harness.calls)
	require.Len(t, sender.messages, 1)
	assert.Equal(t, NonAdminReply, sender.messages[0].content)
}

func TestAssistantAllowsAdmin(t *testing.T) {
	resolver := &fakeResolver{identity: adminIdentity()}
	harness := &fakeHarness{reply: "  status ONT normal  "}
	sender := &fakeSender{}
	assistant := NewAssistant(resolver, harness, sender, 7)

	require.NoError(t, assistant.HandleIncoming(context.Background(), 7, "msg-1", "628@s.whatsapp.net", "628", "cek pelanggan 08123"))
	assert.Equal(t, 1, harness.calls)
	assert.Contains(t, harness.prompt, "Budi")
	assert.Contains(t, harness.prompt, "cek pelanggan 08123")
	require.Len(t, sender.messages, 1)
	assert.Equal(t, "status ONT normal", sender.messages[0].content)
}

func TestAssistantAllowsWildcardScope(t *testing.T) {
	resolver := &fakeResolver{identity: &domainWA.SenderIdentity{
		Role:          domainWA.RoleAdmin,
		DisplayName:   "Owner",
		AllowedScopes: []string{domainWA.ScopeAll},
	}}
	harness := &fakeHarness{reply: "ok"}
	sender := &fakeSender{}
	assistant := NewAssistant(resolver, harness, sender, 7)

	require.NoError(t, assistant.HandleIncoming(context.Background(), 7, "msg-1", "628@s.whatsapp.net", "628", "halo"))
	assert.Equal(t, 1, harness.calls)
	require.Len(t, sender.messages, 1)
	assert.Equal(t, "ok", sender.messages[0].content)
}

func TestAssistantHandlesHarnessFailure(t *testing.T) {
	resolver := &fakeResolver{identity: adminIdentity()}
	harness := &fakeHarness{err: errors.New("harness down")}
	sender := &fakeSender{}
	assistant := NewAssistant(resolver, harness, sender, 7)

	require.NoError(t, assistant.HandleIncoming(context.Background(), 7, "msg-1", "628@s.whatsapp.net", "628", "halo"))
	require.Len(t, sender.messages, 1)
	assert.Equal(t, AssistantUnavailableReply, sender.messages[0].content)
}

func TestAssistantHandlesEmptyReply(t *testing.T) {
	resolver := &fakeResolver{identity: adminIdentity()}
	harness := &fakeHarness{reply: "   "}
	sender := &fakeSender{}
	assistant := NewAssistant(resolver, harness, sender, 7)

	require.NoError(t, assistant.HandleIncoming(context.Background(), 7, "msg-1", "628@s.whatsapp.net", "628", "halo"))
	require.Len(t, sender.messages, 1)
	assert.Equal(t, AssistantUnavailableReply, sender.messages[0].content)
}

func TestAssistantIgnoresEmptyContent(t *testing.T) {
	resolver := &fakeResolver{identity: adminIdentity()}
	harness := &fakeHarness{reply: "ok"}
	sender := &fakeSender{}
	assistant := NewAssistant(resolver, harness, sender, 7)

	require.NoError(t, assistant.HandleIncoming(context.Background(), 7, "msg-1", "628@s.whatsapp.net", "628", "   "))
	assert.Zero(t, resolver.calls)
	assert.Zero(t, harness.calls)
	assert.Empty(t, sender.messages)
}

func TestAssistantTruncatesLongReply(t *testing.T) {
	resolver := &fakeResolver{identity: adminIdentity()}
	harness := &fakeHarness{reply: strings.Repeat("a", maxAssistantReplyRunes+100)}
	sender := &fakeSender{}
	assistant := NewAssistant(resolver, harness, sender, 7)

	require.NoError(t, assistant.HandleIncoming(context.Background(), 7, "msg-1", "628@s.whatsapp.net", "628", "halo"))
	require.Len(t, sender.messages, 1)
	runes := []rune(sender.messages[0].content)
	assert.Len(t, runes, maxAssistantReplyRunes+3)
	assert.True(t, strings.HasSuffix(sender.messages[0].content, "…"))
}

func TestAssistantReturnsSendFailure(t *testing.T) {
	resolver := &fakeResolver{identity: adminIdentity()}
	harness := &fakeHarness{reply: "ok"}
	sender := &fakeSender{err: errors.New("socket closed")}
	assistant := NewAssistant(resolver, harness, sender, 7)

	err := assistant.HandleIncoming(context.Background(), 7, "msg-1", "628@s.whatsapp.net", "628", "halo")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "socket closed")
}

func TestAssistantFormatsWhatsAppMarkdown(t *testing.T) {
	resolver := &fakeResolver{identity: adminIdentity()}
	harness := &fakeHarness{reply: "**Status Ringkas**\n- ONT `online`\n- Redaman *normal*"}
	sender := &fakeSender{}
	assistant := NewAssistant(resolver, harness, sender, 7)

	require.NoError(t, assistant.HandleIncoming(context.Background(), 7, "msg-1", "628@s.whatsapp.net", "628", "cek ont"))
	require.Len(t, sender.messages, 1)
	assert.Equal(t, "*Status Ringkas*\n• ONT ```online```\n• Redaman _normal_", sender.messages[0].content)
}

func TestAssistantSendsTypingPresence(t *testing.T) {
	resolver := &fakeResolver{identity: adminIdentity()}
	harness := &fakeHarness{reply: "siap"}
	sender := &fakeSender{}
	assistant := NewAssistant(resolver, harness, sender, 7)

	require.NoError(t, assistant.HandleIncoming(context.Background(), 7, "msg-1", "628@s.whatsapp.net", "628", "halo"))
	require.NotEmpty(t, sender.presence)
	assert.Equal(t, "composing", sender.presence[0].state)
	assert.Equal(t, "paused", sender.presence[len(sender.presence)-1].state)
}

func TestAssistantSendsPoll(t *testing.T) {
	resolver := &fakeResolver{identity: adminIdentity()}
	harness := &fakeHarness{reply: "Ada 2 perangkat.\n```wa-poll\n" +
		`{"question":"Pilih router","text":"Tentukan target:","select":1,"options":[{"label":"Core-1","value":"core-1"},{"label":"Core-2","value":"core-2"}]}` +
		"\n```"}
	sender := &fakeSender{pollID: "poll-999"}
	assistant := NewAssistant(resolver, harness, sender, 7)

	require.NoError(t, assistant.HandleIncoming(context.Background(), 7, "msg-1", "628@s.whatsapp.net", "628", "cek router"))

	require.Len(t, sender.messages, 1)
	assert.Equal(t, "Tentukan target:", sender.messages[0].content)

	require.Len(t, sender.polls, 1)
	assert.Equal(t, "Pilih router", sender.polls[0].question)
	assert.Equal(t, []string{"Core-1", "Core-2"}, sender.polls[0].options)
	assert.Equal(t, 1, sender.polls[0].selectable)

	pending := assistant.peekPendingPoll("poll-999")
	require.NotNil(t, pending)
	assert.Equal(t, "cek router", pending.request)
	assert.Len(t, pending.options, 2)
}

func TestAssistantPollFallbackOnSendError(t *testing.T) {
	resolver := &fakeResolver{identity: adminIdentity()}
	harness := &fakeHarness{reply: "```wa-poll\n" +
		`{"question":"Pilih router","select":1,"options":[{"label":"Core-1","value":"core-1"},{"label":"Core-2","value":"core-2"}]}` +
		"\n```"}
	sender := &fakeSender{pollErr: errors.New("poll not supported on old client")}
	assistant := NewAssistant(resolver, harness, sender, 7)

	require.NoError(t, assistant.HandleIncoming(context.Background(), 7, "msg-1", "628@s.whatsapp.net", "628", "cek router"))

	require.Len(t, sender.messages, 1)
	assert.Contains(t, sender.messages[0].content, "Pilih router")
	assert.Contains(t, sender.messages[0].content, "• Core-1")
	assert.Contains(t, sender.messages[0].content, "• Core-2")
}

func TestAssistantPollFallbackOnInvalidDirective(t *testing.T) {
	resolver := &fakeResolver{identity: adminIdentity()}
	harness := &fakeHarness{reply: "Silakan pilih salah satu.\n```wa-poll\n{invalid json}\n```"}
	sender := &fakeSender{}
	assistant := NewAssistant(resolver, harness, sender, 7)

	require.NoError(t, assistant.HandleIncoming(context.Background(), 7, "msg-1", "628@s.whatsapp.net", "628", "cek router"))

	require.Len(t, sender.messages, 1)
	assert.Equal(t, "Silakan pilih salah satu.", sender.messages[0].content)
	assert.Empty(t, sender.polls)
}

func TestAssistantHandlePollVoteAdmin(t *testing.T) {
	resolver := &fakeResolver{identity: adminIdentity()}
	sender := &fakeSender{}
	assistant := NewAssistant(resolver, nil, sender, 7)

	options := []pollOption{
		{Label: "Router-Core", Value: "device-1"},
		{Label: "OLT-POP1", Value: "device-2"},
	}
	pending := newPendingPoll(7, "628@s.whatsapp.net", "poll-abc", "cek perangkat", options)
	assistant.storePendingPoll(pending)

	harness := &fakeHarness{reply: "**Status Router-Core normal** (CPU 15%)"}
	assistant.harness = harness

	voteHash := pending.options[0].Hash[:]
	require.NoError(t, assistant.HandlePollVote(context.Background(), 7, "628@s.whatsapp.net", "poll-abc", "628", [][]byte{voteHash}))

	assert.Equal(t, 1, harness.calls)
	assert.Contains(t, harness.prompt, "cek perangkat")
	assert.Contains(t, harness.prompt, "Router-Core")
	assert.Contains(t, harness.prompt, "device-1")

	require.Len(t, sender.messages, 1)
	assert.Equal(t, "*Status Router-Core normal* (CPU 15%)", sender.messages[0].content)

	assert.Nil(t, assistant.peekPendingPoll("poll-abc"))
}

func TestAssistantHandlePollVoteNonAdmin(t *testing.T) {
	resolver := &fakeResolver{identity: customerIdentity()}
	sender := &fakeSender{}
	assistant := NewAssistant(resolver, &fakeHarness{}, sender, 7)

	options := []pollOption{
		{Label: "Option 1", Value: "val-1"},
		{Label: "Option 2", Value: "val-2"},
	}
	pending := newPendingPoll(7, "628@s.whatsapp.net", "poll-xyz", "req", options)
	assistant.storePendingPoll(pending)

	voteHash := pending.options[0].Hash[:]
	require.NoError(t, assistant.HandlePollVote(context.Background(), 7, "628@s.whatsapp.net", "poll-xyz", "628", [][]byte{voteHash}))

	require.Len(t, sender.messages, 1)
	assert.Equal(t, NonAdminReply, sender.messages[0].content)
}

func TestAssistantHandlePollVoteUnknownPoll(t *testing.T) {
	resolver := &fakeResolver{identity: adminIdentity()}
	sender := &fakeSender{}
	assistant := NewAssistant(resolver, &fakeHarness{}, sender, 7)

	require.NoError(t, assistant.HandlePollVote(context.Background(), 7, "628@s.whatsapp.net", "nonexistent-poll", "628", [][]byte{[]byte("hash")}))
	assert.Empty(t, sender.messages)
}

package billing_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainBilling "github.com/quixiq/polyglot/internal/domain/billing"
	"github.com/quixiq/polyglot/internal/port/mocktest"
	uc "github.com/quixiq/polyglot/internal/usecase/billing"
)

func TestReversePayment(t *testing.T) {
	reverser := &mocktest.FakePaymentReverser{Pay: domainBilling.Payment{ID: "pay-1"}}
	useCase := uc.NewPaymentUseCase(reverser)
	actor := uint(9)

	reversed, err := useCase.Reverse(context.Background(), "pay-1", "salah input nominal", &actor)
	require.NoError(t, err)
	assert.Equal(t, "pay-1", reversed.ID)
	require.Len(t, reverser.In, 1)
	assert.Equal(t, "pay-1", reverser.In[0].PaymentID)
	assert.Equal(t, "salah input nominal", reverser.In[0].Reason)
	require.NotNil(t, reverser.In[0].ActorID)
	assert.Equal(t, actor, *reverser.In[0].ActorID)
}

func TestReversePayment_RequiresReason(t *testing.T) {
	reverser := &mocktest.FakePaymentReverser{}
	useCase := uc.NewPaymentUseCase(reverser)
	_, err := useCase.Reverse(context.Background(), "pay-1", " ", nil)
	assert.ErrorIs(t, err, domainBilling.ErrInvalidInput)
	assert.Empty(t, reverser.In)
}

func TestReversePayment_ReverserError(t *testing.T) {
	reverser := &mocktest.FakePaymentReverser{Err: errors.New("boom")}
	useCase := uc.NewPaymentUseCase(reverser)
	_, err := useCase.Reverse(context.Background(), "pay-1", "koreksi", nil)
	require.Error(t, err)
}

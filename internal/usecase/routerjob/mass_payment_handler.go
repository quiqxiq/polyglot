package routerjob

import (
	"context"
	"encoding/json"
	"fmt"

	domainRouterJob "github.com/quixiq/polyglot/internal/domain/routerjob"
)

type massPaymentPayload struct {
	BatchID string `json:"batch_id"`
}

func massPaymentHandler(deps HandlerDeps) Handler {
	return func(ctx context.Context, job domainRouterJob.Job) error {
		if deps.MassPay == nil {
			return fmt.Errorf("mass payment action belum dikonfigurasi")
		}
		var p massPaymentPayload
		if err := json.Unmarshal(job.Payload, &p); err != nil || p.BatchID == "" {
			return fmt.Errorf("payload mass payment tidak valid: %w", err)
		}
		return deps.MassPay.ProcessBatch(ctx, p.BatchID)
	}
}

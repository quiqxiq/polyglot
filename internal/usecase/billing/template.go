package billing

import (
	"context"
	"strings"

	"github.com/quixiq/polyglot/internal/port"
)

// renderTemplateOrDefault mengganti placeholder {{key}} pada template dari
// repo; bila template tidak ditemukan, pakai fallback.
func renderTemplateOrDefault(ctx context.Context, repo port.NotificationRepository, tenantID, key, fallback string, vars map[string]string) string {
	tpl, err := repo.FindTemplateByKey(ctx, tenantID, key)
	if err != nil {
		return fallback
	}
	rep := make([]string, 0, len(vars)*2)
	for k, v := range vars {
		rep = append(rep, "{{"+k+"}}", v)
	}
	return strings.NewReplacer(rep...).Replace(tpl.Content)
}

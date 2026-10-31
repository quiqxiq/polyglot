package ont

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	domainRegistration "github.com/quixiq/polyglot/internal/domain/registration"
	"github.com/quixiq/polyglot/internal/domain/routerjob"
	"github.com/quixiq/polyglot/internal/domain/ont"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/idgen"
)

// BindFromPSB menautkan ONT hasil instalasi PSB ke pelanggan/langganan dan
// menjadwalkan rename ONU ke kode pelanggan (best-effort dari convert).
func (s *Service) BindFromPSB(ctx context.Context, in domainRegistration.PSBBindInput) error {
	o, err := s.findONTForPSB(ctx, in)
	if err != nil {
		return err
	}
	if in.Label != "" {
		if errRename := s.enqueueRenameForPSB(ctx, o, in.Label); errRename != nil {
			return errRename
		}
	}
	return s.Link(ctx, o.ID, LinkInput{
		CustomerID:     in.CustomerID,
		SubscriptionID: in.SubscriptionID,
		PPPoEUsername:  in.PPPoEUsername,
	})
}

func (s *Service) findONTForPSB(ctx context.Context, in domainRegistration.PSBBindInput) (ont.ONT, error) {
	if in.OLTDeviceID != "" && in.MACAddress != "" {
		if o, err := s.deps.ONTs.FindByOLTAndMAC(ctx, in.OLTDeviceID, in.MACAddress); err == nil {
			return o, nil
		}
	}
	wantMAC, _ := ont.NormalizeMAC(in.MACAddress)
	wantSerial := ont.NormalizeSerial(in.SerialNumber)
	search := in.MACAddress
	if strings.TrimSpace(search) == "" {
		search = in.SerialNumber
	}
	list, err := s.deps.ONTs.List(ctx, port.ONTFilter{OLTDeviceID: in.OLTDeviceID, Search: search, Limit: 500})
	if err != nil {
		return ont.ONT{}, err
	}
	for _, o := range list {
		if wantMAC != "" && o.MACAddress == wantMAC {
			return o, nil
		}
		if wantSerial != "" && ont.NormalizeSerial(o.SerialNumber) == wantSerial {
			return o, nil
		}
	}
	return ont.ONT{}, fmt.Errorf("%w: ONT %s/%s belum terdeteksi pada OLT %s",
		ont.ErrNotFound, in.MACAddress, in.SerialNumber, in.OLTDeviceID)
}

func (s *Service) enqueueRenameForPSB(ctx context.Context, o ont.ONT, label string) error {
	if s.deps.Jobs == nil || o.OLTDeviceID == "" {
		return nil
	}
	raw, err := json.Marshal(map[string]any{
		"olt_device_id": o.OLTDeviceID,
		"pon_index":     o.PONIndex,
		"onu_index":     o.ONUIndex,
		"name":          label,
	})
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	_, err = s.deps.Jobs.Enqueue(ctx, routerjob.Job{
		ID:             idgen.New("rj"),
		TenantID:       tenantDefault,
		DeviceID:       o.OLTDeviceID,
		JobType:        routerjob.TypeONTRename,
		Payload:        raw,
		IdempotencyKey: fmt.Sprintf("psb-ont-rename:%s", o.ID),
		Status:         routerjob.StatusQueued,
		MaxAttempts:    jobMaxAttempt,
		NextRunAt:      now,
	})
	return err
}

package olt

import (
	"context"
	"fmt"
	"strconv"
	"time"

	domainOLT "github.com/quixiq/polyglot/internal/domain/olt"
	"github.com/quixiq/polyglot/internal/domain/ont"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/idgen"
)

// SyncResult meringkas hasil satu siklus sync ONU.
type SyncResult struct {
	Discovered int       `json:"discovered"`
	Updated    int       `json:"updated"`
	Removed    int       `json:"removed"`
	Online     int       `json:"online"`
	Offline    int       `json:"offline"`
	PolledAt   time.Time `json:"polled_at"`
}

type syncConfig struct {
	missThreshold  int
	offlineStrikes int
	rxNAStrikes    int
	rxWarning      float64
	rxCritical     float64
}

func (s *Service) loadSyncConfig(ctx context.Context) syncConfig {
	cfg := syncConfig{missThreshold: 12, offlineStrikes: 2, rxNAStrikes: 3, rxWarning: -27, rxCritical: -30}
	if v, err := strconv.Atoi(s.deps.Settings.GetValue(ctx, "olt.roster_miss_threshold", "12")); err == nil && v > 0 {
		cfg.missThreshold = v
	}
	if v, err := strconv.Atoi(s.deps.Settings.GetValue(ctx, "olt.offline_strikes", "2")); err == nil && v > 0 {
		cfg.offlineStrikes = v
	}
	if v, err := strconv.Atoi(s.deps.Settings.GetValue(ctx, "olt.rx_na_strikes", "3")); err == nil && v > 0 {
		cfg.rxNAStrikes = v
	}
	if v, err := strconv.ParseFloat(s.deps.Settings.GetValue(ctx, "olt.rx_warning_dbm", "-27"), 64); err == nil {
		cfg.rxWarning = v
	}
	if v, err := strconv.ParseFloat(s.deps.Settings.GetValue(ctx, "olt.rx_critical_dbm", "-30"), 64); err == nil {
		cfg.rxCritical = v
	}
	return cfg
}

// SyncONUs membaca inventori ONU dari OLT dan menerapkannya ke DB dengan
// semantik roster: carry-forward, anti-flap, event timeline, dan sampel optik.
func (s *Service) SyncONUs(ctx context.Context, oltDeviceID string) (SyncResult, error) {
	result := SyncResult{PolledAt: time.Now().UTC()}
	dev, err := s.deps.Devices.FindByID(ctx, oltDeviceID)
	if err != nil {
		return result, err
	}
	if dev.DriverType != driverTypeHioso {
		return result, fmt.Errorf("%w: device bukan OLT hioso", domainOLT.ErrUnsupportedVariant)
	}
	mgr, err := s.deps.Managers.ForDevice(ctx, oltDeviceID)
	if err != nil {
		return result, err
	}
	rows, err := mgr.ListONUs(ctx, 0)
	if err != nil {
		return result, err
	}
	now := time.Now().UTC()
	cfg := s.loadSyncConfig(ctx)

	existing, err := s.deps.ONTs.ListByOLT(ctx, oltDeviceID)
	if err != nil {
		return result, fmt.Errorf("olt: baca inventori ONT: %w", err)
	}
	byMAC := make(map[string]ont.ONT, len(existing))
	byPos := make(map[[2]int]ont.ONT, len(existing))
	for _, o := range existing {
		byMAC[o.MACAddress] = o
		if o.State != ont.StateRemoved {
			byPos[[2]int{o.PONIndex, o.ONUIndex}] = o
		}
	}
	seen := make(map[string]bool, len(rows))

	for _, row := range rows {
		cur, ok := byMAC[row.MACAddress]
		if !ok {
			// ONU pengganti di posisi yang sama: baris lama jadi REMOVED.
			pos := [2]int{row.PONIndex, row.ONUIndex}
			if old, exists := byPos[pos]; exists {
				old.State = ont.StateRemoved
				if err := s.deps.ONTs.Save(ctx, old); err == nil {
					seen[old.ID] = true
					delete(byMAC, old.MACAddress)
					s.recordEvent(ctx, old, "REMOVED", string(old.State), string(ont.StateRemoved), "posisi PON/ONU dipakai ONU lain")
					result.Removed++
				}
			}
			created, errCreate := s.createONT(ctx, deviceRef{ID: dev.ID, TenantID: dev.TenantID}, row, now, cfg)
			if errCreate != nil {
				return result, errCreate
			}
			seen[created.ID] = true
			result.Discovered++
			if created.State == ont.StateOnline {
				result.Online++
			} else {
				result.Offline++
			}
			continue
		}
		seen[cur.ID] = true

		// ONU yang sama pindah posisi: pastikan posisi baru tidak dipakai.
		if cur.PONIndex != row.PONIndex || cur.ONUIndex != row.ONUIndex {
			newPos := [2]int{row.PONIndex, row.ONUIndex}
			if other, exists := byPos[newPos]; exists && other.ID != cur.ID {
				other.State = ont.StateRemoved
				if err := s.deps.ONTs.Save(ctx, other); err == nil {
					seen[other.ID] = true
					s.recordEvent(ctx, other, "REMOVED", string(other.State), string(ont.StateRemoved), "posisi diambil alih ONU lain")
					result.Removed++
				}
			}
		}

		before := cur
		strikes := 0
		if !row.LinkUp {
			strikes = cur.OfflineStrikes + 1
		}
		newState := ont.DecideState(cur.State, row.LinkUp, strikes, cfg.offlineStrikes)
		cur.PONIndex = row.PONIndex
		cur.ONUIndex = row.ONUIndex
		cur.LLID = intPtr(row.LLID)
		cur.Name = row.Name
		cur.State = newState
		cur.OfflineStrikes = strikes
		cur.MissedPolls = 0
		cur.RxPowerDBM = row.RxPowerDBM
		cur.TxPowerDBM = row.TxPowerDBM
		cur.DistanceM = row.DistanceM
		cur.LastPolledAt = &now
		if row.LinkUp {
			cur.LastSeenAt = &now
		}
		if row.RxPowerDBM == nil {
			cur.RxNAStrikes++
		} else {
			cur.RxNAStrikes = 0
		}
		if err := s.deps.ONTs.Save(ctx, cur); err != nil {
			return result, fmt.Errorf("olt: simpan ONT %s: %w", cur.MACAddress, err)
		}
		result.Updated++
		if cur.State == ont.StateOnline {
			result.Online++
		} else {
			result.Offline++
		}
		if before.State != cur.State {
			s.recordTransition(ctx, before, cur)
			if s.deps.Alerts != nil {
				s.deps.Alerts.ONTStateChanged(ctx, before, cur)
			}
		}
		if s.deps.Alerts != nil {
			beforeClass := ont.ClassifyRx(before.RxPowerDBM, cfg.rxWarning, cfg.rxCritical)
			afterClass := ont.ClassifyRx(cur.RxPowerDBM, cfg.rxWarning, cfg.rxCritical)
			if beforeClass != afterClass {
				s.deps.Alerts.ONTRxChanged(ctx, before, cur)
			}
		}
		s.addSample(ctx, cur, now)
	}

	// Roster carry-forward: ONT yang tidak terlihat tetap disimpan sampai
	// ambang miss tercapai (mencegah status berkedip karena packet loss).
	for _, cur := range existing {
		if seen[cur.ID] || cur.State == ont.StateRemoved {
			continue
		}
		cur.MissedPolls++
		if cur.MissedPolls >= cfg.missThreshold {
			before := cur.State
			cur.State = ont.StateRemoved
			if err := s.deps.ONTs.Save(ctx, cur); err != nil {
				return result, fmt.Errorf("olt: tandai ONT hilang %s: %w", cur.MACAddress, err)
			}
			s.recordEvent(ctx, cur, "REMOVED", string(before), string(ont.StateRemoved), "tidak terlihat pada polling OLT")
			result.Removed++
			if s.deps.Alerts != nil {
				s.deps.Alerts.ONTStateChanged(ctx, ont.ONT{ID: cur.ID, State: ont.State(before)}, cur)
			}
			continue
		}
		if err := s.deps.ONTs.Save(ctx, cur); err != nil {
			return result, fmt.Errorf("olt: simpan ONT tidak terlihat %s: %w", cur.MACAddress, err)
		}
	}
	return result, nil
}

// SyncAll menyinkronkan semua OLT HiOSO yang enabled (untuk scheduler).
func (s *Service) SyncAll(ctx context.Context) (map[string]SyncResult, error) {
	devices, err := s.deps.Devices.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	out := make(map[string]SyncResult)
	var firstErr error
	for _, dev := range devices {
		if dev.DriverType != driverTypeHioso || !dev.Enabled {
			continue
		}
		res, errSync := s.SyncONUs(ctx, dev.ID)
		if errSync != nil {
			if firstErr == nil {
				firstErr = fmt.Errorf("olt %s: %w", dev.Name, errSync)
			}
			continue
		}
		out[dev.ID] = res
	}
	return out, firstErr
}

func (s *Service) createONT(ctx context.Context, dev deviceRef, row port.OLTONU, now time.Time, cfg syncConfig) (ont.ONT, error) {
	state := ont.StateOnline
	strikes := 0
	if !row.LinkUp {
		state = ont.StateUnknown
		strikes = 1
		if strikes >= cfg.offlineStrikes {
			state = ont.StateOffline
		}
	}
	o := ont.ONT{
		ID:             idgen.New("ont"),
		TenantID:       dev.TenantID,
		OLTDeviceID:    dev.ID,
		PONIndex:       row.PONIndex,
		ONUIndex:       row.ONUIndex,
		LLID:           intPtr(row.LLID),
		Name:           row.Name,
		MACAddress:     row.MACAddress,
		State:          state,
		AdminState:     ont.AdminEnable,
		RxPowerDBM:     row.RxPowerDBM,
		TxPowerDBM:     row.TxPowerDBM,
		DistanceM:      row.DistanceM,
		OfflineStrikes: strikes,
		LastSeenAt:     &now,
		LastPolledAt:   &now,
		Source:         "SNMP",
	}
	if row.RxPowerDBM == nil {
		o.RxNAStrikes = 1
	}
	if err := s.deps.ONTs.Save(ctx, o); err != nil {
		return ont.ONT{}, fmt.Errorf("olt: simpan ONT baru %s: %w", o.MACAddress, err)
	}
	eventType := "DISCOVERED"
	if state == ont.StateOnline {
		eventType = "DISCOVERED_ONLINE"
	}
	s.recordEvent(ctx, o, eventType, "", string(state), fmt.Sprintf("ONU terdeteksi di PON %d ONU %d", o.PONIndex, o.ONUIndex))
	if s.deps.Alerts != nil && state != ont.StateOnline {
		s.deps.Alerts.ONTStateChanged(ctx, ont.ONT{ID: o.ID, State: ont.StateUnknown}, o)
	}
	s.addSample(ctx, o, now)
	return o, nil
}

// deviceRef adalah subset device.Device yang dibutuhkan sync (memudahkan test).
type deviceRef struct {
	ID       string
	TenantID string
}

func (s *Service) recordTransition(ctx context.Context, before, after ont.ONT) {
	switch {
	case after.State == ont.StateOnline && before.State != ont.StateOnline:
		s.recordEvent(ctx, after, "ONLINE", string(before.State), string(after.State), "ONU kembali online")
	case after.State == ont.StateOffline && before.State != ont.StateOffline:
		s.recordEvent(ctx, after, "OFFLINE", string(before.State), string(after.State), "ONU terdeteksi offline/LOS")
	}
}

func (s *Service) recordEvent(ctx context.Context, o ont.ONT, eventType, oldState, newState, message string) {
	err := s.deps.ONTs.RecordEvent(ctx, ont.Event{
		ID:        idgen.New("ontev"),
		TenantID:  o.TenantID,
		ONTID:     o.ID,
		EventType: eventType,
		OldState:  oldState,
		NewState:  newState,
		Message:   message,
	})
	_ = err // event timeline bersifat best-effort; jangan gagalkan sync karenanya
}

func (s *Service) addSample(ctx context.Context, o ont.ONT, now time.Time) {
	if o.State != ont.StateOnline {
		return
	}
	_ = s.deps.ONTs.AddSample(ctx, ont.MetricSample{
		ID:           idgen.New("ontsmp"),
		ONTID:        o.ID,
		SampledAt:    now,
		RxPowerDBM:   o.RxPowerDBM,
		TxPowerDBM:   o.TxPowerDBM,
		TemperatureC: o.TemperatureC,
		DistanceM:    o.DistanceM,
		Online:       true,
	})
}

func intPtr(v int) *int {
	if v == 0 {
		return nil
	}
	return &v
}

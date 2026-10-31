package hioso

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gosnmp/gosnmp"

	"github.com/quixiq/polyglot/internal/domain/olt"
	"github.com/quixiq/polyglot/internal/domain/ont"
	"github.com/quixiq/polyglot/internal/port"
)

// pduTransport adalah irisan gosnmp yang dipakai driver; dipisah agar bisa
// di-fake di unit test tanpa SNMP socket.
type pduTransport interface {
	Get(oids []string) (*gosnmp.SnmpPacket, error)
	Walk(rootOID string) ([]gosnmp.SnmpPDU, error)
	Set(pdus []gosnmp.SnmpPDU) error
	Close() error
}

type goSNMPTransport struct {
	conn *gosnmp.GoSNMP
}

// newSNMPTransport membuat (dan connect) transport SNMP v2c ke OLT.
func newSNMPTransport(host string, port int, community string, timeout time.Duration, retries int) (pduTransport, error) {
	if port <= 0 {
		port = 161
	}
	if timeout < 5*time.Second {
		// OLT HiOSO lambat menjawab walk; <5s terbukti rawan timeout parsial.
		timeout = 5 * time.Second
	}
	if retries < 2 {
		retries = 2
	}
	conn := &gosnmp.GoSNMP{
		Target:         host,
		Port:           uint16(port),
		Transport:      "udp",
		Community:      community,
		Version:        gosnmp.Version2c,
		Timeout:        timeout,
		Retries:        retries,
		MaxRepetitions: 20,
	}
	if err := conn.Connect(); err != nil {
		return nil, fmt.Errorf("hioso: snmp connect: %w", err)
	}
	return &goSNMPTransport{conn: conn}, nil
}

func (g *goSNMPTransport) Get(oids []string) (*gosnmp.SnmpPacket, error) {
	return g.conn.Get(oids)
}

func (g *goSNMPTransport) Walk(rootOID string) ([]gosnmp.SnmpPDU, error) {
	return g.conn.WalkAll(rootOID)
}

func (g *goSNMPTransport) Set(pdus []gosnmp.SnmpPDU) error {
	_, err := g.conn.Set(pdus)
	return err
}

func (g *goSNMPTransport) Close() error {
	return g.conn.Conn.Close()
}

// SNMPClient membaca inventori ONU dan identitas OLT HiOSO via SNMP.
type SNMPClient struct {
	t pduTransport
}

func newSNMPClient(t pduTransport) *SNMPClient {
	return &SNMPClient{t: t}
}

// SysInfo membaca sysDescr/sysObjectID/sysUpTime OLT (MIB-II, vendor-neutral).
func (c *SNMPClient) SysInfo(ctx context.Context) (sysDescr, sysObjectID, uptime string, err error) {
	if err := ctx.Err(); err != nil {
		return "", "", "", err
	}
	packet, err := c.t.Get([]string{oidSysDescr, oidSysObjectID, oidSysUptime})
	if err != nil {
		return "", "", "", fmt.Errorf("hioso: snmp get system: %w", err)
	}
	for _, pdu := range packet.Variables {
		switch pdu.Name {
		case oidSysDescr:
			sysDescr = pduString(pdu)
		case oidSysObjectID:
			sysObjectID = pduOIDString(pdu)
		case oidSysUptime:
			uptime = formatUptime(pduUint(pdu))
		}
	}
	return sysDescr, sysObjectID, uptime, nil
}

// FirmwareSignature membaca string signature firmware (deteksi varian).
func (c *SNMPClient) FirmwareSignature(ctx context.Context) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	// Beberapa firmware mengekspos kolom sebagai scalar .0, sebagian sebagai
	// tabel dengan instance .1 — coba Get dulu, fallback Walk.
	for _, suffix := range []string{".0", ".1"} {
		packet, err := c.t.Get([]string{oidFirmwareSignature + suffix})
		if err != nil {
			continue
		}
		for _, pdu := range packet.Variables {
			if v := pduString(pdu); v != "" {
				return v, nil
			}
		}
	}
	pdus, err := c.t.Walk(oidFirmwareSignature)
	if err != nil {
		return "", fmt.Errorf("hioso: snmp walk firmware signature: %w", err)
	}
	if len(pdus) == 0 {
		return "", nil
	}
	return pduString(pdus[0]), nil
}

// ListONUs membaca seluruh ONU satu PON (ponIndex 0 = semua PON) dengan
// walk per kolom. Hasil dipetakan berdasarkan index varian.
func (c *SNMPClient) ListONUs(ctx context.Context, variant olt.Variant, ponIndex int) ([]port.OLTONU, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	names, err := c.walkColumn(ctx, oidONUName)
	if err != nil {
		return nil, err
	}
	macs, err := c.walkColumn(ctx, oidONUMAC)
	if err != nil {
		return nil, err
	}
	links, err := c.walkColumn(ctx, oidONULinkState)
	if err != nil {
		return nil, err
	}
	rx, err := c.walkColumn(ctx, oidONURxPower)
	if err != nil {
		return nil, err
	}
	tx, err := c.walkColumn(ctx, oidONUTxPower)
	if err != nil {
		return nil, err
	}
	dist, err := c.walkColumn(ctx, oidONUDistance)
	if err != nil {
		return nil, err
	}
	llids, err := c.walkColumn(ctx, oidONULLID)
	if err != nil {
		return nil, err
	}

	// Gabungkan berdasarkan suffix index; roster utama = kolom MAC (identitas).
	out := make([]port.OLTONU, 0, len(macs))
	for suffix, rawMAC := range macs {
		pon, onu, ok := parseIndex(variant, suffix)
		if !ok {
			continue
		}
		if ponIndex != 0 && pon != ponIndex {
			continue
		}
		mac, ok := ont.NormalizeMAC(rawMAC)
		if !ok {
			continue
		}
		row := port.OLTONU{
			PONIndex:    pon,
			ONUIndex:    onu,
			MACAddress:  mac,
			Name:        names[suffix],
			LinkUp:      pduMapValueIs(links, suffix, linkStateUp),
			AdminEnable: true,
		}
		if v, ok := ont.ParseOpticalPower(rx[suffix]); ok {
			row.RxPowerDBM = v
		}
		if v, ok := ont.ParseOpticalPower(tx[suffix]); ok {
			row.TxPowerDBM = v
		}
		if v, ok := parseIntPointer(dist[suffix]); ok {
			row.DistanceM = v
		}
		if v, ok := parseIntPointer(llids[suffix]); ok {
			row.LLID = *v
		}
		out = append(out, row)
	}
	// Walk SNMP dikumpulkan lewat map suffix→nilai sehingga urutannya acak;
	// hasil inventori harus deterministik (PON lalu ONU) untuk usecase/UI.
	sort.Slice(out, func(i, j int) bool {
		if out[i].PONIndex != out[j].PONIndex {
			return out[i].PONIndex < out[j].PONIndex
		}
		return out[i].ONUIndex < out[j].ONUIndex
	})
	return out, nil
}

// walkColumn melakukan walk satu kolom ONU lalu memetakan suffix index → nilai string.
func (c *SNMPClient) walkColumn(ctx context.Context, root string) (map[string]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	pdus, err := c.t.Walk(root)
	if err != nil {
		return nil, fmt.Errorf("hioso: snmp walk %s: %w", root, err)
	}
	out := make(map[string]string, len(pdus))
	prefix := root + "."
	for _, pdu := range pdus {
		suffix := strings.TrimPrefix(pdu.Name, prefix)
		if suffix == pdu.Name {
			continue
		}
		out[suffix] = pduString(pdu)
	}
	return out, nil
}

// SetONUName mengganti nama ONU lewat SNMP SET (dipakai HA7302, yang tidak
// mendukung rename via CLI pada firmware tertentu). Membutuhkan community RW.
func (c *SNMPClient) SetONUName(ctx context.Context, variant olt.Variant, ponIndex, onuIndex int, name string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	oid := oidONUName + "." + indexOf(variant, ponIndex, onuIndex)
	err := c.t.Set([]gosnmp.SnmpPDU{{
		Name:  oid,
		Type:  gosnmp.OctetString,
		Value: name,
	}})
	if err != nil {
		return fmt.Errorf("hioso: snmp set onu name: %w", err)
	}
	return nil
}

// Close menutup koneksi SNMP.
func (c *SNMPClient) Close() error {
	if c == nil || c.t == nil {
		return nil
	}
	return c.t.Close()
}

// --- helper parsing PDU ---

func pduString(pdu gosnmp.SnmpPDU) string {
	switch v := pdu.Value.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(v)
	case []byte:
		return strings.TrimSpace(string(v))
	case int:
		return strconv.Itoa(v)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case int64:
		return strconv.FormatInt(v, 10)
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case uint32:
		return strconv.FormatUint(uint64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	default:
		return strings.TrimSpace(fmt.Sprintf("%v", v))
	}
}

func pduOIDString(pdu gosnmp.SnmpPDU) string {
	switch v := pdu.Value.(type) {
	case string:
		return v
	default:
		return pduString(pdu)
	}
}

func pduUint(pdu gosnmp.SnmpPDU) uint64 {
	switch v := pdu.Value.(type) {
	case uint:
		return uint64(v)
	case uint32:
		return uint64(v)
	case uint64:
		return v
	case int:
		return uint64(v)
	case int64:
		return uint64(v)
	case string:
		n, _ := strconv.ParseUint(strings.TrimSpace(v), 10, 64)
		return n
	default:
		return 0
	}
}

func pduMapValueIs(m map[string]string, suffix string, want int) bool {
	raw, ok := m[suffix]
	if !ok {
		return false
	}
	n, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return false
	}
	return n == want
}

func parseIntPointer(raw string) (*int, bool) {
	raw = strings.TrimSpace(strings.Trim(raw, `"`))
	if raw == "" || strings.EqualFold(raw, "na") {
		return nil, false
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return nil, false
	}
	return &n, true
}

// formatUptime mengubah sysUpTime ticks (1/100 detik) ke string durasi ringkas.
func formatUptime(ticks uint64) string {
	if ticks == 0 {
		return ""
	}
	d := time.Duration(ticks) * 10 * time.Millisecond
	return d.Truncate(time.Second).String()
}

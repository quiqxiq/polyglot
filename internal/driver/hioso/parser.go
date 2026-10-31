package hioso

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/quixiq/polyglot/internal/domain/olt"
	"github.com/quixiq/polyglot/internal/domain/ont"
	"github.com/quixiq/polyglot/internal/port"
)

var (
	cliMACRe  = regexp.MustCompile(`(?i)\b[0-9a-f]{2}([:-][0-9a-f]{2}){5}\b`)
	cliPosRe  = regexp.MustCompile(`^\d+/\d+/\d+$`)
	cliFloat  = regexp.MustCompile(`^-?\d+\.\d+$`)
	cliIntRe  = regexp.MustCompile(`^\d+$`)
	stateWord = map[string]bool{
		"up": true, "online": true, "active": true, "working": true,
		"down": true, "offline": true, "inactive": true, "los": true, "failed": true,
	}
	upWord = map[string]bool{"up": true, "online": true, "active": true, "working": true}
)

// parseShowONUInfo mem-parsing output `show onu info epon 0/{PON} all`.
//
// Format antar firmware bervariasi; parser bersifat toleran: mencari MAC
// sebagai penanda baris, lalu mengambil nama di kiri MAC, status online dari
// kata kunci state, dan angka setelah state sebagai distance/Rx/Tx.
func parseShowONUInfo(output string, variant olt.Variant) []port.OLTONU {
	var rows []port.OLTONU
	for _, raw := range strings.Split(strings.ReplaceAll(output, "\r\n", "\n"), "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.Contains(line, "EPON") {
			continue
		}
		macLoc := cliMACRe.FindStringIndex(line)
		if macLoc == nil {
			continue
		}
		mac, ok := ont.NormalizeMAC(line[macLoc[0]:macLoc[1]])
		if !ok {
			continue
		}
		left := strings.Fields(line[:macLoc[0]])
		right := strings.Fields(line[macLoc[1]:])

		row := port.OLTONU{MACAddress: mac, AdminEnable: true}
		// Posisi ONU: token "0/1/5" lebih disukai; fallback integer terakhir di kiri.
		for _, tok := range left {
			if cliPosRe.MatchString(tok) {
				parts := strings.Split(tok, "/")
				row.PONIndex, _ = strconv.Atoi(parts[1])
				row.ONUIndex, _ = strconv.Atoi(parts[2])
				break
			}
		}
		if row.ONUIndex == 0 {
			for i := len(left) - 1; i >= 0; i-- {
				if cliIntRe.MatchString(left[i]) {
					row.ONUIndex, _ = strconv.Atoi(left[i])
					break
				}
			}
		}
		if row.PONIndex == 0 {
			row.PONIndex = 1
		}
		if variant == olt.VariantHA7302 {
			row.PONIndex = 1
		}
		// Nama: token non-angka tepat sebelum MAC.
		for i := len(left) - 1; i >= 0; i-- {
			tok := left[i]
			if !cliIntRe.MatchString(tok) && !cliFloat.MatchString(tok) && !stateWord[strings.ToLower(tok)] {
				row.Name = tok
				break
			}
		}

		stateIdx := -1
		for i, tok := range right {
			if stateWord[strings.ToLower(tok)] {
				stateIdx = i
				row.LinkUp = upWord[strings.ToLower(tok)]
				break
			}
		}
		start := 0
		if stateIdx >= 0 {
			start = stateIdx + 1
		}
		numbers := right[start:]
		// Urutan pasca-state: distance (int), Rx (float), Tx (float).
		var floats []float64
		intSeen := false
		for _, tok := range numbers {
			if cliIntRe.MatchString(tok) && !intSeen && len(floats) == 0 {
				n, _ := strconv.Atoi(tok)
				row.DistanceM = &n
				intSeen = true
				continue
			}
			if cliFloat.MatchString(tok) || (cliIntRe.MatchString(tok) && len(floats) > 0) {
				f, err := strconv.ParseFloat(tok, 64)
				if err == nil {
					floats = append(floats, f)
				}
			}
		}
		if len(floats) > 0 {
			row.RxPowerDBM = &floats[0]
		}
		if len(floats) > 1 {
			row.TxPowerDBM = &floats[1]
		}
		rows = append(rows, row)
	}
	return rows
}

// parseKeyValue mem-parsing output `show version`/`show system` bergaya
// "Key : value" atau "Key    value".
func parseKeyValue(output string) map[string]string {
	out := map[string]string{}
	for _, raw := range strings.Split(strings.ReplaceAll(output, "\r\n", "\n"), "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.Contains(line, "EPON") {
			continue
		}
		if i := strings.Index(line, ":"); i > 0 {
			k := strings.TrimSpace(line[:i])
			v := strings.TrimSpace(line[i+1:])
			if k != "" && v != "" {
				out[k] = v
			}
			continue
		}
		if fields := strings.Fields(line); len(fields) >= 2 {
			out[fields[0]] = strings.Join(fields[1:], " ")
		}
	}
	return out
}

// commandRejected mendeteksi output error CLI HiOSO pada baris awal output.
func commandRejected(output string) bool {
	re := regexp.MustCompile(`(?im)^\s*(error|invalid|unrecognized|unknown command|failed|failure|permission denied|wrong|%)\b`)
	lines := strings.Split(strings.ReplaceAll(output, "\r\n", "\n"), "\n")
	if len(lines) > 6 {
		lines = lines[:6]
	}
	return re.MatchString(strings.Join(lines, "\n"))
}

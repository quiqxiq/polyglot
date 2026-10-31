package rosutil

import (
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Reply represents a RouterOS sentence key-value attribute map.
type Reply = map[string]string

// Cmd represents a low-level wire command sent to a RouterOS device.
type Cmd struct {
	Path     string
	Args     []string
	Timeout  time.Duration
	Finished func()
	Ret      *string
	Done     *map[string]string
}

// SetIfNonEmpty assigns value to args[key] only if value is non-empty.
func SetIfNonEmpty(args map[string]string, key, value string) {
	if value != "" {
		args[key] = value
	}
}

// SetProplist sets the .proplist argument in args by joining the specified field names.
func SetProplist(args map[string]string, fields ...string) {
	if len(fields) > 0 {
		args[".proplist"] = strings.Join(fields, ",")
	}
}

var pingRTTPart = regexp.MustCompile(`(\d*\.?\d+)(us|ms|s|m|h|d|w)`)
var pingRTTBare = regexp.MustCompile(`^\d*\.?\d+$`)

var pingRTTUs = map[string]float64{
	"us": 1, "ms": 1e3, "s": 1e6, "m": 60e6, "h": 3600e6, "d": 86400e6, "w": 604800e6,
}

// ParsePingRTT parses RouterOS RTT duration string into milliseconds.
func ParsePingRTT(val string) *float64 {
	val = strings.TrimSpace(val)
	if val == "" {
		return nil
	}
	if pingRTTBare.MatchString(val) {
		v, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return nil
		}
		return &v
	}
	us, covered := 0.0, 0
	for _, m := range pingRTTPart.FindAllStringSubmatchIndex(val, -1) {
		if m[0] != covered {
			return nil
		}
		n, err := strconv.ParseFloat(val[m[2]:m[3]], 64)
		if err != nil {
			return nil
		}
		us += n * pingRTTUs[val[m[4]:m[5]]]
		covered = m[1]
	}
	if covered == 0 || covered != len(val) {
		return nil
	}
	v := math.Round(us) / 1000
	return &v
}

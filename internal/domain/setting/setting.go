package setting

import "time"

// Setting represents a key-value configuration item in the system.
type Setting struct {
	Key         string    `json:"key"`
	Value       string    `json:"value"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// BotSettings represents typed configuration parameters for the customer service bot.
type BotSettings struct {
	BurstLimit            int    `json:"burst_limit"`
	BurstWindowSecs       int    `json:"burst_window_secs"`
	Mute1HourSecs         int    `json:"mute_1h_secs"`
	Ban24HourSecs         int    `json:"ban_24h_secs"`
	DailyChatLimit        int    `json:"daily_chat_limit"`
	SessionTimeoutMinutes int    `json:"session_timeout_minutes"`
	SlidingWindowSize     int    `json:"sliding_window_size"`
	LLMMaxOutputTokens    int    `json:"llm_max_output_tokens"`
	WhitelistAllStaff     bool   `json:"whitelist_all_staff"`
	CustomWhitelistPhones string `json:"custom_whitelist_phones"`
}

// DefaultBotSettings returns recommended baseline configuration.
func DefaultBotSettings() *BotSettings {
	return &BotSettings{
		BurstLimit:            3,
		BurstWindowSecs:       5,
		Mute1HourSecs:         3600,
		Ban24HourSecs:         86400,
		DailyChatLimit:        10,
		SessionTimeoutMinutes: 30,
		SlidingWindowSize:     10,
		LLMMaxOutputTokens:    1024,
		WhitelistAllStaff:     true,
		CustomWhitelistPhones: "",
	}
}

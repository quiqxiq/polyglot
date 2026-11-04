package mikrotik

// MenuChildItem represents a child menu, command, or parameter in the RouterOS CLI hierarchy.
type MenuChildItem struct {
	Name     string `json:"name"`
	Type     string `json:"type,omitempty"`
	NodeType string `json:"node_type,omitempty"`
	Help     string `json:"help,omitempty"`
}

// CommandArgument represents an argument for a RouterOS command with its syntax and type.
type CommandArgument struct {
	Name        string   `json:"name"`
	Type        string   `json:"type,omitempty"`
	Values      []string `json:"values,omitempty"`
	Default     string   `json:"default,omitempty"`
	Required    bool     `json:"required,omitempty"`
	Description string   `json:"description,omitempty"`
}

// SuggestionItem represents an autocomplete candidate or dynamic value returned by RouterOS.
type SuggestionItem struct {
	Completion string `json:"completion"`
	Style      string `json:"style,omitempty"`
	Preference string `json:"preference,omitempty"`
}

// ValidationResult represents the dry-run syntax check output from RouterOS /console/inspect.
type ValidationResult struct {
	Valid         bool   `json:"valid"`
	ErrorMessage  string `json:"error_message,omitempty"`
	ErrorPosition int    `json:"error_position,omitempty"`
}

// routerosListMenuArgs is the input schema for routeros_list_menu.
type routerosListMenuArgs struct {
	DeviceID string `json:"device_id" jsonschema:"the Mikrotik device ID from inventory"`
	Path     string `json:"path,omitempty" jsonschema:"menu path to inspect (e.g. 'ip/address', 'ip,firewall', or empty for root)"`
}

// routerosListMenuOutput is the output schema for routeros_list_menu.
type routerosListMenuOutput struct {
	DeviceID string          `json:"device_id"`
	Path     string          `json:"path"`
	Items    []MenuChildItem `json:"items"`
	Count    int             `json:"count"`
	Cached   bool            `json:"cached"`
	Summary  string          `json:"summary"`
	Error    string          `json:"error,omitempty"`
}

// routerosDescribeCommandArgs is the input schema for routeros_describe_command.
type routerosDescribeCommandArgs struct {
	DeviceID string `json:"device_id" jsonschema:"the Mikrotik device ID from inventory"`
	Path     string `json:"path" jsonschema:"command path to describe (e.g. 'ip/address/add' or 'ip,firewall,filter,add')"`
}

// routerosDescribeCommandOutput is the output schema for routeros_describe_command.
type routerosDescribeCommandOutput struct {
	DeviceID  string            `json:"device_id"`
	Path      string            `json:"path"`
	Arguments []CommandArgument `json:"arguments"`
	Count     int               `json:"count"`
	Cached    bool              `json:"cached"`
	Summary   string            `json:"summary"`
	Error     string            `json:"error,omitempty"`
}

// routerosSuggestArgs is the input schema for routeros_suggest.
type routerosSuggestArgs struct {
	DeviceID string `json:"device_id" jsonschema:"the Mikrotik device ID from inventory"`
	Input    string `json:"input" jsonschema:"console input line being typed (e.g. '/ip address add interface=')"`
}

// routerosSuggestOutput is the output schema for routeros_suggest.
type routerosSuggestOutput struct {
	DeviceID    string           `json:"device_id"`
	Input       string           `json:"input"`
	Suggestions []SuggestionItem `json:"suggestions"`
	Count       int              `json:"count"`
	Summary     string           `json:"summary"`
	Error       string           `json:"error,omitempty"`
}

// routerosValidateArgs is the input schema for routeros_validate.
type routerosValidateArgs struct {
	DeviceID string `json:"device_id" jsonschema:"the Mikrotik device ID from inventory"`
	Input    string `json:"input" jsonschema:"full console command line to validate without execution (e.g. '/ip address add address=10.0.0.1/99 interface=ether1')"`
}

// routerosValidateOutput is the output schema for routeros_validate.
type routerosValidateOutput struct {
	DeviceID      string `json:"device_id"`
	Input         string `json:"input"`
	Valid         bool   `json:"valid"`
	ErrorMessage  string `json:"error_message,omitempty"`
	ErrorPosition int    `json:"error_position,omitempty"`
	Summary       string `json:"summary"`
	Remediation   string `json:"remediation,omitempty"`
	Error         string `json:"error,omitempty"`
}

// routerosRunArgs is the input schema for routeros_run.
type routerosRunArgs struct {
	DeviceID     string            `json:"device_id" jsonschema:"the Mikrotik device ID from inventory"`
	Command      string            `json:"command" jsonschema:"the RouterOS API path (e.g. /ip/address/add) or console command line"`
	Args         map[string]string `json:"args,omitempty" jsonschema:"optional key-value parameters for the command"`
	AutoValidate *bool             `json:"auto_validate,omitempty" jsonschema:"dry-run validate command syntax before execution (defaults to true)"`
}

// routerosRunOutput is the output schema for routeros_run.
type routerosRunOutput struct {
	DeviceID        string              `json:"device_id"`
	Command         string              `json:"command"`
	Validated       bool                `json:"validated"`
	Executed        bool                `json:"executed"`
	ValidationError string              `json:"validation_error,omitempty"`
	Output          string              `json:"output,omitempty"`
	Rows            []map[string]string `json:"rows,omitempty"`
	RowCount        int                 `json:"row_count,omitempty"`
	Remediation     string              `json:"remediation,omitempty"`
	Error           string              `json:"error,omitempty"`
}

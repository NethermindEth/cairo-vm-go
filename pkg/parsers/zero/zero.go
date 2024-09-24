package zero

import (
	"encoding/json"
	"os"

	builtins "github.com/NethermindEth/cairo-vm-go/pkg/vm/builtins"
	
)

type FlowTrackingData struct {
	ApTracking   ApTracking        `json:"ap_tracking"`
	ReferenceIds map[string]uint64 `json:"reference_ids"`
}

type ApTracking struct {
	Group  int `json:"group"`
	Offset int `json:"offset"`
}

type Location struct {
	EndCol    uint64            `json:"end_col"`
	EndLine   uint64            `json:"end_line"`
	InputFile map[string]string `json:"input_file"`
	StartCol  uint64            `json:"start_col"`
	StartLine uint64            `json:"start_line"`
}

type HintLocation struct {
	Location        Location `json:"location"`
	NPrefixNewlines uint64   `json:"n_prefix_newlines"`
}

type InstructionLocation struct {
	AccessibleScopes []string         `json:"accessible_scopes"`
	FlowTrackingData FlowTrackingData `json:"flow_tracking_data"`
	Hints            []HintLocation   `json:"hints"`
	Inst             Location         `json:"inst"`
}

type DebugInfo struct {
	FileContents         map[string]string              `json:"file_contents"`
	InstructionLocations map[string]InstructionLocation `json:"instruction_locations"`
	SourceCode           string                         `json:"source_code"`
	SourcePath           string                         `json:"source_path"`
}

type Hint struct {
	AccessibleScopes []string         `json:"accessible_scopes"`
	Code             string           `json:"code"`
	FlowTrackingData FlowTrackingData `json:"flow_tracking_data"`
}

type Reference struct {
	ApTrackingData ApTracking `json:"ap_tracking_data"`
	Pc             uint64     `json:"pc"`
	Value          string     `json:"value"`
}

type ReferenceManager struct {
	References []Reference
}

type AttributeScope struct {
	StartPc          uint64           `json:"start_pc"`
	EndPc            uint64           `json:"end_pc"`
	FlowTrackingData FlowTrackingData `json:"flow_tracking_data"`
	AccessibleScopes []string         `json:"accessible_scopes"`
}

type ZeroProgram struct {
	Prime            string                   `json:"prime"`
	Data             []string                 `json:"data"`
	Builtins         []builtins.Builtin		  `json:"builtins"`
	Hints            map[string][]Hint        `json:"hints"`
	CompilerVersion  string                   `json:"version"`
	MainScope        string                   `json:"main_scope"`
	Identifiers      map[string]*Identifier   `json:"identifiers"`
	ReferenceManager ReferenceManager         `json:"reference_manager"`
	Attributes       []AttributeScope         `json:"attributes"`
	DebugInfo        DebugInfo                `json:"debug_info"`
}

type Identifier struct {
	FullName       string         `json:"full_name"`
	IdentifierType string         `json:"type"`
	CairoType      string         `json:"cairo_type"`
	Destination    string         `json:"destination"`
	Pc             int            `json:"pc"`
	Size           int            `json:"size"`
	Members        map[string]any `json:"members"`
	References     []Reference    `json:"references"`

	// These fields are listed as any-typed fields before we need them.
	Decorators any `json:"decorators"`
	Value      any `json:"value"`
}

func (z ZeroProgram) MarshalToFile(filepath string) error {
	// Marshal Output struct into JSON bytes
	data, err := json.MarshalIndent(z, "", "    ")
	if err != nil {
		return err
	}

	// Write JSON bytes to file
	err = os.WriteFile(filepath, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

func ZeroProgramFromFile(filepath string) (zero *ZeroProgram, err error) {
	content, err := os.ReadFile(filepath)
	if err != nil {
		return
	}
	return ZeroProgramFromJSON(content)
}

func ZeroProgramFromJSON(content json.RawMessage) (*ZeroProgram, error) {
	var zero ZeroProgram
	return &zero, json.Unmarshal(content, &zero)
}

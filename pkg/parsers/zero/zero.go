package zero

import (
	"encoding/json"
	"os"
)

type ZeroProgram struct {
	Attributes []string          `json:"attributes"`
	Builtins   map[string]int64  `json:"builtins"`
	Code       string            `json:"code"`
	DebugInfo  DebugInfo         `json:"debug_info"`
	Prime      string            `json:"prime"`
	Version    string            `json:"version"`
	Hints      map[string][]Hint `json:"hints"`
}

type DebugInfo struct {
	InstructionOffsets []int64 `json:"instruction_offsets"`
	SourceCode         string  `json:"source_code"`
	SourcePath         string  `json:"source_path"`
}

type Hint struct {
	AccessibleScopes []string         `json:"accessible_scopes"`
	Code             string           `json:"code"`
	FlowTrackingData FlowTrackingData `json:"flow_tracking_data"`
}

type FlowTrackingData struct {
	ApTracking   ApTracking             `json:"ap_tracking"`
	ReferenceIds map[string]interface{} `json:"reference_ids"`
}

type ApTracking struct {
	Group  int `json:"group"`
	Offset int `json:"offset"`
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

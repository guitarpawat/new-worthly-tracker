package dto

import (
	"encoding/json"
	"testing"

	"github.com/shopspring/decimal"
)

func TestFinancialValuesJSONRoundTripAsDecimals(t *testing.T) {
	t.Parallel()

	var input SaveSnapshotItemInput
	if err := json.Unmarshal([]byte(`{"AssetID":1,"BoughtPrice":"0.10","CurrentPrice":0.20}`), &input); err != nil {
		t.Fatalf("unmarshal decimal input: %v", err)
	}
	if !input.BoughtPrice.Equal(decimal.RequireFromString("0.10")) {
		t.Fatalf("expected bought price 0.10, got %s", input.BoughtPrice)
	}
	if !input.CurrentPrice.Equal(decimal.RequireFromString("0.20")) {
		t.Fatalf("expected current price 0.20, got %s", input.CurrentPrice)
	}

	encoded, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("marshal decimal input: %v", err)
	}
	if string(encoded) != `{"AssetID":1,"BoughtPrice":"0.1","CurrentPrice":"0.2","Remarks":""}` {
		t.Fatalf("unexpected decimal JSON: %s", encoded)
	}
}

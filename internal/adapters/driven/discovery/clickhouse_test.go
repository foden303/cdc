package discovery

import "testing"

func TestIsClickhouseNullableType(t *testing.T) {
	tests := []struct {
		name     string
		dataType string
		want     bool
	}{
		{name: "nullable string", dataType: "Nullable(String)", want: true},
		{name: "nullable with leading space", dataType: " Nullable(UInt64)", want: true},
		{name: "plain string", dataType: "String", want: false},
		{name: "low cardinality nullable wrapper", dataType: "LowCardinality(Nullable(String))", want: true},
		{name: "array nullable element", dataType: "Array(Nullable(String))", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isClickhouseNullableType(tt.dataType); got != tt.want {
				t.Fatalf("isClickhouseNullableType(%q) = %v, want %v", tt.dataType, got, tt.want)
			}
		})
	}
}

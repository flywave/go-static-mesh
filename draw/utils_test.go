package draw

import (
	"math"
	"testing"
)

func TestDegreesToRadians(t *testing.T) {
	tests := []struct {
		name     string
		degrees  float64
		expected float64
	}{
		{
			name:     "zero",
			degrees:  0,
			expected: 0,
		},
		{
			name:     "90 degrees",
			degrees:  90,
			expected: math.Pi / 2,
		},
		{
			name:     "180 degrees",
			degrees:  180,
			expected: math.Pi,
		},
		{
			name:     "360 degrees",
			degrees:  360,
			expected: 2 * math.Pi,
		},
		{
			name:     "45 degrees",
			degrees:  45,
			expected: math.Pi / 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DegreesToRadians(tt.degrees)
			if math.Abs(result-tt.expected) > 1e-10 {
				t.Errorf("DegreesToRadians(%v) = %v, want %v", tt.degrees, result, tt.expected)
			}
		})
	}
}

func TestRadiansToDegrees(t *testing.T) {
	tests := []struct {
		name     string
		radians  float64
		expected float64
	}{
		{
			name:     "zero",
			radians:  0,
			expected: 0,
		},
		{
			name:     "pi/2",
			radians:  math.Pi / 2,
			expected: 90,
		},
		{
			name:     "pi",
			radians:  math.Pi,
			expected: 180,
		},
		{
			name:     "2*pi",
			radians:  2 * math.Pi,
			expected: 360,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RadiansToDegrees(tt.radians)
			if math.Abs(result-tt.expected) > 1e-10 {
				t.Errorf("RadiansToDegrees(%v) = %v, want %v", tt.radians, result, tt.expected)
			}
		})
	}
}

func TestParseLatLon(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantLat float64
		wantLng float64
		wantErr bool
	}{
		{
			name:    "decimal degrees",
			input:   "35.6895, 139.6917",
			wantLat: 35.6895,
			wantLng: 139.6917,
			wantErr: false,
		},
		{
			name:    "hemisphere decimal",
			input:   "N 35.6895 E 139.6917",
			wantLat: 35.6895,
			wantLng: 139.6917,
			wantErr: false,
		},
		{
			name:    "hemisphere decimal minute",
			input:   "N 35 41.37 E 139 41.502",
			wantLat: 35.6895,
			wantLng: 139.6917,
			wantErr: false,
		},
		{
			name:    "hemisphere degree minute second",
			input:   "N 35 41 22.32 E 139 41 30.12",
			wantLat: 35.689533,
			wantLng: 139.6917,
			wantErr: false,
		},
		{
			name:    "south and west",
			input:   "S 35.6895 W 139.6917",
			wantLat: -35.6895,
			wantLng: -139.6917,
			wantErr: false,
		},
		{
			name:    "invalid format",
			input:   "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lat, lng, err := ParseLatLon(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseLatLon() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if math.Abs(lat-tt.wantLat) > 0.001 {
					t.Errorf("ParseLatLon() lat = %v, want %v", lat, tt.wantLat)
				}
				if math.Abs(lng-tt.wantLng) > 0.001 {
					t.Errorf("ParseLatLon() lng = %v, want %v", lng, tt.wantLng)
				}
			}
		})
	}
}

func TestParseD(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantLat float64
		wantLng float64
		wantErr bool
	}{
		{
			name:    "comma separated",
			input:   "35.6895, 139.6917",
			wantLat: 35.6895,
			wantLng: 139.6917,
			wantErr: false,
		},
		{
			name:    "semicolon separated",
			input:   "35.6895; 139.6917",
			wantLat: 35.6895,
			wantLng: 139.6917,
			wantErr: false,
		},
		{
			name:    "space separated",
			input:   "35.6895 139.6917",
			wantLat: 35.6895,
			wantLng: 139.6917,
			wantErr: false,
		},
		{
			name:    "negative coordinates",
			input:   "-35.6895, -139.6917",
			wantLat: -35.6895,
			wantLng: -139.6917,
			wantErr: false,
		},
		{
			name:    "invalid latitude",
			input:   "91.0, 139.6917",
			wantErr: true,
		},
		{
			name:    "invalid longitude",
			input:   "35.6895, 181.0",
			wantErr: true,
		},
		{
			name:    "invalid format",
			input:   "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lat, lng, err := ParseD(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseD() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if math.Abs(lat-tt.wantLat) > 1e-10 {
					t.Errorf("ParseD() lat = %v, want %v", lat, tt.wantLat)
				}
				if math.Abs(lng-tt.wantLng) > 1e-10 {
					t.Errorf("ParseD() lng = %v, want %v", lng, tt.wantLng)
				}
			}
		})
	}
}

func TestParseHD(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantLat float64
		wantLng float64
		wantErr bool
	}{
		{
			name:    "north east",
			input:   "N 35.6895 E 139.6917",
			wantLat: 35.6895,
			wantLng: 139.6917,
			wantErr: false,
		},
		{
			name:    "south west",
			input:   "S 35.6895 W 139.6917",
			wantLat: -35.6895,
			wantLng: -139.6917,
			wantErr: false,
		},
		{
			name:    "lowercase",
			input:   "n 35.6895 e 139.6917",
			wantLat: 35.6895,
			wantLng: 139.6917,
			wantErr: false,
		},
		{
			name:    "invalid format",
			input:   "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lat, lng, err := ParseHD(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseHD() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if math.Abs(lat-tt.wantLat) > 1e-10 {
					t.Errorf("ParseHD() lat = %v, want %v", lat, tt.wantLat)
				}
				if math.Abs(lng-tt.wantLng) > 1e-10 {
					t.Errorf("ParseHD() lng = %v, want %v", lng, tt.wantLng)
				}
			}
		})
	}
}

func TestParseHDM(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantLat float64
		wantLng float64
		wantErr bool
	}{
		{
			name:    "north east",
			input:   "N 35 41.37 E 139 41.502",
			wantLat: 35.6895,
			wantLng: 139.6917,
			wantErr: false,
		},
		{
			name:    "south west",
			input:   "S 35 41.37 W 139 41.502",
			wantLat: -35.6895,
			wantLng: -139.6917,
			wantErr: false,
		},
		{
			name:    "invalid format",
			input:   "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lat, lng, err := ParseHDM(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseHDM() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if math.Abs(lat-tt.wantLat) > 0.001 {
					t.Errorf("ParseHDM() lat = %v, want %v", lat, tt.wantLat)
				}
				if math.Abs(lng-tt.wantLng) > 0.001 {
					t.Errorf("ParseHDM() lng = %v, want %v", lng, tt.wantLng)
				}
			}
		})
	}
}

func TestParseHDMS(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantLat float64
		wantLng float64
		wantErr bool
	}{
		{
			name:    "north east",
			input:   "N 35 41 22.32 E 139 41 30.12",
			wantLat: 35.689533,
			wantLng: 139.6917,
			wantErr: false,
		},
		{
			name:    "south west",
			input:   "S 35 41 22.32 W 139 41 30.12",
			wantLat: -35.689533,
			wantLng: -139.6917,
			wantErr: false,
		},
		{
			name:    "invalid format",
			input:   "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lat, lng, err := ParseHDMS(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseHDMS() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if math.Abs(lat-tt.wantLat) > 0.001 {
					t.Errorf("ParseHDMS() lat = %v, want %v", lat, tt.wantLat)
				}
				if math.Abs(lng-tt.wantLng) > 0.001 {
					t.Errorf("ParseHDMS() lng = %v, want %v", lng, tt.wantLng)
				}
			}
		})
	}
}

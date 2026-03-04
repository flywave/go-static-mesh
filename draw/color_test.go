package draw

import (
	"image/color"
	"math"
	"testing"
)

func TestParseColorString(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    color.Color
		wantErr bool
	}{
		{
			name:  "hex color with hash",
			input: "#ff0000",
			want:  color.RGBA{255, 0, 0, 255},
		},
		{
			name:  "hex color without hash",
			input: "0xff0000",
			want:  color.RGBA{255, 0, 0, 255},
		},
		{
			name:  "hex color uppercase",
			input: "#FF0000",
			want:  color.RGBA{255, 0, 0, 255},
		},
		{
			name:  "hex color with alpha",
			input: "#ff000080",
			want:  color.RGBA{127, 0, 0, 128},
		},
		{
			name:  "named color red",
			input: "red",
			want:  color.RGBA{255, 0, 0, 255},
		},
		{
			name:  "named color blue",
			input: "blue",
			want:  color.RGBA{0, 0, 255, 255},
		},
		{
			name:  "named color green",
			input: "green",
			want:  color.RGBA{0, 255, 0, 255},
		},
		{
			name:  "named color white",
			input: "white",
			want:  color.RGBA{255, 255, 255, 255},
		},
		{
			name:  "named color black",
			input: "black",
			want:  color.RGBA{0, 0, 0, 255},
		},
		{
			name:    "invalid color",
			input:   "invalidcolor",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseColorString(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseColorString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ParseColorString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLuminance(t *testing.T) {
	tests := []struct {
		name  string
		color color.Color
		want  float64
	}{
		{
			name:  "white",
			color: color.RGBA{255, 255, 255, 255},
			want:  1.0,
		},
		{
			name:  "black",
			color: color.RGBA{0, 0, 0, 255},
			want:  0.0,
		},
		{
			name:  "gray",
			color: color.RGBA{128, 128, 128, 255},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Luminance(tt.color)
			if got < 0 || got > 1 {
				t.Errorf("Luminance() = %v, should be between 0 and 1", got)
			}
			if tt.name == "white" && math.Abs(got-1.0) > 0.0001 {
				t.Errorf("Luminance() = %v, want %v", got, tt.want)
			}
			if tt.name == "black" && got != tt.want {
				t.Errorf("Luminance() = %v, want %v", got, tt.want)
			}
		})
	}
}

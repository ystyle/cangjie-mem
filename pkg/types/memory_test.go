package types

import "testing"

func TestCheckPackageVersion(t *testing.T) {
	tests := []struct {
		name    string
		version string
		wantErr bool
	}{
		{"valid version", PackageFormatVersion, false},
		{"empty version", "", true},
		{"unsupported version", "0.5", true},
		{"future version", "2.0", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckPackageVersion(tt.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckPackageVersion(%q) error = %v, wantErr %v", tt.version, err, tt.wantErr)
			}
		})
	}
}

package config

import (
	"strings"
	"testing"
)

func TestValidateProductionConfig(t *testing.T) {
	strongSecret := strings.Repeat("s", 40)
	weakSecret := "short"
	strongPassword := strings.Repeat("p", 12)

	tests := []struct {
		name    string
		mode    string
		secret  string
		pass    string
		wantErr bool
	}{
		{name: "release with strong config ok", mode: "release", secret: strongSecret, pass: strongPassword, wantErr: false},
		{name: "release with weak secret rejected", mode: "release", secret: weakSecret, pass: strongPassword, wantErr: true},
		{name: "release with empty secret rejected", mode: "release", secret: "", pass: strongPassword, wantErr: true},
		{name: "release with placeholder secret rejected", mode: "release", secret: placeholderSecret, pass: strongPassword, wantErr: true},
		{name: "release with empty password rejected", mode: "release", secret: strongSecret, pass: "", wantErr: true},
		{name: "debug skips validation", mode: "debug", secret: "", pass: "", wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if (r != nil) != tt.wantErr {
					t.Fatalf("validateProductionConfig() panic = %v, wantErr %v", r, tt.wantErr)
				}
			}()
			validateProductionConfig(&Config{
				App:   AppConfig{Mode: tt.mode},
				JWT:   JWTConfig{Secret: tt.secret},
				MySQL: MySQLConfig{Password: tt.pass},
			})
		})
	}
}

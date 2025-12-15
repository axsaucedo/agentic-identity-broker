package config

import (
	"testing"
)

func TestLogLevel_Validate(t *testing.T) {
	tests := []struct {
		name    string
		level   LogLevel
		wantErr bool
	}{
		{"valid debug", LogLevelDebug, false},
		{"valid info", LogLevelInfo, false},
		{"valid warn", LogLevelWarn, false},
		{"valid error", LogLevelError, false},
		{"invalid verbose", LogLevel("verbose"), true},
		{"invalid empty", LogLevel(""), true},
		{"invalid trace", LogLevel("trace"), true},
		{"invalid PANIC", LogLevel("PANIC"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.level.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("LogLevel.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLogFormat_Validate(t *testing.T) {
	tests := []struct {
		name    string
		format  LogFormat
		wantErr bool
	}{
		{"valid text", LogFormatText, false},
		{"valid json", LogFormatJSON, false},
		{"invalid xml", LogFormat("xml"), true},
		{"invalid empty", LogFormat(""), true},
		{"invalid yaml", LogFormat("yaml"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.format.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("LogFormat.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLogLevel_String(t *testing.T) {
	tests := []struct {
		name  string
		level LogLevel
		want  string
	}{
		{"debug", LogLevelDebug, "debug"},
		{"info", LogLevelInfo, "info"},
		{"warn", LogLevelWarn, "warn"},
		{"error", LogLevelError, "error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.level.String(); got != tt.want {
				t.Errorf("LogLevel.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLogFormat_String(t *testing.T) {
	tests := []struct {
		name   string
		format LogFormat
		want   string
	}{
		{"text", LogFormatText, "text"},
		{"json", LogFormatJSON, "json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.format.String(); got != tt.want {
				t.Errorf("LogFormat.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

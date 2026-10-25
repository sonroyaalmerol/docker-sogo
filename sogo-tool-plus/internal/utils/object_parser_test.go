package utils

import "testing"

func TestSubscribeTarget(t *testing.T) {
	tests := []struct {
		name   string
		cUID   string
		want   string
		wantOK bool
	}{
		{name: "plain login", cUID: "jdoe", want: "jdoe", wantOK: true},
		{name: "login with domain", cUID: "jdoe@example.com", want: "jdoe@example.com", wantOK: true},
		{name: "group strips prefix", cUID: "@team", want: "team", wantOK: true},
		{name: "group with domain", cUID: "@team@example.com", want: "team@example.com", wantOK: true},
		{name: "default principal", cUID: "<default>", wantOK: false},
		{name: "all principal", cUID: "ALL", wantOK: false},
		{name: "anonymous principal", cUID: "anonymous", wantOK: false},
		{name: "empty", cUID: "", wantOK: false},
		{name: "bare at sign", cUID: "@", wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := SubscribeTarget(tt.cUID)
			if ok != tt.wantOK {
				t.Fatalf("SubscribeTarget(%q) ok = %v, want %v", tt.cUID, ok, tt.wantOK)
			}
			if ok && got != tt.want {
				t.Errorf("SubscribeTarget(%q) = %q, want %q", tt.cUID, got, tt.want)
			}
		})
	}
}

func TestParsePath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantUser string
		wantPath string
		wantErr  bool
	}{
		{
			name:     "calendar acl object",
			input:    "/jdoe/Calendar/personal",
			wantUser: "jdoe",
			wantPath: "Calendar/personal",
		},
		{
			name:     "contacts acl object",
			input:    "/jdoe/Contacts/personal",
			wantUser: "jdoe",
			wantPath: "Contacts/personal",
		},
		{
			name:     "domain qualified owner",
			input:    "/jdoe@example.com/Calendar/personal",
			wantUser: "jdoe@example.com",
			wantPath: "Calendar/personal",
		},
		{name: "missing leading slash", input: "jdoe/Calendar/personal", wantErr: true},
		{name: "too few segments", input: "/jdoe", wantErr: true},
		{name: "empty", input: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, path, err := ParsePath(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ParsePath(%q) expected error, got %q %q", tt.input, user, path)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParsePath(%q) unexpected error: %v", tt.input, err)
			}
			if user != tt.wantUser || path != tt.wantPath {
				t.Errorf("ParsePath(%q) = %q, %q; want %q, %q", tt.input, user, path, tt.wantUser, tt.wantPath)
			}
		})
	}
}

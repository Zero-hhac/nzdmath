package service

import "testing"

func TestValidateResourceUpload(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		size     int64
		wantExt  string
		wantErr  bool
	}{
		{name: "pdf allowed", fileName: "notes.pdf", size: 1024, wantExt: ".pdf"},
		{name: "case insensitive", fileName: "slides.PDF", size: 1024, wantExt: ".pdf"},
		{name: "zip allowed", fileName: "archive.zip", size: 1024, wantExt: ".zip"},
		{name: "too large", fileName: "large.zip", size: ResourceMaxFileSize + 1, wantErr: true},
		{name: "html rejected", fileName: "page.html", size: 1024, wantErr: true},
		{name: "exe rejected", fileName: "tool.exe", size: 1024, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ext, err := validateResourceUpload(tt.fileName, tt.size)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateResourceUpload() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && ext != tt.wantExt {
				t.Fatalf("validateResourceUpload() ext = %q, want %q", ext, tt.wantExt)
			}
		})
	}
}

package service

import (
	"bytes"
	"strings"
	"testing"
)

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
		{name: "py rejected", fileName: "script.py", size: 1024, wantErr: true},
		{name: "js rejected", fileName: "script.js", size: 1024, wantErr: true},
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

func TestValidateUploadContent(t *testing.T) {
	pngBytes := []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A, 0, 0, 0, 0x0D, 'I', 'H', 'D', 'R'}
	pdfBytes := []byte("%PDF-1.4\n%âãÏÓ\n")
	htmlBytes := []byte("<!DOCTYPE html><html><body>hello</body></html>")
	emptyBytes := []byte{}

	tests := []struct {
		name    string
		ext     string
		content []byte
		wantErr bool
	}{
		{name: "png content with jpg ext tolerated", ext: ".jpg", content: pngBytes, wantErr: false},
		{name: "real png ok", ext: ".png", content: pngBytes, wantErr: false},
		{name: "pdf ok", ext: ".pdf", content: pdfBytes, wantErr: false},
		{name: "html disguised as pdf", ext: ".pdf", content: htmlBytes, wantErr: true},
		{name: "html disguised as txt", ext: ".txt", content: htmlBytes, wantErr: true},
		{name: "empty file as png", ext: ".png", content: emptyBytes, wantErr: true},
		{name: "empty file as txt", ext: ".txt", content: emptyBytes, wantErr: false},
		{name: "plain text as txt", ext: ".txt", content: []byte("hello world"), wantErr: false},
		{name: "zip magic as docx", ext: ".docx", content: []byte("PK\x03\x04hello"), wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateUploadContent(tt.ext, bytes.NewReader(tt.content))
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateUploadContent() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateUploadContentRejectsHTMLBytes(t *testing.T) {
	err := validateUploadContent(".png", strings.NewReader("<svg onload=alert(1)>"))
	if err == nil {
		t.Fatal("expected svg/html content to be rejected")
	}
}

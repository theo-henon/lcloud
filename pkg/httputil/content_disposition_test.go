package httputil

import "testing"

func TestContentDispositionAttachment(t *testing.T) {
	header := ContentDispositionAttachment("photo\"bad\r\nname.jpg")
	if header != `attachment; filename="photobadname.jpg"` {
		t.Fatalf("unexpected header: %q", header)
	}
}

func TestContentDispositionAttachmentFallback(t *testing.T) {
	header := ContentDispositionAttachment(`"\`)
	if header != `attachment; filename="download"` {
		t.Fatalf("unexpected header: %q", header)
	}
}

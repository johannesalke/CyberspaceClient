package client

import "testing"

func TestAudioAttachmentsFiltersAndDecodes(t *testing.T) {
	post := Post{
		Attachments: []any{
			map[string]any{"type": "audio", "src": "https://cdn/1.mp3", "title": "Wavelength", "artist": "Hollow", "genre": "ambient"},
			map[string]any{"type": "image", "src": "https://cdn/pic.jpg"},
		},
	}

	audio, err := post.AudioAttachments()
	if err != nil {
		t.Fatalf("AudioAttachments returned error: %v", err)
	}
	if len(audio) != 1 {
		t.Fatalf("got %d audio attachments, want 1", len(audio))
	}
	if got, want := audio[0].Src, "https://cdn/1.mp3"; got != want {
		t.Fatalf("Src = %q, want %q", got, want)
	}
	if got, want := audio[0].Artist, "Hollow"; got != want {
		t.Fatalf("Artist = %q, want %q", got, want)
	}
}

func TestAudioAttachmentsNilAndUnrelated(t *testing.T) {
	if audio, err := (Post{}).AudioAttachments(); err != nil || len(audio) != 0 {
		t.Fatalf("nil attachments: got %v, %v; want empty, nil", audio, err)
	}
	post := Post{Attachments: []any{map[string]any{"type": "image", "src": "x"}}}
	if audio, err := post.AudioAttachments(); err != nil || len(audio) != 0 {
		t.Fatalf("image only: got %v, %v; want empty, nil", audio, err)
	}
}

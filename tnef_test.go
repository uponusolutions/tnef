package tnef

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAttachments(t *testing.T) {
	tests := []struct {
		in              string
		wantAttachments []string
		wantErr         string
	}{
		{"attachments", []string{
			"ZAPPA_~2.JPG",
			"bookmark.htm",
		}, ""},
		// will panic!
		//{"panic", []string{
		//	"ZAPPA_~2.JPG",
		//	"bookmark.htm",
		//}},
		//{"MAPI_ATTACH_DATA_OBJ", []string{
		//	"VIA_Nytt_1402.doc",
		//	"VIA_Nytt_1402.pdf",
		//	"VIA_Nytt_14021.htm",
		//	"MAPI_ATTACH_DATA_OBJ-body.rtf",
		//}},
		//{"MAPI_OBJECT", []string{
		//	"Untitled_Attachment",
		//	"MAPI_OBJECT-body.rtf",
		//}},
		//{"body", []string{
		//	"body-body.html",
		//}},
		//{"data-before-name", []string{
		//	"AUTOEXEC.BAT",
		//	"CONFIG.SYS",
		//	"boot.ini",
		//	"data-before-name-body.rtf",
		//}},
		// {"garbage-at-end", []string{}, ""}, // panics
		//{"long-filename", []string{
		//	"long-filename-body.rtf",
		//}},
		//{"missing-filenames", []string{
		//	"missing-filenames-body.rtf",
		//}},
		{"multi-name-property", []string{}, ""},
		//{"multi-value-attribute", []string{
		//	"208225__5_seconds__Voice_Mail.mp3",
		//	"multi-value-attribute-body.rtf",
		//}},
		{"one-file", []string{
			"AUTHORS",
		}, ""},
		//{"rtf", []string{
		//	"rtf-body.rtf",
		//}},
		//{"triples", []string{
		//	"triples-body.rtf",
		//}},
		{"two-files", []string{
			"AUTHORS",
			"README",
		}, ""},
		{"unicode-mapi-attr-name", []string{
			"spaconsole2.cfg",
			"image001.png",
			"image002.png",
			"image003.png",
		}, ""},
		{"unicode-mapi-attr", []string{
			"example.dat",
		}, ""},

		// Invalid files.
		{"badchecksum", nil, ErrNoMarker.Error()},
		{"empty-file", nil, ErrNoMarker.Error()},
	}

	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			raw, err := os.ReadFile(filepath.Join("testdata", tc.in+".tnef"))
			require.NoError(t, err)

			out, err := Decode(raw)
			if tc.wantErr != "" {
				assert.ErrorContains(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, len(out.Attachments), len(tc.wantAttachments))

			titles := []string{}
			for _, a := range out.Attachments {
				titles = append(titles, a.Title)
			}
			assert.Equal(t, titles, tc.wantAttachments)
		})
	}
}

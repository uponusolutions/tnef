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
		in          string
		attachments []string
		errContains string
	}{
		{"attachments", []string{
			"ZAPPA_~2.JPG",
			"bookmark.htm",
		}, ""},

		// will no longer panic.
		{"panic", []string{
			"ZAPPA_~2.JPG",
			"bookmark.htm",
		}, "not enough bytes read"},
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
		// no longer panics and ignores invalid data
		{"garbage-at-end", []string{}, ""},
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
		//}, ""},
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
		{"badchecksum", nil, "tnef signature not found"},
		{"empty-file", nil, "tnef signature not found"},

		{"signed", []string{"smime.p7m"}, ""},
	}

	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			raw, err := os.ReadFile(filepath.Join("testdata", tc.in+".tnef"))
			require.NoError(t, err)

			out, err := Decode(raw)
			if tc.errContains != "" {
				assert.ErrorContains(t, err, tc.errContains)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, len(tc.attachments), len(out.Attachments))

			titles := []string{}
			for _, a := range out.Attachments {
				titles = append(titles, a.Title)
			}
			assert.Equal(t, tc.attachments, titles)
		})
	}
}

func TestUmlaute(t *testing.T) {
	tnef, err := DecodeFile("testdata/umlaute.tnef")
	require.NoError(t, err)

	htmlBody, found := AttributeByMAPIName(tnef.MAPIAttributes, MAPIBodyHTML)
	assert.True(t, found)
	assert.Equal(t, 790, len(htmlBody.Data))
	assert.Equal(t, "", tnef.Subject)
}

func TestLongFileName(t *testing.T) {
	tnef, err := DecodeFile("testdata/long-filename.tnef")
	require.NoError(t, err)

	attr, found := AttributeByMAPIName(tnef.Attachments[0].MAPIAttributes, MAPIAttachLongFilename)
	assert.True(t, found)

	name, err := attr.AsString()
	require.NoError(t, err)

	assert.Equal(t, "allproductsmar2000.dat", name)
}

func TestDataBeforeName(t *testing.T) {
	tnef, err := DecodeFile("testdata/data-before-name.tnef")
	require.NoError(t, err)

	assert.Equal(t, []string{"AUTOEXEC.BAT", "CONFIG.SYS", "boot.ini"}, allTitles(tnef))
}

///////////////////////////////////////////////////////////////////////////
// Helpers
///////////////////////////////////////////////////////////////////////////

func allTitles(tnef *Data) (titles []string) {
	for _, a := range tnef.Attachments {
		titles = append(titles, a.Title)
	}

	return titles
}

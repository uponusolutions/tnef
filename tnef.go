// Package tnef extracts the body and attachments from Microsoft TNEF files.
package tnef // import "github.com/teamwork/tnef"

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

const (
	tnefSignature = 0x223e9f78
	lvlMessage    = 0x01
	lvlAttachment = 0x02
)

// These can be used to figure out the type of attribute
// an object is.
const (
	ATTOWNER                   = 0x0000 // Owner
	ATTSENTFOR                 = 0x0001 // Sent For
	ATTDELEGATE                = 0x0002 // Delegate
	ATTDATESTART               = 0x0006 // Date Start
	ATTDATEEND                 = 0x0007 // Date End
	ATTAIDOWNER                = 0x0008 // Owner Appointment ID
	ATTREQUESTRES              = 0x0009 // Response Requested.
	ATTFROM                    = 0x8000 // From
	ATTSUBJECT                 = 0x8004 // Subject
	ATTDATESENT                = 0x8005 // Date Sent
	ATTDATERECD                = 0x8006 // Date Received
	ATTMESSAGESTATUS           = 0x8007 // Message Status
	ATTMESSAGECLASS            = 0x8008 // Message Class
	ATTMESSAGEID               = 0x8009 // Message ID
	ATTPARENTID                = 0x800a // Parent ID
	ATTCONVERSATIONID          = 0x800b // Conversation ID
	ATTBODY                    = 0x800c // Body
	ATTPRIORITY                = 0x800d // Priority
	ATTATTACHDATA              = 0x800f // Attachment Data
	ATTATTACHTITLE             = 0x8010 // Attachment File Name
	ATTATTACHMETAFILE          = 0x8011 // Attachment Meta File
	ATTATTACHCREATEDATE        = 0x8012 // Attachment Creation Date
	ATTATTACHMODIFYDATE        = 0x8013 // Attachment Modification Date
	ATTDATEMODIFY              = 0x8020 // Date Modified
	ATTATTACHTRANSPORTFILENAME = 0x9001 // Attachment Transport Filename
	ATTATTACHRENDDATA          = 0x9002 // Attachment Rendering Data
	ATTMAPIPROPS               = 0x9003 // MAPI Properties
	ATTRECIPTABLE              = 0x9004 // Recipients
	ATTATTACHMENT              = 0x9005 // Attachment
	ATTTNEFVERSION             = 0x9006 // TNEF Version
	ATTOEMCODEPAGE             = 0x9007 // OEM Codepage
	ATTORIGNINALMESSAGECLASS   = 0x9008 // Original Message Class
)

type tnefAttribute struct {
	Level    int
	Name     int
	Type     int
	Data     []byte
	Checksum []byte
	Length   int
}

// Data contains the various data from the extracted TNEF file.
type Data struct {
	Body              []byte
	Attachments       []*Attachment
	MAPIAttributes    []*MAPIAttribute
	MessageClass      string
	Subject           string
	CodePagePrimary   int
	CodePageSecondary int
}

// Attachment contains standard attachments that are embedded
// within the TNEF file, with the name and data of the file extracted.
type Attachment struct {
	Title          string
	LongFileName   string
	Data           []byte
	MIMEType       string
	ContentID      string
	MAPIAttributes []*MAPIAttribute
}

func (a *Attachment) addAttr(attr tnefAttribute) error {
	switch attr.Name {
	case ATTDATEMODIFY:
	case ATTATTACHMENT:
		var err error
		a.MAPIAttributes, err = decodeMapi(attr.Data)
		if err != nil {
			return err
		}

		for _, att := range a.MAPIAttributes {
			switch att.Name {
			case MAPIAttachLongFilename:
				a.LongFileName, _ = ToUTF8String(att.Type, att.Data)
			case MAPIAttachMimeTag:
				a.MIMEType, _ = ToUTF8String(att.Type, att.Data)
			case MAPIAttachDataObj:
				// TODO: Replace data?
			case MAPIAttachContentId:
				a.ContentID, _ = ToUTF8String(att.Type, att.Data)
			}
		}
	case ATTATTACHTITLE:
		a.Title = strings.Replace(string(attr.Data), "\x00", "", -1)
	case ATTATTACHDATA:
		a.Data = attr.Data
	}

	return nil
}

// DecodeFile is a utility function that reads the file into memory
// before calling the normal Decode function on the data.
func DecodeFile(path string) (*Data, error) {
	data, err := os.ReadFile(path) //nolint: gosec
	if err != nil {
		return nil, err
	}

	return Decode(data)
}

// Decode will accept a stream of bytes in the TNEF format and extract the
// attachments and body into a Data object.
func Decode(data []byte) (*Data, error) {
	e := func(err error) (*Data, error) {
		return nil, fmt.Errorf("tnef decode error: %w", err)
	}

	r := newTNEFReader(data)

	sigValid := false
	err := r.rb(4, func(v []byte) {
		sigValid = byteToInt(v) == tnefSignature
	})
	if !sigValid {
		return e(fmt.Errorf("tnef signature not found: %w", err))
	}

	// Skip key.
	// key := binary.LittleEndian.Uint32(data[4:6])
	if err := r.skip(2); err != nil {
		return e(err)
	}

	var attachment *Attachment
	tnef := &Data{
		Attachments: []*Attachment{},
	}

	for {
		attr, err := readTNEFAttribute(r)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return e(err)
		}

		if attr.Name == ATTATTACHRENDDATA {
			attachment = new(Attachment)
			tnef.Attachments = append(tnef.Attachments, attachment)
		}

		switch attr.Level {
		case lvlMessage:
			switch attr.Name {
			case ATTBODY:
				tnef.Body = attr.Data
			case ATTMAPIPROPS:
				tnef.MAPIAttributes, err = decodeMapi(attr.Data)
				if err != nil {
					return e(err)
				}
			case ATTMESSAGECLASS:
				tnef.MessageClass, _ = ToUTF8String(szmapiString, attr.Data)
			case ATTSUBJECT:
				tnef.Subject, _ = ToUTF8String(szmapiUnicodeString, attr.Data)
			case ATTOEMCODEPAGE:
				if attr.Length < 6 {
					continue
				}
				tnef.CodePagePrimary = byteToInt(attr.Data[0:2])
				tnef.CodePageSecondary = byteToInt(attr.Data[2:])
			}
		case lvlAttachment:
			if attachment == nil {
				return e(errors.New("attachment level reached, but attachment is nil"))
			}

			if err := attachment.addAttr(attr); err != nil {
				return e(err)
			}
		default:
			return e(fmt.Errorf("invalid level type attribute: %d", attr.Level))
		}
	}

	return tnef, nil
}

func readTNEFAttribute(r *tnefBytesReader) (attr tnefAttribute, err error) {
	e := func(err error) (tnefAttribute, error) {
		return attr, fmt.Errorf("tnef attribute decode error: %w", err)
	}

	err = r.rb(1, func(v []byte) { attr.Level = byteToInt(v) })
	if err != nil {
		return e(fmt.Errorf("can't read level: %w", err))
	}

	err = r.rb(2, func(v []byte) { attr.Name = byteToInt(v) })
	if err != nil {
		return e(fmt.Errorf("can't read name: %w", err))
	}

	err = r.rb(2, func(v []byte) { attr.Type = byteToInt(v) })
	if err != nil {
		return e(fmt.Errorf("can't read type: %w", err))
	}

	err = r.rb(4, func(v []byte) { attr.Length = byteToInt(v) })
	if err != nil {
		return e(fmt.Errorf("can't read length: %w", err))
	}

	if attr.Length > 0 {
		err = r.rb(attr.Length, func(v []byte) { attr.Data = v })
		if err != nil {
			return e(fmt.Errorf("can't read data: %w", err))
		}
	}

	err = r.rb(2, func(v []byte) { attr.Checksum = v })
	if err != nil {
		return e(err)
	}

	attr.Length = r.read

	return attr, nil
}

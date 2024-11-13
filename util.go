package tnef

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
	"unicode/utf16"
)

type tnefBytesReader struct {
	in   *bytes.Buffer
	read int
}

func newTNEFReader(raw []byte) *tnefBytesReader {
	return &tnefBytesReader{
		in: bytes.NewBuffer(raw),
	}
}

func (r *tnefBytesReader) readBytes(length int) ([]byte, error) {
	if length == 0 {
		return nil, errors.New("can not read 0 bytes")
	}

	b := make([]byte, length)
	read, err := r.in.Read(b)
	if err != nil {
		return nil, err
	}

	r.read += read

	if length != read {
		return nil, errors.New("not enough bytes read")
	}

	return b, nil
}

func (r tnefBytesReader) skip(length int) error {
	if length == 0 {
		return nil
	}

	_, err := r.readBytes(length)
	return err
}

func (r *tnefBytesReader) rb(length int, f func(v []byte)) error {
	b, err := r.readBytes(length)
	if err != nil {
		return err
	}

	f(b)

	return nil
}

func (r *tnefBytesReader) readInt16() (num int, err error) {
	err = r.rb(2, func(v []byte) { num = byteToInt(v) })
	return num, err
}

func (r *tnefBytesReader) readInt32() (num int, err error) {
	err = r.rb(4, func(v []byte) { num = byteToInt(v) })
	return num, err
}

func (r *tnefBytesReader) readString(length int) (s string, err error) {
	err = r.rb(length, func(v []byte) {
		// TODO: Other encodings?
		s, _ = ToUTF8String(szmapiUnicodeString, v)
		if err != nil {
			s, _ = ToUTF8String(szmapiString, v)
		}
	})

	return s, err
}

func byteToInt(data []byte) int {
	var num int
	var n uint
	for _, b := range data {
		num += (int(b) << n)
		n += 8
	}
	return num
}

// ToUTF8String converts data of type t to an UTF-8 string.
func ToUTF8String(t int, data []byte) (string, error) {
	switch t {
	case szmapiUnicodeString:
		if len(data)%2 != 0 ||
			len(data) < 2 ||
			!bytes.Equal(data[len(data)-2:], []byte{0, 0}) {
			return "", errors.New("invalid format for type 31, not a straight number of bytes and not two null bytes at the end")
		}

		arr := bytesToUint16(data[0 : len(data)-2])

		runes := utf16.Decode(arr)
		return string(runes), nil
	case szmapiString:
		return strings.Replace(string(data), "\x00", "", -1), nil
	default:
		return "", errors.New("no string mapping found")
	}
}

func bytesToUint16(a []byte) []uint16 {
	arr := make([]uint16, len(a)/2)
	for i := 0; i < len(arr); i++ {
		arr[i] = binary.LittleEndian.Uint16(a[i*2 : i*2+2])
	}
	return arr
}

// DebugAttachments prints attachments to stdout.
func DebugAttachments(attachments []*Attachment) {
	for _, a := range attachments {
		DebugAttachment(a)
	}
}

// DebugAttachment prints attachment to stdout.
func DebugAttachment(a *Attachment) {
	if a == nil {
		return
	}

	fmt.Printf("Title %s, LongFileName: %s, MIMEType %s, ContentID %s, DataLen %d\n",
		a.Title, a.LongFileName, a.MIMEType, a.ContentID, len(a.Data))
	DebugAttributes(a.MAPIAttributes)
}

// DebugAttributes prints attributes to stdout.
func DebugAttributes(attrs []*MAPIAttribute) {
	for _, a := range attrs {
		fmt.Printf("Name %x, Type %x, Names %+v, Guid %x, Data %s\n",
			a.Name, a.Type, a.Names, a.GUID, a.Data)
	}
}

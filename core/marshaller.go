package core

import (
	"bytes"
	"strconv"
	"unicode/utf8"

	"github.com/glerchundi/go-systemd/sdjournal"
)

const hex = "0123456789abcdef"

type Buffer struct {
	bytes.Buffer
	scratch [64]byte
}

func (b *Buffer) WriteJsonString(s string) (int, error) {
	len0 := b.Len()
	b.WriteByte('"')
	start := 0
	for i := 0; i < len(s); {
		if r := s[i]; r < utf8.RuneSelf {
			if 0x20 <= r && r != '\\' && r != '"' && r != '<' && r != '>' && r != '&' {
				i++
				continue
			}
			if start < i {
				b.WriteString(s[start:i])
			}
			switch r {
			case '\\', '"':
				b.WriteByte('\\')
				b.WriteByte(r)
			case '\n':
				b.WriteByte('\\')
				b.WriteByte('n')
			case '\r':
				b.WriteByte('\\')
				b.WriteByte('r')
			case '\t':
				b.WriteByte('\\')
				b.WriteByte('t')
			default:
				// This encodes bytes < 0x20 except for \n and \r,
				// as well as <, > and &. The latter are escaped because they
				// can lead to security holes when user-controlled strings
				// are rendered into JSON and served to some browsers.
				b.WriteString(`\u00`)
				b.WriteByte(hex[r >>4])
				b.WriteByte(hex[r &0xF])
			}
			i++
			start = i
			continue
		}
		c, size := utf8.DecodeRuneInString(s[i:])
		if c == utf8.RuneError && size == 1 {
			if start < i {
				b.WriteString(s[start:i])
			}
			b.WriteString(`\ufffd`)
			i += size
			start = i
			continue
		}
		// U+2028 is LINE SEPARATOR.
		// U+2029 is PARAGRAPH SEPARATOR.
		// They are both technically valid characters in JSON strings,
		// but don't work in JSONP, which has to be evaluated as JavaScript,
		// and can lead to security holes there. It is valid JSON to
		// escape them, so we do so unconditionally.
		// See http://timelessrepo.com/json-isnt-a-javascript-subset for discussion.
		if c == '\u2028' || c == '\u2029' {
			if start < i {
				b.WriteString(s[start:i])
			}
			b.WriteString(`\u202`)
			b.WriteByte(hex[c&0xF])
			i += size
			start = i
			continue
		}
		i += size
	}
	if start < len(s) {
		b.WriteString(s[start:])
	}
	b.WriteByte('"')
	return b.Len() - len0, nil
}

func (b *Buffer) WriteUint(v uint64) {
	d := strconv.AppendUint(b.scratch[:0], v, 10)
	b.Write(d)
}

func (b *Buffer) Rewind(n int) {
	b.Truncate(b.Len()-n)
}

// Marshaller

type JournalEntryMarshaller struct {
	buf Buffer
}

func (m *JournalEntryMarshaller) MarshalOne(e *sdjournal.JournalEntry) []byte {
	m.buf.Reset()
	m.marshalOne(e)
	return m.buf.Bytes()
}

func (m *JournalEntryMarshaller) MarshalAll(ea []*sdjournal.JournalEntry) []byte {
	m.buf.Reset()
	m.buf.WriteByte('[')
	if ea != nil {
		for _, e := range ea {
			m.marshalOne(e)
			m.buf.WriteByte(',')
		}
		m.buf.Rewind(1)
	}
	m.buf.WriteByte(']')
	return m.buf.Bytes()
}

func (m *JournalEntryMarshaller) Bytes() []byte {
	return m.buf.Bytes()
}

func (m *JournalEntryMarshaller) String() string {
	return m.buf.String()
}

func (m *JournalEntryMarshaller) marshalOne(e *sdjournal.JournalEntry) {
	m.buf.WriteString(`{"__CURSOR":`)
	m.buf.WriteJsonString(e.Cursor)
	m.buf.WriteString(`,"__REALTIME_TIMESTAMP":`)
	m.buf.WriteUint(uint64(e.RealtimeTimestamp))
	m.buf.WriteString(`,"__MONOTONIC_TIMESTAMP":`)
	m.buf.WriteUint(uint64(e.MonotonicTimestamp))
	if e.Fields != nil {
		m.buf.WriteByte(',')
		for key, value := range e.Fields {
			m.buf.WriteJsonString(key)
			m.buf.WriteString(`:`)
			m.buf.WriteJsonString(string(value))
			m.buf.WriteByte(',')
		}
		m.buf.Rewind(1)
	}
	m.buf.WriteByte('}')
}
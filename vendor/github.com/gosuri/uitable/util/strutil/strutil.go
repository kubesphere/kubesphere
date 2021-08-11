// Package strutil provides various utilities for manipulating strings
package strutil

import (
	"bytes"
	"github.com/mattn/go-runewidth"
)

// PadRight returns a new string of a specified length in which the end of the current string is padded with spaces or with a specified Unicode character.
func PadRight(str string, length int, pad byte) string {
	slen := runewidth.StringWidth(str)
	if slen >= length {
		return str
	}
	buf := bytes.NewBufferString(str)
	for i := 0; i < length-slen; i++ {
		buf.WriteByte(pad)
	}
	return buf.String()
}

// PadLeft returns a new string of a specified length in which the beginning of the current string is padded with spaces or with a specified Unicode character.
func PadLeft(str string, length int, pad byte) string {
	slen := runewidth.StringWidth(str)
	if slen >= length {
		return str
	}
	var buf bytes.Buffer
	for i := 0; i < length-slen; i++ {
		buf.WriteByte(pad)
	}
	buf.WriteString(str)
	return buf.String()
}

// Resize resizes the string with the given length. It ellipses with '...' when the string's length exceeds
// the desired length or pads spaces to the right of the string when length is smaller than desired
func Resize(s string, length uint, rightAlign bool) string {
	slen := runewidth.StringWidth(s)
	n := int(length)
	if slen == n {
		return s
	}
	// Pads only when length of the string smaller than len needed
	if rightAlign {
		s = PadLeft(s, n, ' ')
	} else {
		s = PadRight(s, n, ' ')
	}
	if slen > n {
		rs := []rune(s)
		var buf bytes.Buffer
		w := 0
		for _, r := range rs {
			buf.WriteRune(r)
			rw := runewidth.RuneWidth(r)
			if w+rw >= n-3 {
				break
			}
			w += rw
		}
		buf.WriteString("...")
		s = buf.String()
	}
	return s
}

// Join joins the list of the string with the delim provided
func Join(list []string, delim string) string {
	var buf bytes.Buffer
	for i := 0; i < len(list)-1; i++ {
		buf.WriteString(list[i] + delim)
	}
	buf.WriteString(list[len(list)-1])
	return buf.String()
}

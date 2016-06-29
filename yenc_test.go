package yenc

import (
	"reflect"
	"strings"
	"testing"
)

func TestDecode(t *testing.T) {
	f := strings.NewReader("=ybegin part=1 line=128 size=123 name=yhello\r\n\x92\x8f\x96\x96\x994\xa1\x99\x9c\x96\x8e4\r\n=yend size=123 part=1 pcrc32=1a2a3a4a")

	want := Part{
		BeginPart: 1,
		BeginSize: 123,
		Name:      "yhello",
		EndPart:   1,
		EndSize:   123,
		CRC32:     "1a2a3a4a",
		Body:      []byte(string("hello\nworld\n")),
	}

	p, err := Decode(f)

	if err != nil {
		t.Errorf("Got error %+v", err)
	}

	if !reflect.DeepEqual(p, &want) {
		t.Errorf("Got %+v, want %+v", p, want)
	}
}

func TestDecodeMinimalBegin(t *testing.T) {
	f := strings.NewReader(`=ybegin line=128 size=1234 name=foo bar
hello
=yend size=2345`)

	want := Part{
		BeginSize: 1234,
		Name:      "foo bar",
		EndSize:   2345,
		Body:      []byte{62, 59, 66, 66, 69},
	}

	p, err := Decode(f)

	if err != nil {
		t.Error("Got err", err)
	}

	if !reflect.DeepEqual(p, &want) {
		t.Errorf("Got %+v, want %+v", p, want)
	}
}

func TestDecodeMultiPart(t *testing.T) {
	f := strings.NewReader(`=ybegin part=1 line=128 size=1234 name=foo bar
=ypart begin=55 end=60
hello
=yend size=2345 part=2 pcrc32=1a2a3a4a`)

	want := Part{
		BeginPart: 1,
		BeginSize: 1234,
		PartBegin: 55,
		PartEnd:   60,
		Name:      "foo bar",
		EndPart:   2,
		EndSize:   2345,
		CRC32:     "1a2a3a4a",
		Body:      []byte{62, 59, 66, 66, 69},
	}

	p, err := Decode(f)

	if err != nil {
		t.Errorf("Got error %+v", err)
	}

	if !reflect.DeepEqual(p, &want) {
		t.Errorf("Got %+v, want %+v", p, want)
	}
}

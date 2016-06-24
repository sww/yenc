package yenc

import (
	"bufio"
	"io"
	"strconv"
	"strings"
)

func parseLine(line string) map[string]string {
	mapping := make(map[string]string)
	split := strings.Split(line, " ")[1:]
	for _, s := range split {
		kv := strings.Split(s, "=")
		if len(kv) < 2 {
			continue
		}
		mapping[kv[0]] = kv[1]
	}

	return mapping
}

type Part struct {
	// Values from begin.
	BeginPart int
	BeginSize int
	Name      string
	Total     int

	// Values from part, if available.
	PartBegin int
	PartEnd   int

	// Values from end.
	EndPart int
	EndSize int
	CRC32   string

	Body []byte
}

const nameHeader = "name="
const nameLength = len(nameHeader)

func (p *Part) parseBegin(line string) {
	nameIndex := strings.Index(line, nameHeader)
	if nameIndex < 0 {
		panic("Could not find name= parameter!")
	}

	p.Name = line[nameIndex+nameLength:]

	for k, v := range parseLine(line[:nameIndex]) {
		switch k {
		case "part":
			part, err := strconv.Atoi(v)
			if err != nil {
				continue
			}
			p.BeginPart = part
		case "size":
			size, err := strconv.Atoi(v)
			if err != nil {
				continue
			}
			p.BeginSize = size
		}
	}
}

func (p *Part) parsePart(line string) {
	for k, v := range parseLine(line) {
		switch k {
		case "begin":
			begin, err := strconv.Atoi(v)
			if err != nil {
				continue
			}
			p.PartBegin = begin
		case "end":
			end, err := strconv.Atoi(v)
			if err != nil {
				continue
			}
			p.PartEnd = end
		}
	}
}

func (p *Part) parseEnd(line string) {
	for k, v := range parseLine(line) {
		switch k {
		case "part":
			part, err := strconv.Atoi(v)
			if err != nil {
				continue
			}
			p.EndPart = part
		case "size":
			size, err := strconv.Atoi(v)
			if err != nil {
				continue
			}
			p.EndSize = size
		case "crc32", "pcrc32":
			p.CRC32 = string(v)
		}
	}
}

func Decode(f io.Reader) (*Part, error) {
	p := Part{}

	escapeChar := false
	newLine := true

	r := bufio.NewReader(f)
	for {
		c, err := r.ReadByte()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		s := string(c)

		if newLine && s == "=" {
			peek, err := r.Peek(8)
			if err != nil {
				continue
			}

			peekString := s + string(peek)

			switch {
			case strings.HasPrefix(peekString, "=ybegin "), strings.HasPrefix(peekString, "=ypart "), strings.HasPrefix(peekString, "=yend "):
				line, _, err := r.ReadLine()
				if err != nil {
					continue
				}
				l := string(line)
				switch {
				case strings.HasPrefix(l, "ybegin "):
					p.parseBegin(l)
					continue
				case strings.HasPrefix(l, "ypart "):
					p.parsePart(l)
					continue
				case strings.HasPrefix(l, "yend "):
					p.parseEnd(l)
					continue
				}
			}

		}

		newLine = false

		switch {
		case strings.EqualFold(s, "="):
			escapeChar = true
		case escapeChar:
			p.Body = append(p.Body, c-0x6a)
			escapeChar = false
		case s == "\n" || s == "\r":
			newLine = true
		default:
			p.Body = append(p.Body, c-0x2a)
		}

	}

	return &p, nil
}

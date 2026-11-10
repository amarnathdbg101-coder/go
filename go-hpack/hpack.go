package hpack

// Version defines the release version of go-hpack.
const Version = "1.0.0"

// -------------------------------------------------------------
// Integer Representation (RFC 7541 Section 5.1)
// -------------------------------------------------------------

func appendVarInt(dst []byte, n int, i uint64) []byte {
	maxPrefix := uint64((1 << n) - 1)
	if i < maxPrefix {
		// Fits within prefix bits
		dst[len(dst)-1] |= byte(i)
		return dst
	}

	dst[len(dst)-1] |= byte(maxPrefix)
	i -= maxPrefix

	for i >= 128 {
		dst = append(dst, byte(i%128)|0x80)
		i /= 128
	}
	dst = append(dst, byte(i))
	return dst
}

func readVarInt(n int, b []byte) (uint64, []byte, error) {
	if len(b) == 0 {
		return 0, nil, ErrUnexpectedEOF
	}

	maxPrefix := uint64((1 << n) - 1)
	val := uint64(b[0]) & maxPrefix
	b = b[1:]

	if val < maxPrefix {
		return val, b, nil
	}

	var m uint
	for {
		if len(b) == 0 {
			return 0, nil, ErrUnexpectedEOF
		}
		cur := b[0]
		b = b[1:]

		val += uint64(cur&0x7F) << m
		m += 7
		if m > 63 {
			return 0, nil, ErrIntegerOverflow
		}
		if (cur & 0x80) == 0 {
			break
		}
	}
	return val, b, nil
}

// -------------------------------------------------------------
// Encoder
// -------------------------------------------------------------

// Encoder serializes HTTP/2 header lists into compressed HPACK byte streams.
type Encoder struct {
	dynTable *dynamicTable
}

// NewEncoder creates an HPACK encoder with the specified dynamic table maximum capacity.
func NewEncoder(maxTableSize int) *Encoder {
	if maxTableSize < 0 {
		maxTableSize = 4096
	}
	return &Encoder{
		dynTable: newDynamicTable(maxTableSize),
	}
}

// Encode compresses a slice of headers into HPACK format.
func (e *Encoder) Encode(headers []HeaderField) []byte {
	var buf []byte

	for _, h := range headers {
		// 1. Check for exact match in static or dynamic table
		if idx := e.findExactMatch(h); idx > 0 {
			// Indexed Header Field Representation (1xxxxxxx)
			buf = append(buf, 0x80)
			buf = appendVarInt(buf, 7, uint64(idx))
			continue
		}

		// 2. Check for name match
		nameIdx := e.findNameMatch(h.Name)
		if h.Sensitive {
			// Literal Never Indexed (0001xxxx)
			buf = append(buf, 0x10)
			if nameIdx > 0 {
				buf = appendVarInt(buf, 4, uint64(nameIdx))
			} else {
				buf = appendVarInt(buf, 4, 0)
				buf = appendString(buf, h.Name)
			}
			buf = appendString(buf, h.Value)
			continue
		}

		// Literal with Incremental Indexing (01xxxxxx)
		buf = append(buf, 0x40)
		if nameIdx > 0 {
			buf = appendVarInt(buf, 6, uint64(nameIdx))
		} else {
			buf = appendVarInt(buf, 6, 0)
			buf = appendString(buf, h.Name)
		}
		buf = appendString(buf, h.Value)

		// Add to dynamic table
		e.dynTable.add(h)
	}

	return buf
}

func (e *Encoder) findExactMatch(h HeaderField) int {
	// Search static table
	for i := 1; i <= 61; i++ {
		if staticTable[i].Name == h.Name && staticTable[i].Value == h.Value {
			return i
		}
	}
	// Search dynamic table
	for i, entry := range e.dynTable.entries {
		if entry.Name == h.Name && entry.Value == h.Value {
			return 61 + 1 + i
		}
	}
	return 0
}

func (e *Encoder) findNameMatch(name string) int {
	for i := 1; i <= 61; i++ {
		if staticTable[i].Name == name {
			return i
		}
	}
	for i, entry := range e.dynTable.entries {
		if entry.Name == name {
			return 61 + 1 + i
		}
	}
	return 0
}

func appendString(dst []byte, s string) []byte {
	// Without Huffman: H bit is 0, 7-bit prefix length
	dst = append(dst, 0x00)
	dst = appendVarInt(dst, 7, uint64(len(s)))
	return append(dst, []byte(s)...)
}

// -------------------------------------------------------------
// Decoder
// -------------------------------------------------------------

// Decoder decompresses an HPACK byte stream back into HTTP/2 header lists.
type Decoder struct {
	dynTable *dynamicTable
}

// NewDecoder creates an HPACK decoder with the specified dynamic table maximum capacity.
func NewDecoder(maxTableSize int) *Decoder {
	if maxTableSize < 0 {
		maxTableSize = 4096
	}
	return &Decoder{
		dynTable: newDynamicTable(maxTableSize),
	}
}

// Decode parses compressed HPACK bytes into header fields.
func (d *Decoder) Decode(b []byte) ([]HeaderField, error) {
	var headers []HeaderField

	for len(b) > 0 {
		b0 := b[0]

		if (b0 & 0x80) != 0 {
			// Indexed Header Field (1xxxxxxx, 7-bit prefix)
			idx, rest, err := readVarInt(7, b)
			if err != nil {
				return nil, err
			}
			b = rest

			hf, err := d.lookupIndex(int(idx))
			if err != nil {
				return nil, err
			}
			headers = append(headers, hf)

		} else if (b0 & 0x40) != 0 {
			// Literal with Incremental Indexing (01xxxxxx, 6-bit prefix)
			nameIdx, rest, err := readVarInt(6, b)
			if err != nil {
				return nil, err
			}
			b = rest

			var name, val string
			if nameIdx > 0 {
				hf, err := d.lookupIndex(int(nameIdx))
				if err != nil {
					return nil, err
				}
				name = hf.Name
			} else {
				str, restBytes, err := readString(b)
				if err != nil {
					return nil, err
				}
				name = str
				b = restBytes
			}

			vStr, restBytes, err := readString(b)
			if err != nil {
				return nil, err
			}
			val = vStr
			b = restBytes

			hf := HeaderField{Name: name, Value: val}
			d.dynTable.add(hf)
			headers = append(headers, hf)

		} else if (b0 & 0x20) != 0 {
			// Dynamic Table Size Update (001xxxxx, 5-bit prefix)
			newSize, rest, err := readVarInt(5, b)
			if err != nil {
				return nil, err
			}
			b = rest
			d.dynTable.setMaxSize(int(newSize))

		} else {
			// Literal without Indexing (0000xxxx) or Never Indexed (0001xxxx)
			isNeverIndexed := (b0 & 0x10) != 0
			nameIdx, rest, err := readVarInt(4, b)
			if err != nil {
				return nil, err
			}
			b = rest

			var name, val string
			if nameIdx > 0 {
				hf, err := d.lookupIndex(int(nameIdx))
				if err != nil {
					return nil, err
				}
				name = hf.Name
			} else {
				str, restBytes, err := readString(b)
				if err != nil {
					return nil, err
				}
				name = str
				b = restBytes
			}

			vStr, restBytes, err := readString(b)
			if err != nil {
				return nil, err
			}
			val = vStr
			b = restBytes

			headers = append(headers, HeaderField{
				Name:      name,
				Value:     val,
				Sensitive: isNeverIndexed,
			})
		}
	}

	return headers, nil
}

func (d *Decoder) lookupIndex(idx int) (HeaderField, error) {
	if idx <= 0 {
		return HeaderField{}, ErrInvalidIndex
	}
	if idx <= 61 {
		return staticTable[idx], nil
	}

	dynIdx := idx - 61
	hf, exists := d.dynTable.get(dynIdx)
	if !exists {
		return HeaderField{}, ErrInvalidIndex
	}
	return hf, nil
}

func readString(b []byte) (string, []byte, error) {
	if len(b) == 0 {
		return "", nil, ErrUnexpectedEOF
	}

	// 7-bit prefix length (H bit ignored for plain literals)
	strLen, rest, err := readVarInt(7, b)
	if err != nil {
		return "", nil, err
	}

	if uint64(len(rest)) < strLen {
		return "", nil, ErrUnexpectedEOF
	}

	val := string(rest[:strLen])
	return val, rest[strLen:], nil
}

// HeaderListsEqual compares two header slices for exact equality.
func HeaderListsEqual(a, b []HeaderField) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Name != b[i].Name || a[i].Value != b[i].Value {
			return false
		}
	}
	return true
}

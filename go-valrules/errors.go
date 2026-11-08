package valrules

import "errors"

var (
	// ErrInvalidLuhn is returned when a card/IMEI number fails the Mod-10 Luhn checksum.
	ErrInvalidLuhn = errors.New("valrules: invalid luhn checksum")
	// ErrInvalidGSTIN is returned when a GSTIN does not match legal Indian GST specifications.
	ErrInvalidGSTIN = errors.New("valrules: invalid GSTIN format or checksum")
	// ErrInvalidPAN is returned when a PAN card number violates IT department format.
	ErrInvalidPAN = errors.New("valrules: invalid PAN format")
	// ErrInvalidIFSC is returned when an IFSC code violates RBI branch format.
	ErrInvalidIFSC = errors.New("valrules: invalid IFSC code format")
	// ErrInvalidPhone is returned when a phone number does not conform to E.164.
	ErrInvalidPhone = errors.New("valrules: invalid E.164 phone number")
	// ErrInvalidUUID is returned when a UUID string format is invalid.
	ErrInvalidUUID = errors.New("valrules: invalid UUID format")
	// ErrInvalidURL is returned when a URL is malformed or lacks scheme/host.
	ErrInvalidURL = errors.New("valrules: invalid URL format")
)

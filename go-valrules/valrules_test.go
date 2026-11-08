package valrules

import (
	"testing"
)

func TestValidateLuhn(t *testing.T) {
	// Standard valid test card numbers (Luhn compliant)
	validCards := []string{
		"49927398716",
		"49927398716",
		"79927398713",
		"4242 4242 4242 4242",
		"4242-4242-4242-4242",
	}

	for _, card := range validCards {
		if !ValidateLuhn(card) {
			t.Errorf("expected card %s to pass Luhn check", card)
		}
	}

	invalidCards := []string{
		"49927398717",
		"4242424242424241",
		"abcd-efgh-ijkl-mnop",
		"123",
	}

	for _, card := range invalidCards {
		if ValidateLuhn(card) {
			t.Errorf("expected card %s to fail Luhn check", card)
		}
	}
}

func TestValidatePAN(t *testing.T) {
	validPANs := []string{
		"ABCDE1234F",
		"BKZPR8372K",
		"AAACT1234A",
	}

	for _, pan := range validPANs {
		if !ValidatePAN(pan) {
			t.Errorf("expected PAN %s to be valid", pan)
		}
	}

	invalidPANs := []string{
		"ABCDE12345",  // ends in digit
		"ABCD12345F",  // 4 letters instead of 5
		"ABCDE123F",   // too short
		"ABCDE12345FF",// too long
	}

	for _, pan := range invalidPANs {
		if ValidatePAN(pan) {
			t.Errorf("expected PAN %s to be invalid", pan)
		}
	}
}

func TestValidateGSTIN(t *testing.T) {
	// 27 (Maharashtra) + AAAAA0000A + 1 + Z + 5
	validGSTIN := "27ABCDE1234F1Z5"
	details, err := ValidateGSTIN(validGSTIN)
	if err != nil {
		t.Fatalf("expected valid GSTIN %s to pass, got error: %v", validGSTIN, err)
	}

	if details.StateCode != "27" || details.StateName != "Maharashtra" {
		t.Fatalf("unexpected state details: %+v", details)
	}
	if details.PAN != "ABCDE1234F" {
		t.Fatalf("unexpected PAN: %s", details.PAN)
	}

	// Invalid state code "99"
	invalidGSTIN := "99ABCDE1234F1Z5"
	if _, err := ValidateGSTIN(invalidGSTIN); err == nil {
		t.Fatalf("expected invalid state code to fail")
	}

	// Invalid PAN portion
	invalidPAN := "2712345678901Z5"
	if _, err := ValidateGSTIN(invalidPAN); err == nil {
		t.Fatalf("expected invalid PAN inside GSTIN to fail")
	}
}

func TestValidateIFSC(t *testing.T) {
	validIFSC := []string{
		"SBIN0001234",
		"HDFC0000001",
		"ICIC0001928",
	}

	for _, ifsc := range validIFSC {
		if !ValidateIFSC(ifsc) {
			t.Errorf("expected IFSC %s to be valid", ifsc)
		}
	}

	invalidIFSC := []string{
		"SBIN1001234", // 5th char must be '0'
		"SBI0001234",  // only 3 letters
		"SBIN00012345",// too long
	}

	for _, ifsc := range invalidIFSC {
		if ValidateIFSC(ifsc) {
			t.Errorf("expected IFSC %s to be invalid", ifsc)
		}
	}
}

func TestValidatePhones(t *testing.T) {
	// E.164
	if !ValidateE164("+919876543210") || !ValidateE164("+14155552671") {
		t.Errorf("expected valid E.164 phone numbers")
	}
	if ValidateE164("9876543210") || ValidateE164("+0123456") {
		t.Errorf("expected invalid E.164 numbers")
	}

	// Indian Mobile
	if !ValidateIndianMobile("9876543210") || !ValidateIndianMobile("+91 8876543210") || !ValidateIndianMobile("07876543210") {
		t.Errorf("expected valid Indian mobile numbers")
	}
	if ValidateIndianMobile("5876543210") || ValidateIndianMobile("98765") {
		t.Errorf("expected invalid Indian mobile numbers")
	}
}

func TestValidateNetworkAndIdentifiers(t *testing.T) {
	// MAC
	if !ValidateMAC("00:1A:2B:3C:4D:5E") || !ValidateMAC("00-1A-2B-3C-4D-5E") {
		t.Errorf("expected valid MAC address")
	}
	if ValidateMAC("00:1A:2B:3C:4D:ZZ") {
		t.Errorf("expected invalid MAC address")
	}

	// UUID
	if !ValidateUUID("f47ac10b-58cc-4372-a567-0e02b2c3d479") {
		t.Errorf("expected valid UUID v4")
	}
	if ValidateUUID("not-a-uuid") {
		t.Errorf("expected invalid UUID")
	}

	// Base64
	if !ValidateBase64("SGVsbG8gV29ybGQ=") {
		t.Errorf("expected valid base64")
	}
	if ValidateBase64("@@@not-base-64@@@") {
		t.Errorf("expected invalid base64")
	}

	// Hex
	if !ValidateHex("deadbeef0123") {
		t.Errorf("expected valid hex")
	}
	if ValidateHex("deadbeefg") { // odd length or invalid character
		t.Errorf("expected invalid hex")
	}

	// URL
	if !ValidateURL("https://example.com/api/v1?query=1") || !ValidateURL("http://localhost:8080/health") {
		t.Errorf("expected valid URL")
	}
	if ValidateURL("ftp-invalid") || ValidateURL("https://example.com:99999") {
		t.Errorf("expected invalid URL")
	}
}

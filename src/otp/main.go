package otp

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/binary"
	"math"
	"strconv"
	"time"
)

func GenerateTOTP(now time.Time, key []byte, interval time.Duration, digits int) string {
	counter := uint64(now.Unix()) / uint64(interval.Seconds())
	return GenerateHOTP(key, counter, digits)
}

func VerifyTOTP(now time.Time, key []byte, interval time.Duration, digits int, otp string) bool {
	if len(otp) != digits {
		return false
	}
	generated := GenerateTOTP(now, key, interval, digits)
	valid := subtle.ConstantTimeCompare([]byte(generated), []byte(otp)) == 1
	return valid
}

func VerifyTOTPWithGracePeriod(now time.Time, key []byte, interval time.Duration, digits int, otp string, gracePeriod time.Duration) bool {
	counter1 := uint64(now.Add(-1*gracePeriod).Unix()) / uint64(interval.Seconds())
	generated := GenerateHOTP(key, counter1, digits)
	valid := subtle.ConstantTimeCompare([]byte(generated), []byte(otp)) == 1
	if valid {
		return true
	}
	counter2 := uint64(now.Unix()) / uint64(interval.Seconds())
	if counter2 != counter1 {
		generated = GenerateHOTP(key, counter2, digits)
		valid = subtle.ConstantTimeCompare([]byte(generated), []byte(otp)) == 1
		if valid {
			return true
		}
	}
	counter3 := uint64(now.Add(gracePeriod).Unix()) / uint64(interval.Seconds())
	if counter3 != counter1 && counter3 != counter2 {
		generated = GenerateHOTP(key, counter3, digits)
		valid = subtle.ConstantTimeCompare([]byte(generated), []byte(otp)) == 1
		if valid {
			return true
		}
	}
	return false
}

func GenerateHOTP(key []byte, counter uint64, digits int) string {
	if digits < 6 || digits > 8 {
		panic("invalid totp digits")
	}
	counterBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(counterBytes, counter)
	mac := hmac.New(sha1.New, key)
	mac.Write(counterBytes)
	hs := mac.Sum(nil)
	offset := hs[len(hs)-1] & 0x0f
	truncated := hs[offset : offset+4]
	truncated[0] &= 0x7f
	snum := binary.BigEndian.Uint32(truncated)
	d := snum % (uint32(math.Pow10(digits)))
	otp := strconv.Itoa(int(d))
	for len(otp) < digits {
		otp = "0" + otp
	}
	return otp
}

func VerifyHOTP(key []byte, counter uint64, digits int, otp string) bool {
	return GenerateHOTP(key, counter, digits) == otp
}

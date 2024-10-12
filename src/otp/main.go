package otp

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/binary"
	"math"
	"strconv"
	"time"
)

func GenerateTOTP(key []byte, interval time.Duration, digits int) string {
	counter := uint64(time.Now().Unix()) / uint64(interval.Seconds())
	return GenerateHOTP(key, counter, digits)
}

func VerifyTOTP(key []byte, interval time.Duration, digits int, otp string) bool {
	return GenerateTOTP(key, interval, digits) == otp
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

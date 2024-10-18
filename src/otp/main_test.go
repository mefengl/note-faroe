package otp

import (
	"fmt"
	"testing"
)

func TestGenerateHOTP(t *testing.T) {
	key := make([]byte, 20)
	for i := 0; i < len(key); i++ {
		key[i] = 0xff
	}
	tests := []struct {
		counter  uint64
		expected string
	}{
		{
			0, "103905",
		},
		{
			1, "463444",
		},
		{
			10, "413510",
		},
		{
			100, "632126",
		},
		{
			10000, "529078",
		},
		{
			100000000, "818472",
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("Counter: %d", test.counter), func(t *testing.T) {
			result := GenerateHOTP(key, test.counter, 6)
			if result != test.expected {
				t.Errorf("got %s, expected %s", result, test.expected)
			}
		})
	}
}

func TestVerifyHOTP(t *testing.T) {
	key := make([]byte, 20)
	for i := 0; i < len(key); i++ {
		key[i] = 0xff
	}
	validTests := []struct {
		counter uint64
		otp     string
	}{
		{
			0, "103905",
		},
		{
			1, "463444",
		},
		{
			10, "413510",
		},
		{
			100, "632126",
		},
		{
			10000, "529078",
		},
		{
			100000000, "818472",
		},
	}
	invlaidTests := []struct {
		counter uint64
		otp     string
	}{
		{
			0, "103906",
		},
	}

	for _, test := range validTests {
		t.Run(fmt.Sprintf("Counter: %d", test.counter), func(t *testing.T) {
			result := VerifyHOTP(key, test.counter, 6, test.otp)
			if !result {
				t.Error("got false, expected true")
			}
		})
	}
	for _, test := range invlaidTests {
		t.Run(fmt.Sprintf("Counter: %d", test.counter), func(t *testing.T) {
			result := VerifyHOTP(key, test.counter, 6, test.otp)
			if result {
				t.Error("got true, expected false")
			}
		})
	}

}

package argon2id

import "testing"

func Test(t *testing.T) {
	hash, err := Hash("123456")
	if err != nil {
		t.Fatal(err)
	}
	valid, err := Verify(hash, "123456")
	if err != nil {
		t.Fatal(err)
	}
	if !valid {
		t.Fatalf("Expected hash to match")
	}
	valid, err = Verify(hash, "12345")
	if err != nil {
		t.Fatal(err)
	}
	if valid {
		t.Fatalf("Expected hash to not match")
	}
}

package main

import (
	"forest/auth/forms"
	"testing"
)

func TestSignInFormIsFilled(t *testing.T) {
	testForm := forms.SignInForm{Email: "test@test.com", Password: []byte("12345678")}
	filled := testForm.IsFilled()
	if !filled {
		t.Error("testForm.filled = false; want true")
	}
}

package main

import (
	"encoding/base64"
	"testing"
)

func TestStdioRequested(t *testing.T) {
	if !stdioRequested([]string{"-stdio"}) || !stdioRequested([]string{"--stdio"}) {
		t.Fatal("stdio flag not recognised")
	}
	if stdioRequested(nil) || stdioRequested([]string{"-other"}) {
		t.Fatal("stdio enabled without the flag")
	}
}

func TestStdioCredentials(t *testing.T) {
	env := func(m map[string]string) func(string) string { return func(k string) string { return m[k] } }
	basic := "Basic " + base64.StdEncoding.EncodeToString([]byte("me@example.com:secret:with:colons"))

	email, password, err := stdioCredentials(env(map[string]string{"THINGS_AUTH": basic}))
	if err != nil || email != "me@example.com" || password != "secret:with:colons" {
		t.Fatalf("THINGS_AUTH: got %q %q %v", email, password, err)
	}
	email, password, err = stdioCredentials(env(map[string]string{"THINGS_EMAIL": "a@b.c", "THINGS_PASSWORD": "pw"}))
	if err != nil || email != "a@b.c" || password != "pw" {
		t.Fatalf("THINGS_EMAIL/PASSWORD: got %q %q %v", email, password, err)
	}
	if _, _, err := stdioCredentials(env(map[string]string{})); err == nil {
		t.Fatal("missing credentials accepted")
	}
	if _, _, err := stdioCredentials(env(map[string]string{"THINGS_AUTH": "Basic !!!"})); err == nil {
		t.Fatal("malformed THINGS_AUTH accepted")
	}
}

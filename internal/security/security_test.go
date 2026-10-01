package security

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"
)

func TestSecretsSurviveRestartAndRejectWrongKeyOrIdentity(t *testing.T) {
	dir := t.TempDir()
	v, e := Open(dir, "")
	if e != nil {
		t.Fatal(e)
	}
	sealed := v.Encrypt("secret-a", "private-token")
	v2, e := Open(dir, "")
	if e != nil {
		t.Fatal(e)
	}
	p, e := v2.Decrypt("secret-a", sealed)
	if e != nil || p != "private-token" {
		t.Fatalf("decrypt: %q %v", p, e)
	}
	if _, e = v2.Decrypt("secret-b", sealed); e == nil {
		t.Fatal("ciphertext rebound to wrong identity")
	}
	wrong, e := Open(t.TempDir(), base64.StdEncoding.EncodeToString(make([]byte, 32)))
	if e != nil {
		t.Fatal(e)
	}
	if _, e = wrong.Decrypt("secret-a", sealed); e == nil {
		t.Fatal("wrong key accepted")
	}
	b, _ := base64.StdEncoding.DecodeString(sealed)
	b[len(b)-1] ^= 1
	if _, e = v2.Decrypt("secret-a", base64.StdEncoding.EncodeToString(b)); e == nil {
		t.Fatal("tamper accepted")
	}
	info, e := os.Stat(filepath.Join(dir, "encryption.key"))
	if e != nil {
		t.Fatal(e)
	}
	if info.Mode().Perm() != 0600 {
		t.Fatalf("key permission %o", info.Mode().Perm())
	}
}
func TestPasswordsAndTokenHash(t *testing.T) {
	h, e := HashPassword("correct horse battery")
	if e != nil {
		t.Fatal(e)
	}
	if !CheckPassword(h, "correct horse battery") || CheckPassword(h, "incorrect password") {
		t.Fatal("password verification")
	}
	if _, e = HashPassword("short"); e == nil {
		t.Fatal("short password accepted")
	}
	a, b := Token(), Token()
	if a == b || HashToken(a) == a || !Equal(HashToken(a), HashToken(a)) || Equal(a, b) {
		t.Fatal("token generation/comparison")
	}
}

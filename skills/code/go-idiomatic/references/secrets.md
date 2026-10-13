# Secrets Reference

How tokens and sensitive columns are stored, and what a credential change revokes. Assume the
database is read by someone it should not be: a leaked backup, a replica, a support query.

## Contents

- [Tokens Are Stored Hashed](#tokens-are-stored-hashed)
- [Credential Changes Revoke Sessions](#credential-changes-revoke-sessions)
- [Encryption at Rest](#encryption-at-rest)

---

## Tokens Are Stored Hashed

Session tokens, password-reset tokens and API keys are stored as a hash, never in plaintext. A
plaintext token in a dump is a working login.

Tokens are random (32 bytes from `crypto/rand`), so a fast hash is enough: SHA-256 of the token,
looked up by equality. bcrypt or argon2 is for low-entropy passwords only, and its per-row salt
would make the lookup impossible.

```go
func newToken() (token string, hash []byte, err error) {
    b := make([]byte, 32)
    if _, err := rand.Read(b); err != nil {
        return "", nil, fmt.Errorf("read random: %w", err)
    }
    token = base64.RawURLEncoding.EncodeToString(b)
    sum := sha256.Sum256([]byte(token))
    return token, sum[:], nil
}
```

The token goes to the client once; only `hash` is written. Lookups hash the presented token and
query by the hash.

---

## Credential Changes Revoke Sessions

A password reset or password change deletes every session belonging to **that user**, in the same
transaction as the credential write. A reset usually means someone else had the password; leaving
their session alive defeats the reset.

Scope the delete to the user. `DELETE FROM sessions WHERE user_id = $1` is right; a delete keyed on
anything broader logs everyone out. Re-issue a session for the request that made the change if the
user should stay signed in.

---

## Encryption at Rest

A column holding a third-party credential or other secret the app must read back is encrypted with
an AEAD (AES-GCM, or XChaCha20-Poly1305 via `golang.org/x/crypto`).

- **Bind the ciphertext to its row with associated data**: table name and owner id. Without it, an
  attacker with write access copies another user's ciphertext into their own row and the app
  decrypts it for them.
- **Prefix every ciphertext with a version or key id** (`v1:` + nonce + sealed bytes). Rotation and
  format changes then read old and new rows side by side.
- **Rotate lazily**: decrypt any known version, always encrypt with the current one, so each row is
  rewritten on its next write. Then run a one-off sweep command over the rows still on the old
  version, then delete support for the legacy format.

```go
func (c *Cipher) Seal(table string, ownerID int64, plaintext []byte) ([]byte, error) {
    nonce := make([]byte, c.aead.NonceSize())
    if _, err := rand.Read(nonce); err != nil {
        return nil, fmt.Errorf("read nonce: %w", err)
    }
    aad := []byte(fmt.Sprintf("%s:%d", table, ownerID))
    out := append([]byte(c.version+":"), nonce...)
    return c.aead.Seal(out, nonce, plaintext, aad), nil
}
```

The key comes from config parsed in `main`, never from a constant or a column in the same database.

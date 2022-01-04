package cookieauth

// Это все можно было бы сделать миддлаварой и класть UserID в контекст. Много где так делается (например chi JWTAuth)
// но раз у нас в курсе говорилось, что данные аутентификации в контекст класть не стоит, то в учебном проекте сделаем отдельными методами с непосредственным вызовом

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

type CookieAuth struct {
	key        []byte
	cookieName string
}

var (
	ErrNoTokenFound = errors.New("no token found")
	ErrInvalidToken = errors.New("token is invalid")
)

func New(key []byte, cookieName string) *CookieAuth {
	ca := &CookieAuth{
		key:        key,
		cookieName: cookieName,
	}
	return ca
}

func (ca *CookieAuth) GetUserID(r *http.Request) (string, error) {
	tokenString := ca.getTokenFromCookie(r)
	if tokenString == "" {
		return "", ErrNoTokenFound
	}

	return ca.verifyToken(tokenString)
}

func (ca *CookieAuth) SetUserIDCookie(w http.ResponseWriter, uid string) {
	token := fmt.Sprintf("%s:%s", uid, ca.calcHash(uid))
	cookie := &http.Cookie{
		Name:  ca.cookieName,
		Value: token,
	}
	http.SetCookie(w, cookie)
}

func (ca *CookieAuth) verifyToken(tokenString string) (string, error) {
	parts := strings.Split(tokenString, ":")
	if len(parts) != 2 {
		return "", ErrInvalidToken
	}
	uid, h := parts[0], parts[1]
	if !ca.checkHash(uid, h) {
		return "", ErrInvalidToken
	}
	return uid, nil
}

func (ca *CookieAuth) getTokenFromCookie(r *http.Request) string {
	cookie, err := r.Cookie(ca.cookieName)
	if err != nil {
		return ""
	}
	return cookie.Value
}

func (ca *CookieAuth) calcHash(uid string) string {
	h := hmac.New(sha256.New, ca.key)
	h.Write([]byte(uid))
	return hex.EncodeToString(h.Sum(nil))
}

func (ca *CookieAuth) checkHash(uid string, hash string) bool {
	return hash == ca.calcHash(uid)
}

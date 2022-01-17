package cookieauth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"net/http"
	"strings"
)

type CookieAuth struct {
	key []byte
}

var (
	ErrNoTokenFound = errors.New("no token found")
	ErrInvalidToken = errors.New("token is invalid")
)

var (
	UserIDCtxKey = &contextKey{"UserID"}
	ErrorCtxKey  = &contextKey{"Error"}
)

var cookieName = "UserID"

func New(key []byte) *CookieAuth {
	ca := &CookieAuth{
		key: key,
	}
	return ca
}

// Verifier проверяет наличие куки UserId и прописывает в контекст UserID если кука есть и валидна, и ошибку, если не найдено или невалидно.
func Verifier(ca *CookieAuth) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return verify(ca)(next)
	}
}

// Authenticator проверяет наличие в контексте UserID или ошибок (установленных в Verifier). Если ошибка - отсутствие куки или невалидное значение куки - устанавливает новые
func Authenticator(ca *CookieAuth) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return authenticate(ca)(next)
	}
}

// FromContext достает из контекста UserID и ошибки связанные с аутентификацией
func FromContext(ctx context.Context) (string, error) {
	token, _ := ctx.Value(UserIDCtxKey).(string)
	err, _ := ctx.Value(ErrorCtxKey).(error)
	return token, err
}

func verify(ca *CookieAuth) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		hfn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			token, err := verifyRequest(ca, r)
			ctx = newContext(ctx, token, err)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
		return http.HandlerFunc(hfn)
	}
}

func verifyRequest(ca *CookieAuth, r *http.Request) (string, error) {
	tokenString := getTokenFromCookie(r)
	if tokenString == "" {
		return "", ErrNoTokenFound
	}

	return verifyToken(ca, tokenString)
}

func verifyToken(ca *CookieAuth, tokenString string) (string, error) {
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

func authenticate(ca *CookieAuth) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		hfn := func(w http.ResponseWriter, r *http.Request) {
			_, err := FromContext(r.Context())
			if err != nil {
				if errors.Is(err, ErrNoTokenFound) || errors.Is(err, ErrInvalidToken) {
					userID := uuid.NewString()
					ca.setUserIDCookie(w, userID)
					ctx := newContext(r.Context(), userID, nil)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)

		}
		return http.HandlerFunc(hfn)
	}
}

func newContext(ctx context.Context, uid string, err error) context.Context {
	ctx = context.WithValue(ctx, UserIDCtxKey, uid)
	ctx = context.WithValue(ctx, ErrorCtxKey, err)
	return ctx
}

func (ca *CookieAuth) setUserIDCookie(w http.ResponseWriter, uid string) {
	token := fmt.Sprintf("%s:%s", uid, ca.calcHash(uid))
	cookie := &http.Cookie{
		Name:  cookieName,
		Value: token,
		Path:  "/",
	}
	http.SetCookie(w, cookie)
}

func getTokenFromCookie(r *http.Request) string {
	cookie, err := r.Cookie(cookieName)
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

// Утащено из go-chi/jwtauth
// contextKey is a value for use with context.WithValue. It's used as
// a pointer so it fits in an interface{} without allocation. This technique
// for defining context keys was copied from Go 1.7's new use of context in net/http.
type contextKey struct {
	name string
}

func (k *contextKey) String() string {
	return "cookieauth context value " + k.name
}

package auth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
)

const secretKey = "abwLpqp4uKfxiQJIbfYIudou7K7qbtXE"
const cookieName = "login"

// struct to pass user id through request context
type requestContextUserIDValue struct{}

func ValidateCookie(cookie *http.Cookie) (bool, error) {
	return validateID(cookie.Value)
}

func validateID(cookieValue string) (bool, error) {
	if cookieValue == "" {
		return false, nil
	}

	decodedCookieVal, err := hex.DecodeString(cookieValue)
	if err != nil {
		return false, err
	}

	id := getIDFromCookieVal(decodedCookieVal)

	valid := validateIDWithSignature(id, decodedCookieVal)
	if !valid {
		return false, nil
	}
	return true, nil
}

// id is the first 36 bytes of cookie value
func getIDFromCookieVal(decodedCookieVal []byte) (id string) {
	return string(decodedCookieVal[:36])
}

// To validate ID we should create a signature from id and compare it with signature in request
func validateIDWithSignature(id string, decodedCookieVal []byte) (valid bool) {
	validSignature := string(createSignature([]byte(id)))

	requestSignature := string(decodedCookieVal[36:])

	return validSignature == requestSignature
}

func CreateCookie(id string) *http.Cookie {

	byteID := []byte(id)

	//create signature
	sign := createSignature(byteID)

	//create final hex string for value in cookie
	cookieValue := concatIDAndSignature(byteID, sign)
	return &http.Cookie{Name: cookieName, Value: cookieValue}

}

func createSignature(src []byte) (sign []byte) {
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write(src)
	sign = h.Sum(nil)

	return sign
}

// concat id and sign. It should be two hex strings, NOT slices of bytes !!!
func concatIDAndSignature(byteID []byte, sign []byte) (resultStr string) {

	strEncodedID := hex.EncodeToString(byteID)
	strSign := hex.EncodeToString(sign)
	resultStr = strEncodedID + strSign
	return resultStr
}

func SetUUIDToRequestContext(r *http.Request, id string) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), requestContextUserIDValue{}, id))

}

func GetIDFromCookie(cookieValue string) (id string) {
	decodedCookieVal, _ := hex.DecodeString(cookieValue)
	id = getIDFromCookieVal(decodedCookieVal)
	return id
}

func GetUUIDFromContext(ctx context.Context) string {
	userID, _ := ctx.Value(requestContextUserIDValue{}).(string)
	return userID
}

package rand

import (
	cr "crypto/rand"
	"encoding/base32"
)

var encoding *base32.Encoding

func init (){
	encoding = base32.NewEncoding("0123456789abcdefghijklmnopqrstuv").WithPadding(base32.NoPadding)
}

func String(n int) string {
	b := make([]byte, encoding.DecodedLen(n + 7))
	_, err := cr.Read(b)
	if err != nil {
		panic("crypto/rand failed: "+err.Error())
	}
	return encoding.EncodeToString(b)[0:n]
}

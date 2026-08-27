package password

import (
	"crypto/rand"
	"math/big"
	"runtime"

	"github.com/alexedwards/argon2id"
)

var defaultHashParams = argon2id.Params{
	Memory:      64 * 1024,
	Iterations:  4,
	Parallelism: uint8(runtime.NumCPU()),
	SaltLength:  16,
	KeyLength:   32,
}

func Hash(password string) (string, error) {
	return argon2id.CreateHash(password, &defaultHashParams)
}

func Verify(hashedPassword, password string) (bool, error) {
	return argon2id.ComparePasswordAndHash(password, hashedPassword)
}

const generatedChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func GenerateAPIKey() (string, error) {
	const length = 32
	out := make([]byte, length)
	max := big.NewInt(int64(len(generatedChars)))
	for i := range out {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		out[i] = generatedChars[n.Int64()]
	}
	return "sa_" + string(out), nil
}

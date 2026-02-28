package utils

import (
	"crypto/rand"
	"fmt"
	"math"
	"math/big"
	"strconv"
)

func GenerateCryptoRandomInt(n uint8) (string, error) {
	if n == 0 || n > 9 {
		return "", fmt.Errorf("n must be between 1 and 9")
	}
	min := int64(math.Pow10((int(n) - 1)))
	max := int64(math.Pow10(int(n)))
	rangeSize := big.NewInt(max - min)

	number, err := rand.Int(rand.Reader, rangeSize)
	if err != nil {
		return "", err
	}

	return strconv.FormatInt(number.Int64()+min, 10), nil
}

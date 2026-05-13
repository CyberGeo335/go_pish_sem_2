package cache

import (
	cryptorand "crypto/rand"
	"math/big"
	"time"
)

func TTLWithJitter(base time.Duration, jitter time.Duration) time.Duration {
	if jitter <= 0 {
		return base
	}

	max := big.NewInt(int64(jitter) + 1)
	extra, err := cryptorand.Int(cryptorand.Reader, max)
	if err != nil {
		return base
	}

	return base + time.Duration(extra.Int64())
}

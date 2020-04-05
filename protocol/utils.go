package protocol

import (
	"math/rand"
	"time"
)

// Returns a random duration between min and max
func getRandomDuration(min time.Duration, max time.Duration) time.Duration {
	return time.Duration(rand.Intn(int(max-min))) + min
}

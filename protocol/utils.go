package protocol

import (
	"math/rand"
	"time"
)

// Returns a random duration between min and max
func getRandomDuration(min time.Duration, max time.Duration) time.Duration {
	return time.Duration(rand.Intn(int(max-min))) + min
}

func isMajority(total int, actual int) bool {
	return total/2+1 <= actual
}
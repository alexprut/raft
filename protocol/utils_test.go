package protocol

import "testing"

func TestGetRandomDuration(t *testing.T) {
	duration := getRandomDuration(minElectionMsTimeout, maxElectionMsTimeout)
	if duration < minElectionMsTimeout || duration > maxElectionMsTimeout {
		t.Errorf("getRandomDuration() is not in the range [%d, %d], it is: %d", minElectionMsTimeout.Milliseconds(), maxElectionMsTimeout.Milliseconds(), duration.Milliseconds())
	}
}

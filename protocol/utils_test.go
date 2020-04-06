package protocol

import "testing"

func TestGetRandomDuration(t *testing.T) {
	duration := getRandomDuration(minElectionMsTimeout, maxElectionMsTimeout)
	if duration < minElectionMsTimeout || duration > maxElectionMsTimeout {
		t.Errorf("getRandomDuration() is not in the range [%d, %d], it is: %d", minElectionMsTimeout.Milliseconds(), maxElectionMsTimeout.Milliseconds(), duration.Milliseconds())
	}
}

func TestIsMajority(t *testing.T) {
	if isMajority(1, 0) || isMajority(2, 1) || isMajority(3, 1) || isMajority(4, 2) {
		t.Errorf("Should not be: True")
	}
	if !isMajority(1, 1) || !isMajority(2, 2) || !isMajority(3, 2) || !isMajority(4, 3) {
		t.Errorf("Should not be: False")
	}
}

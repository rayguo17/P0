// Tests for SquarerImpl. Students should write their code in this file.

package p0partB

import (
	"fmt"
	"testing"
	"time"
)

const (
	timeoutMillis = 5000
)

func TestBasicCorrectness(t *testing.T) {
	fmt.Println("Running TestBasicCorrectness.")
	input := make(chan int)
	sq := SquarerImpl{}
	squares := sq.Initialize(input)
	go func() {
		input <- 2
	}()
	timeoutChan := time.After(time.Duration(timeoutMillis) * time.Millisecond)
	select {
	case <-timeoutChan:
		t.Error("Test timed out.")
	case result := <-squares:
		if result != 4 {
			t.Error("Error, got result", result, ", expected 4 (=2^2).")
		}
	}
}

func TestYourFirstGoTest(t *testing.T) {
	// TODO: Write a test here.
	fmt.Println("Test Initialization return")
	input := make(chan int)
	sq := SquarerImpl{}
	squares := sq.Initialize(input)
	for i := 0; i < 10; i++ {
		l := i
		input <- l

	}
	timeoutCHan := time.After(time.Duration(timeoutMillis) * time.Millisecond)
	for i := 0; i < 10; i++ {
		select {
		case <-timeoutCHan:
			t.Error("Test timed out.")
		case result := <-squares:
			if result != i*i {
				t.Error("Error, got result", result, ", expected", i*i, ".")
			}
		}

	}
}

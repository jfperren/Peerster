package tests

import (
	"github.com/jfperren/Peerster/gossiper"
	"testing"
)

func TestMatch(t *testing.T) {

	if !gossiper.Match("hello.txt", []string{"hell"}) {
		t.Errorf("'hell' should match 'hello.txt'")
	}

	if !gossiper.Match("hello.txt", []string{".txt"}) {
		t.Errorf("'.txt' should match 'hello.txt'")
	}

	if !gossiper.Match("hello.txt", []string{"ll"}) {
		t.Errorf("'ll' should match 'hello.txt'")
	}

	if !gossiper.Match("message33.txt", []string{"ge33"}) {
		t.Errorf("'ge33' should match 'message33.txt'")
	}

	if !gossiper.Match("hello.txt", []string{"ab", "hell", "cd"}) {
		t.Errorf("'ab,hell,cd' should match 'hello.txt'")
	}

	if gossiper.Match("hello.txt", []string{"ab"}) {
		t.Errorf("'ab' should not match 'hello.txt'")
	}

	if gossiper.Match("hello.txt", []string{"adf", "bsd"}) {
		t.Errorf("'adf,bsd' should not match 'hello.txt'")
	}

	if gossiper.Match("hello.txt", []string{"*.txt"}) {
		t.Errorf("'*.txt' should not match 'hello.txt'")
	}
}

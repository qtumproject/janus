package notifier

import (
	"testing"
	"time"
)

func TestRollingLimit(t *testing.T) {
	l := newRollingLimit(4)
	if l.oldest() != 1 {
		t.Fatalf("Expected oldest slot to be 1, not %d", l.oldest())
	}
	if l.newest() != 0 {
		t.Fatalf("Expected newest slot to be 0, not %d", l.newest())
	}
	now := time.Now()
	first := now.Add(-1 * time.Second)
	second := now
	third := now.Add(1 * time.Second)
	fourth := third.Add(1 * time.Second)

	l.Push(first)
	if l.oldest() != 2 {
		t.Fatalf("Expected oldest slot to be 2, not %d", l.oldest())
	}
	if l.newest() != 1 {
		t.Fatalf("Expected newest slot to be 1, not %d", l.newest())
	}

	l.Push(second)
	if l.oldest() != 3 {
		t.Fatalf("Expected oldest slot to be 3, not %d", l.oldest())
	}
	if l.newest() != 2 {
		t.Fatalf("Expected newest slot to be 2, not %d", l.newest())
	}

	l.Push(third)
	if l.oldest() != 0 {
		t.Fatalf("Expected oldest slot to be 0, not %d", l.oldest())
	}
	if l.newest() != 3 {
		t.Fatalf("Expected newest slot to be 3, not %d", l.newest())
	}

	l.Push(fourth)
	if l.oldest() != 1 {
		t.Fatalf("Expected oldest slot to be 1, not %d", l.oldest())
	}
	if l.newest() != 0 {
		t.Fatalf("Expected newest slot to be 0, not %d", l.newest())
	}
}

package userd

import (
	"fmt"
	"testing"
	"time"
)

func TestClient(t *testing.T) {

	client := NewClient(3*time.Second, 100*time.Millisecond)

	err := client.Dial()
	if err != nil {
		t.Error(err)
	}

	defer client.Hangup()

	if _, err := client.Check("hello"); err != nil {

		t.Error(err)
	}
}

func TestControl(t *testing.T) {

	ctrl := NewControl(3 * time.Second)

	err := ctrl.Dial()
	if err != nil {
		t.Error(err)
	}

	defer ctrl.Hangup()

	if _, err := ctrl.Set("hello"); err != nil {
		t.Error(err)
	}

	if _, err := ctrl.Check("hello"); err != nil {
		t.Error(err)
	}

	if _, err := ctrl.Remove("hello"); err != nil {
		t.Error(err)
	}

	if _, err := ctrl.Purge(); err != nil {
		t.Error(err)
	}
}

func BenchmarkClientCheck(b *testing.B) {

	client := NewClient(1*time.Millisecond, 3*time.Second)
	err := client.Dial()
	if err != nil {
		b.Error(err)
	}
	defer client.Hangup()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {

		if _, err := client.Check("hello"); err != nil {

			b.Error(err)
		}
	}
}

func BenchmarkClientCheckL(b *testing.B) {

	client := NewClient(1*time.Second, 3*time.Second)
	err := client.Dial()
	if err != nil {
		b.Error(err)
	}
	defer client.Hangup()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := client.Check("hello"); err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkControlSet(b *testing.B) {

	ctrl := NewControl(3 * time.Second)
	err := ctrl.Dial()
	if err != nil {
		b.Error(err)
	}

	defer ctrl.Hangup()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {

		if _, err := ctrl.Set(fmt.Sprintf("hello-%d", i)); err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkControlCheck(b *testing.B) {

	ctrl := NewControl(3 * time.Second)
	err := ctrl.Dial()
	if err != nil {
		b.Error(err)
	}
	defer ctrl.Hangup()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := ctrl.Check(fmt.Sprintf("hello-%d", i)); err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkControlRemove(b *testing.B) {

	ctrl := NewControl(3 * time.Second)
	err := ctrl.Dial()
	if err != nil {
		b.Error(err)
	}
	defer ctrl.Hangup()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := ctrl.Remove(fmt.Sprintf("hello-%d", i)); err != nil {
			b.Error(err)
		}
	}
}

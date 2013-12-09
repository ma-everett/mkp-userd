
package userd

import (
	"testing"
	"time"
)

func TestClient(t *testing.T) {

	client := NewClient(3 * time.Second,100 * time.Millisecond)
	
	err := client.Dial()
	if err != nil {
		t.Error(err)
	}

	defer client.Hangup()

	
	if _,err := client.Check("hello"); err != nil {

		t.Error(err)
	}
}

func BenchmarkClientCheck(b *testing.B) {

	client := NewClient(3 * time.Second,100 * time.Millisecond)
	err := client.Dial()
	if err != nil {
		b.Error(err)
	}

	defer client.Hangup()
	
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		
		if _,err := client.Check("hello"); err != nil {
			
			b.Error(err)
		}
	}
}

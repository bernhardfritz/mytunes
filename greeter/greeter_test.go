package greeter

import "testing"

func TestGreet(t *testing.T) {
	want := "Hello Bernhard"
	if got := Greet("Bernhard"); got != want {
		t.Errorf("Greet(\"Bernhard\") = %q, want %q", got, want)
	}
}

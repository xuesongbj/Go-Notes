package closure

import (
    "testing"
)

func TestClosure(t *testing.T) {
    f := Closure(100)
    f()
}

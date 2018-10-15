package main

import "testing"

func TestSomething(t *testing.T) {

    x := true
    if !x {
       t.Errorf("Whoops")
    }

}

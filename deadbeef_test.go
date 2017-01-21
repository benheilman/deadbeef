package deadbeef

import (
	"strconv"
	"testing"
)

var (
	chanCount int = 10
)

func TestReflector(t *testing.T) {
	outbound := make(chan string)
	inbound := Reflector(outbound)
	outbound <- "a"

	if in := <-inbound; in != "a a" {
		t.Errorf("Unexpected output %s", in)
	}

	outbound <- "b"
	close(outbound)
	cases := 0
	for perm := range inbound {
		if perm != "a b" && perm != "b a" && perm != "b b" {
			t.Errorf("Unexpected output %s", perm)
		}
		cases++
	}
	if cases != 3 {
		t.Errorf("There should be 3 perms, not %d", cases)
	}
}

func TestCombiner(t *testing.T) {
	left, right := make(chan string), make(chan string)
	out := Combiner(left, right)
	left <- "a"
	right <- "bb"
	close(left)
	close(right)

	cases := 0
	for perm := range out {
		if perm != "a bb" && perm != "bb a" {
			t.Errorf("Unexpected output %s", perm)
		}
		cases++
	}
	if cases != 2 {
		t.Errorf("There should be 2 perms, not %d", cases)
	}

}

func TestRepeater(t *testing.T) {
	message := "test"
	in1, in2 := make(chan string, 10), make(chan string, 10)
	out := Repeater(in1, in2)
	out <- message
	for _, ch := range []<-chan string{in1, in2} {
		if in, ok := <-ch; ok {
			if in != message {
				t.Errorf("output %s does not match message", in)
			}
		} else {
			t.Errorf("input channel is closed")
		}
	}

	close(out)
	for _, ch := range []<-chan string{in1, in2} {
		if _, ok := <-ch; ok {
			t.Errorf("output channel should be closed")
		}
	}
}

func TestSplitter(t *testing.T) {
	outbound := make(chan string)
	inbounds := Splitter(outbound, chanCount)

	for i := 0; i < chanCount; i++ {
		output := ""
		for j := 0; j < i; j++ {
			output = output + "a"
		}
		outbound <- output
		if in, ok := <-inbounds[i]; ok {
			if in != output {
				t.Errorf("Channel[%d] should not have received an '%s'", i, in)
			}
		}
		for j := 0; j < chanCount; j++ {
			if j == i {
				continue
			}
			select {
			case in := <-inbounds[j]:
				t.Errorf("Channel[%d] should not have received an '%s'", j, in)
			default:
			}
		}
	}
}

func TestMerge(t *testing.T) {
	ins := []<-chan string{}
	for i := 0; i < chanCount; i++ {
		in := make(chan string, 10)
		in <- strconv.Itoa(i)
		ins = append(ins, in)
		close(in)
	}
	out := Merge(ins...)
	count := 0
	for range out {
		count++
	}
	if count != chanCount {
		t.Errorf("Merged channel had %d messages instead of %d", count, chanCount)
	}
}

func TestGraph(t *testing.T) {
	outbound := make(chan string, 10)
	inbound := Graph(outbound, 3)
	outbound <- "a"
	outbound <- "bb"
	close(outbound)
	if in := <-inbound; in != "bb a" && in != "a bb" {
		t.Errorf("Unexpected message %s", in)
	}
}

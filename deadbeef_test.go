package deadbeef

import (
	"strconv"
	"testing"
)

var (
	chanCount int = 10
)

func TestCombiner(t *testing.T) {
	left, right := make(chan string), make(chan string)
	out := Combiner(left, right)
	left <- "a"
	right <- "bb"
	close(left)
	close(right)
	
	cases := 0
	for perm := range out {
		if perm != "a bb" && perm != "bb a"{
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
	inboundMap := Splitter(outbound, 2)
	if len(inboundMap[1]) != 2 {
		t.Errorf("Center length does not have two channels")
	}

	outbound <- ""
	if in, ok := <-inboundMap[0][0]; ok {
		if in != "" {
			t.Errorf("Zero length channel should have received an empty string")
		}
	}
	select {
	case in := <-inboundMap[1][0]:
		t.Errorf("One length channel should not have received %s", in)
	case in := <-inboundMap[2][0]:
		t.Errorf("Two length channel should not have received %s", in)
	default:
	}

	outbound <- "a"
	for i, ch := range inboundMap[1] {
		if in, ok := <-ch; ok {
			if in != "a" {
				t.Errorf("one length channel[%d] should not have received an %s", i, in)
			}
		}
	}
	select {
	case in := <-inboundMap[0][0]:
		t.Errorf("Zero length channel should not have received %s", in)
	case in := <-inboundMap[2][0]:
		t.Errorf("Two length channel should not have received %s", in)
	default:
	}
}

func TestDuplicate(t *testing.T) {
	input := make(chan string, 10)
	input <- "test"
	outputs := Duplicate(input, chanCount)
	for i := 0; i < chanCount; i++ {
		message, ok := <-outputs[i]
		if !ok {
			t.Errorf("output channel closed")
		}
		if message != "test" {
			t.Errorf("message was not test")
		}
	}
	close(input)
	for i := 0; i < chanCount; i++ {
		_, ok := <-outputs[i]
		if ok {
			t.Errorf("output channel was not closed after input was")
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
	if in := <- inbound; in != "bb a" && in != "a bb" {
		t.Errorf("Unexpected message %s", in)
	}
}

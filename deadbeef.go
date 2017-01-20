// Read words and find all combinations that can be made using hex representation
package deadbeef

import (
	"regexp"
	"strings"
	"sync"
)

var regex *regexp.Regexp = regexp.MustCompile("^[a-f]*$")

type cache struct {
	lefts []string
	rights []string
	mux *sync.Mutex
	wg *sync.WaitGroup
}

func newCache() *cache {
	c := new(cache)
	c.mux = new(sync.Mutex)
	c.wg = new(sync.WaitGroup)
	c.wg.Add(2)
	return c
}

func Combiner(left <-chan string, right <-chan string) <-chan string {
	outbound := make(chan string, 10)
	c := newCache()

	go func() {
		for l := range left {
			c.mux.Lock()
			c.lefts = append(c.lefts, l)
			for _, r := range c.rights {
				outbound <- l + " " + r
				outbound <- r + " " + l
			}
			c.mux.Unlock()
		}
		c.wg.Done()
	}()

	go func() {
		for r := range right {
			c.mux.Lock()
			c.rights = append(c.rights, r)
			for _, l := range c.lefts {
				outbound <- l + " " + r
				outbound <- r + " " + l
			}
			c.mux.Unlock()
		}
		c.wg.Done()
	}()

	go func() {
		c.wg.Wait()
		close(outbound)
	}()

	return outbound

}

func Repeater(outbounds ...chan<- string) chan<- string {
	inbound := make(chan string, 10)
	go func() {
		for in := range inbound {
			for _, out := range outbounds {
				out <- in
			}
		}
		for _, out := range outbounds {
			close(out)
		}
	}()
	return inbound
}

func Splitter(inbound <-chan string, stringLen int) map[int][]<-chan string {
	outbounds := make(map[int][]<-chan string)
	splits := make(map[int]chan<- string)
	for i, j := 0, stringLen; i <= j; i, j = i+1, j-1 {
		ch1, ch2 := make(chan string, 10), make(chan string, 10)
		if i == j {
			outbounds[i] = []<-chan string{ch1, ch2}
			splits[i] = Repeater(ch1, ch2)
		} else {
			outbounds[i] = []<-chan string{ch1}
			splits[i] = ch1
			outbounds[j] = []<-chan string{ch2}
			splits[j] = ch2
		}
	}
	go func() {
		for in := range inbound {
			if length := strings.Count(in, "") - 1; length <= stringLen {
				splits[length] <- in
			}
		}
		for _, split := range splits {
			close(split)
		}
	}()
	return outbounds
}

func Matcher(in <-chan string) <-chan string {
	out := make(chan string, 10)
	go func() {
		for s := range in {
			if regex.MatchString(s) {
				out <- s
			}
		}
		close(out)
	}()
	return out
}

func Duplicate(input <-chan string, num int) []<-chan string {
	outputs := []<-chan string{}
	senders := []chan<- string{}
	for i := 0; i < num; i++ {
		ch := make(chan string, 10)
		outputs = append(outputs, ch)
		senders = append(senders, ch)
	}
	go func() {
		for in := range input {
			for _, s := range senders {
				s <- in
			}
		}
		for _, s := range senders {
			close(s)
		}
	}()
	return outputs
}

func Merge(inputs ...<-chan string) <-chan string {
	wg := new(sync.WaitGroup)
	out := make(chan string, 10)

	output := func(c <-chan string) {
		for n := range c {
			out <- n
		}
		wg.Done()
	}
	wg.Add(len(inputs))

	for _, in := range inputs {
		go output(in)
	}

	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

func Graph(inbound <-chan string, length int) <-chan string {
	primer := make(chan string, 10)
	primer <- ""
	chans := Splitter(Merge(inbound, primer), length)
	close(primer)

	combines := []<-chan string{}
	for i, j := 0, length; i <= j; i, j = i + 1, j - 1 {
		if i == j {
			combines = append(combines, Combiner(Matcher(chans[i][0]), Matcher(chans[i][1])))
		} else {
			combines = append(combines, Combiner(Matcher(chans[i][0]), Matcher(chans[j][0])))
		}
	}

	return Merge(combines...)
}

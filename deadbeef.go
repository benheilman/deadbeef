// Read words and find all combinations that can be made using hex representation
package deadbeef

import (
	"regexp"
	"strings"
	"sync"
)

var regex *regexp.Regexp = regexp.MustCompile("^[a-f]*$")

type cache struct {
	lefts  []string
	rights []string
	mux    *sync.Mutex
	wg     *sync.WaitGroup
}

func newCache(chanCount int) *cache {
	c := new(cache)
	c.mux = new(sync.Mutex)
	c.wg = new(sync.WaitGroup)
	c.wg.Add(chanCount)
	return c
}

func Reflector(inbound <-chan string) <-chan string {
	outbound := make(chan string, 10)
	c := newCache(1)

	go func() {
		for in := range inbound {
			c.mux.Lock()
			for _, l := range c.lefts {
				outbound <- in + " " + l
				outbound <- l + " " + in
			}
			outbound <- in + " " + in
			c.lefts = append(c.lefts, in)
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

func Combiner(left <-chan string, right <-chan string) <-chan string {
	outbound := make(chan string, 10)
	c := newCache(2)

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

func Splitter(inbound <-chan string, stringLen int) []<-chan string {
	outbounds := make([]<-chan string, stringLen+1)
	splits := make([]chan<- string, stringLen+1)
	for i := 0; i <= stringLen; i++ {
		ch := make(chan string, 10)
		outbounds[i] = ch
		splits[i] = ch
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
	chans := Splitter(inbound, length)

	combines := []<-chan string{}
	for i, j := 0, length; i <= j; i, j = i+1, j-1 {
		if i == 0 {
			combines = append(combines, Matcher(chans[0]), Matcher(chans[length]))
		} else if i == j {
			combines = append(combines, Reflector(Matcher(chans[i])))
		} else {
			combines = append(combines, Combiner(Matcher(chans[i]), Matcher(chans[j])))
		}
	}

	return Merge(combines...)
}

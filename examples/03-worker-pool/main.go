// Worker pool: goroutines, channels, WaitGroup, Mutex, select and pipelines.
package main

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

type Job struct {
	ID    int
	Input int
}

type Result struct {
	JobID  int
	Worker int
	Output int
}

func collatzSteps(n int) int {
	steps := 0
	for n != 1 {
		if n%2 == 0 {
			n /= 2
		} else {
			n = 3*n + 1
		}
		steps++
	}
	return steps
}

func worker(id int, jobs <-chan Job, results chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()
	for job := range jobs {
		results <- Result{job.ID, id, collatzSteps(job.Input)}
	}
}

func generate(n int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for i := 1; i <= n; i++ {
			out <- i
		}
	}()
	return out
}

func square(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for v := range in {
			out <- v * v
		}
	}()
	return out
}

func filterOdd(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for v := range in {
			if v%2 == 1 {
				out <- v
			}
		}
	}()
	return out
}

type Counter struct {
	mu sync.Mutex
	n  map[string]int
}

func (c *Counter) Inc(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.n[key]++
}

func main() {
	jobs := make(chan Job, 10)
	results := make(chan Result, 10)
	var wg sync.WaitGroup
	for w := 1; w <= 4; w++ {
		wg.Add(1)
		go worker(w, jobs, results, &wg)
	}
	go func() {
		for i := 1; i <= 20; i++ {
			jobs <- Job{i, i * 7}
		}
		close(jobs)
	}()
	go func() {
		wg.Wait()
		close(results)
	}()

	var all []Result
	for r := range results {
		all = append(all, r)
	}
	sort.Slice(all, func(i, j int) bool { return all[i].JobID < all[j].JobID })
	for _, r := range all {
		fmt.Printf("job %2d -> %3d steps\n", r.JobID, r.Output)
	}

	sum := 0
	for v := range filterOdd(square(generate(10))) {
		sum += v
	}
	fmt.Println("sum of odd squares up to 10:", sum)

	c := Counter{n: map[string]int{}}
	var wg2 sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg2.Add(1)
		go func(i int) {
			defer wg2.Done()
			c.Inc([]string{"red", "green", "blue"}[i%3])
		}(i)
	}
	wg2.Wait()
	fmt.Println(c.n["red"], c.n["green"], c.n["blue"])

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	timeout := time.After(55 * time.Millisecond)
	ticks := 0
loop:
	for {
		select {
		case <-ticker.C:
			ticks++
		case <-timeout:
			break loop
		}
	}
	fmt.Println("ticker fired repeatedly:", ticks >= 1)

	done := make(chan struct{})
	data := make(chan string, 1)
	go func() {
		select {
		case m := <-data:
			fmt.Println("got", m)
		case <-time.After(time.Second):
			fmt.Println("too slow")
		}
		close(done)
	}()
	data <- "hello from select"
	<-done
}

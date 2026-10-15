package ui

import (
	"fmt"
	"io"
	"sync"
	"time"
)

// StartGenerating writes a spinner to w (typically os.Stderr) until stop is called.
func StartGenerating(w io.Writer) (stop func()) {
	done := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		frames := []string{"|", "/", "-", "\\"}
		i := 0
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		fmt.Fprintf(w, "\rGenerating... %s", frames[0])
		for {
			select {
			case <-done:
				fmt.Fprint(w, "\r                \r")
				return
			case <-ticker.C:
				i = (i + 1) % len(frames)
				fmt.Fprintf(w, "\rGenerating... %s", frames[i])
			}
		}
	}()
	var once sync.Once
	return func() {
		once.Do(func() {
			close(done)
			wg.Wait()
		})
	}
}

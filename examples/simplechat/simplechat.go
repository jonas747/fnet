package simplechat

import (
	"fmt"
	"runtime"
	"time"
)

func Monitor() {
	tikcer := time.NewTicker(2 * time.Second)

	for {
		select {
		case <-tikcer.C:
			routines := runtime.NumGoroutine()
			var memstats runtime.MemStats
			runtime.ReadMemStats(&memstats)
			fmt.Printf("Goroutines: %d, Memory in use %dKB\n", routines, int(memstats.Alloc/1000))
		}
	}
}

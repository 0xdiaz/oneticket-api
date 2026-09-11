package main

import (
	"fmt"
	"strings"
	"time"
)

func printHeader(base string, total int) {
	fmt.Println()
	fmt.Println(strings.Repeat("=", 56))
	fmt.Println("  SMOKE TEST")
	fmt.Println(strings.Repeat("=", 56))
	fmt.Printf("  Target : %s\n", base)
	fmt.Printf("  Cek    : %d\n", total)
	fmt.Println(strings.Repeat("=", 56))
	fmt.Println()
}

func printProgress(r result) {
	mark := "OK  "
	if r.err != nil {
		mark = "GAGAL"
	}
	fmt.Printf("  [%-5s] %-40s %s\n", mark, r.name, r.elapsed.Round(time.Millisecond))
	if r.err != nil {
		fmt.Printf("           %s\n", r.err)
	}
}

// report prints the verdict and returns the exit code: 0 when every check
// passed, 1 when any failed.
func report(results []result) int {
	failed := make([]result, 0)
	var total time.Duration
	for _, r := range results {
		total += r.elapsed
		if r.err != nil {
			failed = append(failed, r)
		}
	}

	fmt.Println()
	if len(failed) == 0 {
		fmt.Println("  " + strings.Repeat("-", 54))
		fmt.Printf("  SEHAT, %d dari %d cek lolos dalam %s\n",
			len(results), len(results), total.Round(time.Millisecond))
		fmt.Println("  " + strings.Repeat("-", 54))
		fmt.Println()
		return 0
	}

	fmt.Println("  " + strings.Repeat("!", 54))
	fmt.Printf("  GAGAL, %d dari %d cek tidak lolos\n", len(failed), len(results))
	fmt.Println("  " + strings.Repeat("!", 54))
	for _, r := range failed {
		fmt.Printf("      %s\n           %s\n", r.name, r.err)
	}
	fmt.Println()
	fmt.Println("  Smoke test membuktikan wiring, bukan logika. Kegagalan di sini")
	fmt.Println("  hampir selalu soal lingkungan: server mati, migrasi belum jalan,")
	fmt.Println("  config menunjuk database yang salah, atau route hilang.")
	fmt.Println()
	return 1
}

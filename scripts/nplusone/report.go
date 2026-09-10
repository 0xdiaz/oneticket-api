package main

import (
	"fmt"
	"strings"
	"time"
)

func printHeader(sizes []int, ticketsPer int) {
	fmt.Println()
	fmt.Println(strings.Repeat("=", 56))
	fmt.Println("  QUERY COUNT PROBE — GET /api/v1/events")
	fmt.Println(strings.Repeat("=", 56))
	fmt.Printf("  Ukuran     : %v event\n", sizes)
	fmt.Printf("  Per event  : %d tiket\n", ticketsPer)
	fmt.Println(strings.Repeat("=", 56))
	fmt.Println()
}

// report prints the table and returns the exit code: 1 when the query count
// scales with the row count, 0 when it stays flat.
func report(results []result) int {
	fmt.Println("  HASIL")
	fmt.Println("  " + strings.Repeat("-", 54))
	fmt.Printf("  %8s  %10s  %12s  %s\n", "EVENT", "QUERY", "DURASI", "QUERY/EVENT")
	for _, r := range results {
		per := float64(r.queries) / float64(r.events)
		fmt.Printf("  %8d  %10d  %12s  %.2f\n",
			r.events, r.queries, r.elapsed.Round(time.Microsecond), per)
	}
	fmt.Println()

	if len(results) < 2 {
		fmt.Println("  Butuh minimal dua ukuran untuk menilai pertumbuhan.")
		return 2
	}

	first, last := results[0], results[len(results)-1]
	growth := last.queries - first.queries
	eventGrowth := int64(last.events - first.events)

	// N+1 signature: one extra query per extra row. Anything at or above half
	// that slope is the same defect, just partially batched.
	scales := eventGrowth > 0 && growth*2 >= eventGrowth

	if scales {
		fmt.Println("  " + strings.Repeat("!", 54))
		fmt.Printf("  N+1 — query ikut tumbuh bersama jumlah baris\n")
		fmt.Println("  " + strings.Repeat("!", 54))
		fmt.Printf("      %d event -> %d query\n", first.events, first.queries)
		fmt.Printf("      %d event -> %d query\n", last.events, last.queries)
		fmt.Printf("      +%d event menambah +%d query\n", eventGrowth, growth)
		fmt.Println()
		fmt.Println("      Satu request seharusnya berharga tetap, berapa pun barisnya.")
		fmt.Println()
		return 1
	}

	fmt.Println("  " + strings.Repeat("-", 54))
	fmt.Printf("  AMAN — query tetap %d meski event naik dari %d ke %d\n",
		last.queries, first.events, last.events)
	fmt.Println("  " + strings.Repeat("-", 54))
	fmt.Println()
	return 0
}

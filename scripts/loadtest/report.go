package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
)

// ticketCodeKeys are the field names a checkout implementation is likely to
// use for the ticket it allocated. The tool checks all of them so it does not
// depend on a contract that has not been designed yet.
var ticketCodeKeys = []string{"ticket_code", "code", "ticket", "seat", "seat_code"}

// ticketCode digs a ticket identifier out of a response body, wherever it sits.
// Returns "" when the response carries no recognisable code, which simply
// disables duplicate detection for that request.
func ticketCode(payload []byte) string {
	var root any
	if err := json.Unmarshal(payload, &root); err != nil {
		return ""
	}
	return findCode(root)
}

func findCode(node any) string {
	switch v := node.(type) {
	case map[string]any:
		for _, key := range ticketCodeKeys {
			if raw, ok := v[key]; ok {
				if s, ok := raw.(string); ok && s != "" {
					return s
				}
			}
		}
		// Recurse: the code is usually nested under "data".
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			if s := findCode(v[k]); s != "" {
				return s
			}
		}
	case []any:
		for _, item := range v {
			if s := findCode(item); s != "" {
				return s
			}
		}
	}
	return ""
}

func printHeader(cfg config, before eventState) {
	fmt.Println()
	fmt.Println(strings.Repeat("=", 56))
	fmt.Printf("  FLASH SALE LOAD TEST\n")
	fmt.Println(strings.Repeat("=", 56))
	fmt.Printf("  Event      : %s (id %d)\n", before.Name, cfg.eventID)
	fmt.Printf("  Inventory  : %d tiket, %d tersedia\n", before.Total, before.Available)
	fmt.Printf("  Serangan   : %d request, %d bersamaan\n", cfg.requests, cfg.concurrency)
	fmt.Printf("  Target     : %s %s\n", cfg.method, strings.ReplaceAll(cfg.pathTmpl, "{id}", fmt.Sprint(cfg.eventID)))
	fmt.Println(strings.Repeat("=", 56))
	fmt.Println()
}

// report prints the outcome and returns the process exit code:
// 0 when the sale held, 1 when it oversold, 3 when nothing was sold at all.
func report(cfg config, before, after eventState, res *results) int {
	sold := res.succeeded()
	consumed := before.Available - after.Available
	dupes := res.duplicates()

	fmt.Println("  HASIL")
	fmt.Println("  " + strings.Repeat("-", 54))
	fmt.Printf("  Durasi                     : %s\n", res.elapsed.Round(time.Millisecond))
	fmt.Printf("  Order sukses (2xx)         : %d\n", sold)
	fmt.Printf("  Tersedia sebelum → sesudah : %d → %d  (terpakai %d)\n", before.Available, after.Available, consumed)

	if len(res.statuses) > 0 {
		fmt.Printf("  Status HTTP                : %s\n", formatStatuses(res.statuses))
	}
	if len(res.failures) > 0 {
		fmt.Printf("  Gagal di transport         : %d jenis\n", len(res.failures))
		for msg, count := range res.failures {
			fmt.Printf("      %3dx %s\n", count, truncate(msg, 60))
		}
	}
	fmt.Println()

	// Nothing sold: usually the endpoint does not exist yet.
	if sold == 0 {
		fmt.Println("  " + strings.Repeat("-", 54))
		fmt.Println("  BELUM ADA YANG TERJUAL")
		fmt.Println()
		fmt.Println("  Kalau statusnya 404, endpoint checkout memang belum dibuat —")
		fmt.Println("  itu kondisi awal yang diharapkan sebelum story dikerjakan.")
		fmt.Println("  Kalau 401, jalankan ulang dengan -token atau cek user seed.")
		fmt.Println("  Sesuaikan -path kalau kontraknya berbeda.")
		fmt.Println()
		return 3
	}

	oversold := sold > before.Total || consumed > int64(before.Total) || len(dupes) > 0 || after.Available < 0

	if oversold {
		fmt.Println("  " + strings.Repeat("!", 54))
		fmt.Printf("  OVERSELL — terjual %d dari %d tiket\n", sold, before.Total)
		fmt.Println("  " + strings.Repeat("!", 54))
		if sold > before.Total {
			fmt.Printf("      %d order lebih banyak dari tiket yang ada\n", sold-before.Total)
		}
		if after.Available < 0 {
			fmt.Printf("      sisa tiket negatif: %d\n", after.Available)
		}
		if len(dupes) > 0 {
			fmt.Printf("      %d tiket terjual ke lebih dari satu pembeli", len(dupes))
			worst := worstDuplicates(dupes, maxDuplicatesShown)
			fmt.Printf(", terparah:\n")
			for _, d := range worst {
				fmt.Printf("          %s -> %d pembeli\n", d.code, d.buyers)
			}
			if len(dupes) > len(worst) {
				fmt.Printf("          (+%d tiket lain)\n", len(dupes)-len(worst))
			}
		}
		fmt.Println()
		return 1
	}

	fmt.Println("  " + strings.Repeat("-", 54))
	fmt.Printf("  AMAN — terjual %d dari %d tiket, tidak ada yang dobel\n", sold, before.Total)
	fmt.Println("  " + strings.Repeat("-", 54))
	fmt.Println()
	return 0
}

func formatStatuses(statuses map[int]int) string {
	codes := make([]int, 0, len(statuses))
	for code := range statuses {
		codes = append(codes, code)
	}
	sort.Ints(codes)

	parts := make([]string, 0, len(codes))
	for _, code := range codes {
		parts = append(parts, fmt.Sprintf("%d→%d×", code, statuses[code]))
	}
	return strings.Join(parts, "  ")
}

// maxDuplicatesShown keeps the verdict readable on a projector; the count in
// the line above is the real figure.
const maxDuplicatesShown = 5

type duplicate struct {
	code   string
	buyers int
}

// worstDuplicates returns the most-oversold codes first, capped at limit.
func worstDuplicates(dupes map[string]int, limit int) []duplicate {
	out := make([]duplicate, 0, len(dupes))
	for code, buyers := range dupes {
		out = append(out, duplicate{code: code, buyers: buyers})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].buyers != out[j].buyers {
			return out[i].buyers > out[j].buyers
		}
		return out[i].code < out[j].code
	})
	if len(out) > limit {
		out = out[:limit]
	}
	return out
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}

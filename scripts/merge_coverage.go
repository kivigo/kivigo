package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
)

// merge multiple Go coverprofile files into one coverprofile.
// Basic algorithm:
// - ensure all files use the same "mode: <mode>"
// - treat each coverage line "<file>:<start>.<col>,<end>.<col> <numStmts> <count>"
// - key = "<file>:<start>.<col>,<end>.<col> <numStmts>"
// - sum counts for identical keys across files
func main() { //nolint:cyclop
	outPath := flag.String("o", "", "output merged coverprofile")
	flag.Parse()
	inputs := flag.Args()

	if *outPath == "" || len(inputs) == 0 {
		fmt.Fprintf(os.Stderr, "usage: %s -o merged.out inp1.out inp2.out ...\n", os.Args[0])
		os.Exit(2)
	}

	mode := ""
	counts := make(map[string]int64)
	stmts := make(map[string]string) // key -> numStmts (string)

	for _, p := range inputs {
		f, err := os.Open(p)
		if err != nil {
			log.Fatalf("open %s: %v", p, err)
		}
		sc := bufio.NewScanner(f)
		lineNo := 0
		for sc.Scan() {
			line := sc.Text()
			lineNo++
			if strings.HasPrefix(line, "mode:") {
				m := strings.TrimSpace(strings.TrimPrefix(line, "mode:"))
				if mode == "" {
					mode = m
				} else if mode != m {
					log.Fatalf("coverage mode mismatch: %s vs %s (file %s)", mode, m, p)
				}
				continue
			}
			if strings.TrimSpace(line) == "" {
				continue
			}
			parts := strings.Fields(line)
			// Expect 3 parts: pos numStmts count
			if len(parts) != 3 {
				log.Fatalf("unexpected coverage line (%s:%d): %q", p, lineNo, line)
			}
			pos := parts[0]
			num := parts[1]
			cntStr := parts[2]
			cnt, err := strconv.ParseInt(cntStr, 10, 64)
			if err != nil {
				log.Fatalf("invalid count in %s:%d: %v", p, lineNo, err)
			}
			key := pos + " " + num
			counts[key] += cnt
			stmts[key] = num
		}
		if err := sc.Err(); err != nil {
			log.Fatalf("scan %s: %v", p, err)
		}
		_ = f.Close()
	}

	// write output
	out, err := os.Create(*outPath)
	if err != nil {
		log.Fatalf("create %s: %v", *outPath, err)
	}
	defer out.Close()

	if mode == "" {
		log.Fatalf("no mode found in inputs") //nolint:gocritic
	}
	fmt.Fprintf(out, "mode: %s\n", mode)

	// deterministic ordering
	keys := make([]string, 0, len(counts))
	for k := range counts {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		// k = pos + " " + numStmts
		parts := strings.SplitN(k, " ", 2)
		pos := parts[0]
		num := stmts[k]
		fmt.Fprintf(out, "%s %s %d\n", pos, num, counts[k])
	}
}

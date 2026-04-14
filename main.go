package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
)

var (
	ignoreCommon = flag.Bool("i", false, "Ignore common env vars")
	jsonOut      = flag.Bool("j", false, "JSON output")
)

func main() {
	flag.Parse()

	var files []string
	for _, f := range flag.Args() {
		data, err := os.ReadFile(f)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Error reading %s: %v\n", f, err)
			continue
		}
		lines := strings.Split(string(data), "\n")
		envs := make(map[string]string)
		for _, line := range lines {
			if idx := strings.Index(line, "="); idx > 0 {
				key := line[:idx]
				val := line[idx+1:]
				envs[key] = val
			}
		}
		files = append(files, f)
	}

	args := flag.Args()
	if len(args) < 2 {
		fmt.Println("EnvDiff - Compare environment files")
		fmt.Println("Usage: envdiff file1.env file2.env")
		return
	}

	map1 := parseEnvFile(args[0])
	map2 := parseEnvFile(args[1])

	allKeys := mergeKeys(map1, map2)

	var added, removed, changed []string

	for _, key := range allKeys {
		if _, in1 := map1[key]; !in1 {
			added = append(added, key)
		} else if _, in2 := map2[key]; !in2 {
			removed = append(removed, key)
		} else if map1[key] != map2[key] {
			changed = append(changed, key)
		}
	}

	if *jsonOut {
		fmt.Printf(`{"added":%d,"removed":%d,"changed":%d}`, len(added), len(removed), len(changed))
	} else {
		fmt.Println("📊 Environment Diff")
		fmt.Println("====================")
		if len(added) > 0 {
			fmt.Printf("\n🟢 Added (%d):\n", len(added))
			for _, k := range added {
				fmt.Printf("  + %s\n", k)
			}
		}
		if len(removed) > 0 {
			fmt.Printf("\n🔴 Removed (%d):\n", len(removed))
			for _, k := range removed {
				fmt.Printf("  - %s\n", k)
			}
		}
		if len(changed) > 0 {
			fmt.Printf("\n🟡 Changed (%d):\n", len(changed))
			for _, k := range changed {
				fmt.Printf("  ~ %s\n", k)
			}
		}
		if len(added)+len(removed)+len(changed) == 0 {
			fmt.Println("\n✅ No differences found")
		}
	}
}

func parseEnvFile(filename string) map[string]string {
	result := make(map[string]string)
	f, err := os.Open(filename)
	if err != nil {
		return result
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if idx := strings.Index(line, "="); idx > 0 {
			key := strings.TrimSpace(line[:idx])
			val := strings.TrimSpace(line[idx+1:])
			result[key] = val
		}
	}
	return result
}

func mergeKeys(m1, m2 map[string]string) []string {
	keys := make(map[string]bool)
	for k := range m1 {
		keys[k] = true
	}
	for k := range m2 {
		keys[k] = true
	}
	var result []string
	for k := range keys {
		result = append(result, k)
	}
	sort.Strings(result)
	return result
}

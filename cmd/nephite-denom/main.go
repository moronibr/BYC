package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

var denominations = map[string]int{
	"leah":    1,
	"shiblum": 2,
	"shiblon": 4,
	"senum":   8,
	"senine":  8,
	"antion":  12,
	"amnor":   16,
	"seon":    16,
	"ezrom":   32,
	"shum":    32,
	"onti":    56,
	"limnah":  56,
}

var preferredOrder = map[string][]string{
	"gold":   {"limnah", "shum", "seon", "antion", "senine", "shiblon", "shiblum", "leah"},
	"silver": {"onti", "ezrom", "amnor", "senum", "shiblon", "shiblum", "leah"},
}

func normalizeUnit(name string) (string, error) {
	unit := strings.ToLower(strings.TrimSpace(name))
	if _, ok := denominations[unit]; !ok {
		return "", fmt.Errorf("unrecognized denomination: %s", name)
	}
	return unit, nil
}

func toLeah(amount int, unit string) (int, error) {
	if amount < 0 {
		return 0, errors.New("amount must be non-negative")
	}
	normalized, err := normalizeUnit(unit)
	if err != nil {
		return 0, err
	}
	return amount * denominations[normalized], nil
}

type breakdownItem struct {
	denomination string
	quantity     int
}

func fromLeah(value int, preference string) ([]breakdownItem, error) {
	if value < 0 {
		return nil, errors.New("value must be non-negative")
	}
	pref := strings.ToLower(preference)
	order, ok := preferredOrder[pref]
	if !ok {
		return nil, fmt.Errorf("preference must be one of %v", sortedKeys(preferredOrder))
	}
	remainder := value
	breakdown := make([]breakdownItem, 0, len(order))
	for _, denom := range order {
		denomValue := denominations[denom]
		qty := remainder / denomValue
		if qty > 0 {
			breakdown = append(breakdown, breakdownItem{
				denomination: denom,
				quantity:     qty,
			})
			remainder -= qty * denomValue
		}
	}
	if remainder != 0 {
		breakdown = append(breakdown, breakdownItem{
			denomination: "leah",
			quantity:     remainder,
		})
	}
	return breakdown, nil
}

func formatBreakdown(items []breakdownItem) string {
	if len(items) == 0 {
		return "0 leah"
	}
	parts := make([]string, 0, len(items))
	for _, item := range items {
		name := item.denomination
		if item.quantity != 1 {
			name += "s"
		}
		parts = append(parts, fmt.Sprintf("%d %s", item.quantity, name))
	}
	if len(parts) == 1 {
		return parts[0]
	}
	return strings.Join(parts[:len(parts)-1], ", ") + ", and " + parts[len(parts)-1]
}

func sortedKeys[T any](m map[string]T) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "  nephite-denom to-leah <amount> <unit>\n")
	fmt.Fprintf(os.Stderr, "  nephite-denom breakdown <value> [--preference gold|silver]\n")
	fmt.Fprintf(os.Stderr, "\nAvailable denominations: %s\n", strings.Join(sortedKeys(denominations), ", "))
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "to-leah":
		toLeahCmd := flag.NewFlagSet("to-leah", flag.ExitOnError)
		if err := toLeahCmd.Parse(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if toLeahCmd.NArg() != 2 {
			fmt.Fprintln(os.Stderr, "to-leah requires <amount> and <unit>")
			os.Exit(1)
		}
		amount, err := strconv.Atoi(toLeahCmd.Arg(0))
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid amount: %v\n", err)
			os.Exit(1)
		}
		unit := toLeahCmd.Arg(1)
		result, err := toLeah(amount, unit)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Println(result)

	case "breakdown":
		breakdownCmd := flag.NewFlagSet("breakdown", flag.ExitOnError)
		preference := breakdownCmd.String("preference", "silver", "prefer gold or silver series when formatting")
		if err := breakdownCmd.Parse(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if breakdownCmd.NArg() != 1 {
			fmt.Fprintln(os.Stderr, "breakdown requires <value>")
			os.Exit(1)
		}
		value, err := strconv.Atoi(breakdownCmd.Arg(0))
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid value: %v\n", err)
			os.Exit(1)
		}
		items, err := fromLeah(value, *preference)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Println(formatBreakdown(items))

	case "-h", "--help", "help":
		usage()

	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", os.Args[1])
		usage()
		os.Exit(1)
	}
}

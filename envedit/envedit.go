package main

import (
	"fmt"
	"strings"

	"golang.org/x/sys/windows/registry"
)

func main() {
	regEnv, err := registry.OpenKey(registry.CURRENT_USER, `Environment`, registry.ALL_ACCESS)
	if err != nil {
		fmt.Printf("Error: Failed to get user env, details: %v\n", err)
		return
	}
	defer regEnv.Close()

	var content strings.Builder
	names, err := regEnv.ReadValueNames(0)
	if err != nil {
		fmt.Printf("Error: Failed to READ env names, details: %v\n", err)
		return
	}

	content.WriteString(`
# Format: Key=Value
# Accepts Duplicate Keys
# - Duplicate Keys will be joined into one line with a semicolon

	`)

	oldEnv := make(map[string]string)

	for _, name := range names {
		val, _, err := regEnv.GetStringValue(name)
		if err != nil {
			fmt.Printf("Name: %s (Failed to read as string)\n", name)
			continue
		}
		oldEnv[name] = val
		if strings.Contains(val, ";") {
			for v := range strings.SplitSeq(val, ";") {
				fmt.Fprintf(&content, "%s=%s\n", name, strings.TrimSpace(v))
			}
		} else {
			fmt.Fprintf(&content, "%s=%s\n", name, strings.TrimSpace(val))
		}
	}

	res, saved := EditText(content.String())

	if !saved {
		fmt.Println("No changes made")
		return
	}

	newEnv := make(map[string]string)

	for line := range strings.SplitSeq(res, "\n") {
		if strings.HasPrefix(line, "#") {
			continue
		}
		if strings.Contains(line, "=") {
			kv := strings.Split(line, "=")
			if len(kv) == 2 {
				k, v := strings.TrimSpace(kv[0]), strings.TrimSpace(kv[1])
				if ev, ok := newEnv[k]; ok {
					newEnv[k] = ev + ";" + v
				} else {
					newEnv[k] = v
				}
			}
		}
	}

	const UPDATED = "___UPDATED___"

	for k, v := range newEnv {

		fmt.Printf("SET (%s=%s)\n", k, v)

		err = regEnv.SetExpandStringValue(k, v)
		if err != nil {
			fmt.Printf("Name: %s (Failed to set as string, details: %v)\n", k, err)
			continue
		}
		oldEnv[k] = UPDATED
	}

	for k, v := range oldEnv {

		if v != UPDATED {
			fmt.Printf("DEL (%s)\n", k)
			err = regEnv.DeleteValue(k)
			if err != nil {
				fmt.Printf("Name: %s (Failed to delete, details: %v)\n", k, err)
				continue
			}
		}
	}

	fmt.Println("Done")

}

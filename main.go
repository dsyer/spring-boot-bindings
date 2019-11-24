package main

import (
	"flag"
	"strings"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

func main() {
	output := flag.String("f", "-", "output target (or - for stdout)")
	flag.Parse()
	path := getBindingsPath(flag.Args())
	fmt.Fprintf(os.Stderr, "Reading from: %s\n", path)
	result := getProperties(path)
	target := *output
	if target != "-" {
		dir := filepath.Dir(target)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			if err = os.MkdirAll(dir, 0755); err != nil {
				fmt.Fprintf(os.Stderr, "Cannot create directory: %s\n", dir)
				os.Exit(1)
			}
			fmt.Fprintf(os.Stderr, "Created directory: %s\n", dir)
		}
		fmt.Fprintf(os.Stderr, "Writing file: %s\n", target)
		ioutil.WriteFile(target, result, 0644)
	} else {
		fmt.Printf(string(result))
	}
}

func getProperties(path string) []byte {
	properties := []byte{}
	paths, _ := ioutil.ReadDir(path)
	for _, dir := range paths {
		if dir.IsDir() {
			name := dir.Name()
			properties = append(properties, readMetaData(path, name)...)
			properties = append(properties, readSecret(path, name)...)
		}
	}
	return properties
}

func readProperties(base string, prefix string) string {
	paths, _ := ioutil.ReadDir(base)
	result := ""
	for _, file := range paths {
		if !file.IsDir() {
			key := file.Name()
			value, err := ioutil.ReadFile(base + "/" + key)
			if err == nil {
                newline := "\\n"
                if key == "tags" {
                    newline = ","
                }
				prop := prefix + "." + key + "=" + strings.ReplaceAll(string(value), "\n", newline) + "\n"
				result = result + prop
			}
		}
	}
	return result
}

func readMetaData(path string, name string) []byte {
	return []byte(readProperties(path+"/"+name+"/metadata", "cnb.metadata." + name))
}

func readSecret(path string, name string) []byte {
	return []byte(readProperties(path+"/"+name+"/secret", "cnb.secret." + name))
}

func getEnv(name string, defaultValue string) string {
	if value, ok := os.LookupEnv(name); ok {
		return value
	}
	return defaultValue
}

func getBindingsPath(args []string) string {
	if len(args) == 0 {
		return getEnv("CNB_BINDINGS", "/config/bindings")
	}
	return args[0]
}

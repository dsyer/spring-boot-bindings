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
	result := map[string]string{}
	paths, _ := ioutil.ReadDir(path)
	for _, dir := range paths {
		if dir.IsDir() {
			name := dir.Name()
			result = addAll(result, readMetaData(path, name))
            result = addAll(result, readSecret(path, name))
            // TODO: add map entries keyed on the kind: e.g. for mysql add spring.datasource.*
		}
	}
	return properties(result)
}

func properties(values map[string]string) []byte {
    result := []string{}
    for k,v := range values {
        result = append(result, property(k,v))
    }
    return []byte(strings.Join(result, "\n"))
}

func property(key string, value string) string {
    return key + "=" + value
}

func addAll(values map[string]string, added map[string]string) map[string]string {
    for k,v := range added {
        values[k] = v
    }
    return values
}

func readProperties(base string, prefix string) map[string]string {
	paths, _ := ioutil.ReadDir(base)
	result := map[string]string{}
	for _, file := range paths {
		if !file.IsDir() {
			key := file.Name()
			bytes, err := ioutil.ReadFile(base + "/" + key)
			if err == nil {
                newline := "\\n"
                if key == "tags" {
                    newline = ","
                }
                value := string(bytes)
                value = strings.TrimSuffix(value, "\n")
				result[prefix + "." + key] = strings.ReplaceAll(value, "\n", newline) 
			}
		}
	}
	return result
}

func readMetaData(path string, name string) map[string]string {
	return readProperties(path+"/"+name+"/metadata", "cnb.metadata." + name)
}

func readSecret(path string, name string) map[string]string {
	return readProperties(path+"/"+name+"/secret", "cnb.secret." + name)
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

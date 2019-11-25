package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
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

// Binding represents CNB binding
type Binding struct {
	Name     string
	Metadata Metadata
	Secret   map[string]string
}

// Metadata represents CNB binding metadata
type Metadata struct {
	Kind       string
	Tags       []string
	Provider   string
	Additional map[string]string
}

func getProperties(path string) []byte {
	result := map[string]string{}
	paths, _ := ioutil.ReadDir(path)
	for _, dir := range paths {
		if dir.IsDir() {
            name := dir.Name()
            bindings := Binding {
                Name: name,
                Metadata: readMetaData(path, name),
                Secret: readSecret(path, name),
            }
			result = addAll(result, flattenMetadata(bindings.Metadata, name))
			result = addAll(result, flattenSecret(bindings.Secret, name))
			// TODO: add map entries keyed on the kind: e.g. for mysql add spring.datasource.*
		}
	}
	return properties(result)
}

func flattenSecret(secret map[string]string, name string) map[string]string {
    result := map[string]string{}
    for k,v := range secret {
        result["cnb.secret." + name  + "." + k] = v
    }
	return result
}

func flattenMetadata(metadata Metadata, name string) map[string]string {
    result := map[string]string{}
    for k,v := range metadata.Additional {
        result["cnb.metadata." + name + "." + k] = v
    }
    result["cnb.metadata." + name  + ".kind"] = metadata.Kind
    result["cnb.metadata." + name  + ".provider"] = metadata.Provider
    result["cnb.metadata." + name  + ".tags"] = strings.Join(metadata.Tags, ",")
	return result
}

func properties(values map[string]string) []byte {
	result := []string{}
	for k, v := range values {
		result = append(result, property(k, v))
	}
	return []byte(strings.Join(result, "\n"))
}

func property(key string, value string) string {
	return key + "=" + value
}

func addAll(values map[string]string, added map[string]string) map[string]string {
	for k, v := range added {
		values[k] = v
	}
	return values
}

func readProperties(base string) map[string]string {
	paths, _ := ioutil.ReadDir(base)
	result := map[string]string{}
	for _, file := range paths {
		if !file.IsDir() {
			key := file.Name()
			bytes, err := ioutil.ReadFile(base + "/" + key)
			if err == nil {
				value := string(bytes)
				value = strings.TrimSuffix(value, "\n")
				result[key] = strings.ReplaceAll(value, "\n", "\\n")
			}
		}
	}
	return result
}

func readMetaData(path string, name string) Metadata {
	values := readProperties(path + "/" + name + "/metadata")
	tags := []string{}
	if value := values["tags"]; value != "" {
		tags = strings.Split(value, "\\n")
		delete(values, "tags")
	}
	kind := "unknown"
	if value := values["kind"]; value != "" {
		kind = value
		delete(values, "kind")
	}
	provider := "unknown"
	if value := values["provider"]; value != "" {
		provider = value
		delete(values, "provider")
	}
	result := Metadata{
		Kind:       kind,
		Tags:       tags,
		Provider:   provider,
		Additional: values,
	}
	return result
}

func readSecret(path string, name string) map[string]string {
	return readProperties(path + "/" + name + "/secret")
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

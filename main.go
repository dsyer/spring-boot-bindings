package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

func main() {
	output := flag.String("f", "-", "output target (or - for stdout)")
	flag.Parse()
	path := getBindingsPath(flag.Args())
	fmt.Fprintf(os.Stderr, "Reading from: %s\n", path)
	result := getProperties(path, loadTemplates(path+"/../templates"))
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
		ioutil.WriteFile(target, []byte(result), 0644)
	} else {
		fmt.Printf(result)
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

// Templates represents a mapping from a binding kind to a set of templates
type Templates struct {
	Kind     string
	Main     []template.Template
	Optional []template.Template
}

func getProperties(path string, templates map[string]Templates) string {
	result := map[string]string{}
	paths, _ := ioutil.ReadDir(path)
	fragments := []string{}
	for _, dir := range paths {
		if dir.IsDir() {
			name := dir.Name()
			binding := readBinding(path, name)
			result = addAll(result, flattenMetadata(binding.Metadata, name))
			result = addAll(result, flattenSecret(binding.Secret, name))
			rendered, err := render(templates[binding.Metadata.Kind], binding)
			if err == nil {
				fragments = append(fragments, rendered)
			}
		}
	}
	if len(fragments) > 0 {
		fragments = append(fragments, "")
	}
	fragments = append(fragments, properties(result))
	return strings.Join(fragments, "\n")
}

func render(current Templates, binding Binding) (string, error) {
	fragments := []string{}
	for _, t := range current.Main {
		buffer := &bytes.Buffer{}
		err := t.Execute(buffer, binding)
		value := buffer.String()
		if err == nil && !strings.Contains(value, "<no value>") {
			fragments = append(fragments, value)
		} else {
			if err == nil {
				err = errors.New("Cannot render: " + current.Kind)
			}
			return strings.Join(fragments, "\n"), err
		}
	}
	for _, t := range current.Optional {
		buffer := &bytes.Buffer{}
		err := t.Execute(buffer, binding)
		value := buffer.String()
		if err == nil && !strings.Contains(value, "<no value>") {
			fragments = append(fragments, value)
		}
	}
	return strings.Join(fragments, "\n"), nil
}

func loadTemplates(path string) map[string]Templates {
	result := map[string]Templates{}
	paths, _ := ioutil.ReadDir(path)
	for _, dir := range paths {
		if dir.IsDir() {
			kind := dir.Name()
			result[kind] = Templates{
				Kind:     kind,
				Main:     loadMainTemplates(path + "/" + kind),
				Optional: loadOptionalTemplates(path + "/" + kind),
			}
		}
	}
	return result
}

func loadOptionalTemplates(path string) []template.Template {
	result := []template.Template{}
	files, _ := ioutil.ReadDir(path)
	for _, file := range files {
		if file.IsDir() {
			result = append(result, loadMainTemplates(path+"/"+file.Name())...)
		}
	}
	return result
}

func loadMainTemplates(path string) []template.Template {
	result := []template.Template{}
	files, _ := ioutil.ReadDir(path)
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".tmpl") {
			bytes, err := ioutil.ReadFile(path + "/" + file.Name())
			if err == nil {
				value := string(bytes)
				value = strings.TrimSuffix(value, "\n")
				var tmpl *template.Template
				tmpl, err = template.New(file.Name()).Parse(value)
				if err == nil {
					result = append(result, *tmpl)
				}
			}
		}
	}
	return result
}

func readBinding(path string, name string) Binding {
	return Binding{
		Name:     name,
		Metadata: readMetaData(path, name),
		Secret:   readSecret(path, name),
	}
}

func flattenSecret(secret map[string]string, name string) map[string]string {
	result := map[string]string{}
	for k, v := range secret {
		result["cnb.secret."+name+"."+k] = v
	}
	return result
}

func flattenMetadata(metadata Metadata, name string) map[string]string {
	result := map[string]string{}
	for k, v := range metadata.Additional {
		result["cnb.metadata."+name+"."+k] = v
	}
	result["cnb.metadata."+name+".kind"] = metadata.Kind
	result["cnb.metadata."+name+".provider"] = metadata.Provider
	result["cnb.metadata."+name+".tags"] = strings.Join(metadata.Tags, ",")
	return result
}

func properties(values map[string]string) string {
	result := []string{}
	for k, v := range values {
		result = append(result, property(k, v))
	}
	return strings.Join(result, "\n")
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

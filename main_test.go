package main

import (
	"strings"
	"testing"
)

func TestReadMeta(t *testing.T) {
	props := readMetaData("samples/input", "mysql")
	if props["cnb.metadata.mysql.test"] != "Hello\\nWorld" {
		t.Errorf("Props = %s; want 'mysql.test=Hello\\nWorld'", props)
	}
}

func TestReadSecret(t *testing.T) {
	props := readSecret("samples/input", "mysql")
	if props["cnb.secret.mysql.password"] != "secret" {
		t.Errorf("Props = %s; want 'mysql.password=secret'", props)
	}
}

func TestMetaKeyValue(t *testing.T) {
	props := readProperties("samples/input/mysql/metadata", "mysql")
	if props["mysql.test"] != "Hello\\nWorld" {
		t.Errorf("Props = %s; want 'mysql.test=Hello\\nWorld'", props)
	}
}

func TestMetaTags(t *testing.T) {
	props := readProperties("samples/input/mysql/metadata", "mysql")
	if props["mysql.tags"] != "one,two,three" {
		t.Errorf("Props = %s; want 'mysql.tags=one,two,three'", props)
	}
}

func TestProperties(t *testing.T) {
	props := map[string]string{
		"foo": "bar",
		"spam.foo": "bar.foo",
	}
	result := string(properties(props))
	if !contains(strings.Split(result, "\n"), "spam.foo=bar.foo") {
		t.Errorf("Props = %s; want 'spam.foo=bar.foo'", result)
	}
}

func contains(s []string, e string) bool {
    for _, a := range s {
        if a == e {
            return true
        }
    }
    return false
}
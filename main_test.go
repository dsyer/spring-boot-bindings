package main

import (
	"strings"
	"testing"
)

func TestReadMeta(t *testing.T) {
	props := readMetaData("samples/input", "mysql")
	if props.Additional["test"] != "Hello\\nWorld" {
		t.Errorf("Props = %s; want 'test=Hello\\nWorld'", props)
	}
}

func TestReadSecret(t *testing.T) {
	props := readSecret("samples/input", "mysql")
	if props["password"] != "secret" {
		t.Errorf("Props = %s; want 'password=secret'", props)
	}
}

func TestMetaKeyValue(t *testing.T) {
	props := readProperties("samples/input/mysql/metadata")
	if props["test"] != "Hello\\nWorld" {
		t.Errorf("Props = %s; want '.test=Hello\\nWorld'", props)
	}
}

func TestMetaTags(t *testing.T) {
	props := readMetaData("samples/input", "mysql")
	if !containsAll(props.Tags, []string{"one","two","three"}) {
		t.Errorf("Tags = %s; want 'tags=one,two,three'", props.Tags)
	}
}

func TestFlattenMeta(t *testing.T) {
	props := flattenMetadata(readMetaData("samples/input", "mysql"), "mysql")
	if props["cnb.metadata.mysql.test"] != "Hello\\nWorld" {
		t.Errorf("Props = %s; want 'test=Hello\\nWorld'", props)
	}
	if props["cnb.metadata.mysql.tags"] != "one,two,three" {
		t.Errorf("Props = %s; want 'tags=one,two,three'", props)
	}
}

func TestFlattenSecret(t *testing.T) {
	props := flattenSecret(readSecret("samples/input", "mysql"), "mysql")
	if props["cnb.secret.mysql.password"] != "secret" {
		t.Errorf("Props = %s; want 'password=secret'", props)
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

func containsAll(s []string, e []string) bool {
    for _, a := range e {
        if !contains(s, a) {
            return false
        }
    }
    return true
}
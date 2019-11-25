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
	if !containsAll(props.Tags, []string{"one", "two", "three"}) {
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
		"foo":      "bar",
		"spam.foo": "bar.foo",
	}
	result := string(properties(props))
	if !contains(strings.Split(result, "\n"), "spam.foo=bar.foo") {
		t.Errorf("Props = %s; want 'spam.foo=bar.foo'", result)
	}
}

func TestLoadTemplates(t *testing.T) {
	mysql := loadTemplates("samples/templates").Lookup("missing")
	if len(mysql.Templates()) != 3 {
		t.Errorf("Wrong templates for mysql (expected 3, found %d)", len(mysql.Templates()))
	}
}

func TestMainTemplate(t *testing.T) {
	mysql := loadTemplates("templates").Lookup("mysql")
	buffer := render(mysql, Binding {
		Name: "mysql",
		Metadata: Metadata{
			Additional: map[string]string {
				"host": "mysql",
			}, 
		},
		Secret: map[string]string {
			"database": "test",
			"user": "test",
			"password": "test",
		},
	})
	if !strings.Contains(buffer, "spring.datasource.url=jdbc:mysql://mysql/test") {
		t.Errorf("Wrong result: %s", buffer)
	}
	if !strings.Contains(buffer, "spring.datasource.password") {
		t.Errorf("Wrong result: %s", buffer)
	}
	if !strings.Contains(buffer, "spring.datasource.user") {
		t.Errorf("Wrong result: %s", buffer)
	}
}

func TestMissingTemplates(t *testing.T) {
	mysql := loadTemplates("samples/templates").Lookup("missing")
	buffer := render(mysql, Binding {
		Name: "mysql",
		Metadata: Metadata{
			Additional: map[string]string {
				"host": "mysql",
				"spam": "spam",
			}, 
		},
		Secret: map[string]string {
			"database": "test",
			"user": "test",
			"password": "test",
		},
	})
	if !strings.Contains(buffer, "spring.datasource.url=jdbc:mysql://mysql/test") {
		t.Errorf("Wrong result: %s", buffer)
	}
	if !strings.Contains(buffer, "spring.datasource.password") {
		t.Errorf("Wrong result: %s", buffer)
	}
	if !strings.Contains(buffer, "spring.datasource.user") {
		t.Errorf("Wrong result: %s", buffer)
	}
	if !strings.Contains(buffer, "spam=spam") {
		t.Errorf("Wrong result: %s", buffer)
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

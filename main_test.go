package main

import (
	"strings"
	"testing"
)

func TestReadMeta(t *testing.T) {
	props := string(readMetaData("samples/input", "mysql"))
	if !strings.Contains(props, "cnb.metadata.mysql.test=Hello\\nWorld") {
		t.Errorf("Props = %s; want 'mysql.test=Hello\\nWorld'", props)
	}
}

func TestReadSecret(t *testing.T) {
	props := string(readSecret("samples/input", "mysql"))
	if !strings.Contains(props, "cnb.secret.mysql.password=secret") {
		t.Errorf("Props = %s; want 'mysql.password=secret'", props)
	}
}

func TestMetaKeyValue(t *testing.T) {
	props := readProperties("samples/input/mysql/metadata", "mysql")
	if !contains(strings.Split(props, "\n"), "mysql.test=Hello\\nWorld") {
		t.Errorf("Props = %s; want 'mysql.test=Hello\\nWorld'", props)
	}
}

func TestMetaTags(t *testing.T) {
	props := readProperties("samples/input/mysql/metadata", "mysql")
	if !contains(strings.Split(props, "\n"), "mysql.tags=one,two,three") {
		t.Errorf("Props = %s; want 'mysql.tags=one,two,three'", props)
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
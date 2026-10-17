// Copyright 2026 3eAI Labs Ltd
// SPDX-License-Identifier: Apache-2.0

package sbiclient

import (
	"encoding/json"
	"flag"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"
)

var update = flag.Bool("update", false, "rewrite testdata/schema.golden")

// TestWireSchema pins the JSON shape of every envelope: field names, nesting and JSON
// types. A diff here is a wire-contract change — renaming a field or changing its type
// breaks every consumer, including those in other languages. Regenerate with
// `go test -run TestWireSchema -update` only for an intended, versioned change.
func TestWireSchema(t *testing.T) {
	var lines []string
	for _, root := range []any{
		AuthRequest{}, AuthenticationInfoResult{}, AuthConfirmRequest{}, AuthConfirmReply{},
		MobilityRequest{}, MobilityReply{}, MobilityQuery{},
		Amf3GppAccessRegistration{}, SmfRegistration{}, SmsfRegistration{},
		LocationInfoRequest{}, ProvideLocInfoResult{},
		DataRequest{}, DataResponse{}, NotificationPayload{},
		ErrorReply{},
		NudmUecmMessage{}, SbiOtaMessage{}, DeliveryEvent{},
	} {
		typ := reflect.TypeOf(root)
		walk(typ.Name(), typ, map[reflect.Type]bool{}, &lines)
	}
	got := strings.Join(lines, "\n") + "\n"

	const golden = "testdata/schema.golden"
	if *update {
		if err := os.WriteFile(golden, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatalf("%v (run with -update to create it)", err)
	}
	if got != string(want) {
		t.Errorf("wire schema changed; diff against %s:\n%s", golden, diffLines(string(want), got))
	}
}

// walk appends one "path json-type" line per JSON field, sorted per struct.
func walk(path string, typ reflect.Type, seen map[reflect.Type]bool, out *[]string) {
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	if typ.Kind() != reflect.Struct || seen[typ] {
		return
	}
	seen[typ] = true
	defer delete(seen, typ)

	fields := jsonFields(typ)
	sort.Slice(fields, func(i, j int) bool { return jsonName(fields[i]) < jsonName(fields[j]) })

	for _, f := range fields {
		p := path + "." + jsonName(f)
		*out = append(*out, p+" "+jsonType(f.Type))
		elem := f.Type
		for elem.Kind() == reflect.Pointer || elem.Kind() == reflect.Slice {
			if elem == rawMessage || elem.Kind() == reflect.Slice && elem.Elem().Kind() == reflect.Uint8 {
				break // []byte is a base64 string, not an array
			}
			elem = elem.Elem()
		}
		if elem.Kind() == reflect.Struct && elem.PkgPath() == typ.PkgPath() {
			walk(p, elem, seen, out)
		}
	}
}

// jsonFields lists the fields encoding/json writes for typ: an embedded struct without a
// JSON name contributes its own fields at the same level.
func jsonFields(typ reflect.Type) []reflect.StructField {
	var fields []reflect.StructField
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		if f.Anonymous && f.Type.Kind() == reflect.Struct && f.Tag.Get("json") == "" {
			fields = append(fields, jsonFields(f.Type)...)
			continue
		}
		if f.IsExported() && jsonName(f) != "-" {
			fields = append(fields, f)
		}
	}
	return fields
}

func jsonName(f reflect.StructField) string {
	name, _, _ := strings.Cut(f.Tag.Get("json"), ",")
	if name == "" {
		return f.Name
	}
	return name
}

var rawMessage = reflect.TypeOf(json.RawMessage(nil))

func jsonType(t reflect.Type) string {
	if t == rawMessage {
		return "json" // any JSON value, passed through unchanged
	}
	switch t.Kind() {
	case reflect.Pointer:
		return jsonType(t.Elem()) + "?"
	case reflect.Slice:
		if t.Elem().Kind() == reflect.Uint8 {
			return "base64"
		}
		return "[]" + jsonType(t.Elem())
	case reflect.Map:
		return "map[" + jsonType(t.Key()) + "]" + jsonType(t.Elem())
	case reflect.Struct:
		if t.PkgPath() == "time" && t.Name() == "Time" {
			return "rfc3339"
		}
		return "object"
	case reflect.String:
		return "string"
	case reflect.Bool:
		return "bool"
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int, reflect.Uint:
		return t.Kind().String()
	default:
		return t.Kind().String()
	}
}

func diffLines(want, got string) string {
	w := map[string]bool{}
	for _, l := range strings.Split(want, "\n") {
		w[l] = true
	}
	g := map[string]bool{}
	var b strings.Builder
	for _, l := range strings.Split(got, "\n") {
		g[l] = true
		if !w[l] {
			b.WriteString("+ " + l + "\n")
		}
	}
	for _, l := range strings.Split(want, "\n") {
		if !g[l] {
			b.WriteString("- " + l + "\n")
		}
	}
	return b.String()
}

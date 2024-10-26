package main

import (
	"bytes"
	"encoding/json"
	"log"
	"maps"
	"net/http"
	"os"
	"strings"
)

type schema struct {
	Type       string         `json:"type"`
	AllOf      []ref          `json:"allOf"`
	Properties map[string]any `json:"properties"`
}

type ref struct {
	Ref string `json:"$ref"`
}

func main() {
	args := os.Args

	if len(args) < 3 || len(args) > 3 {
		log.Fatal("No arguments")
	}

	body, err := os.ReadFile(args[1])
	if err != nil {
		panic(err)
	}
	var s schema
	if err := json.Unmarshal(body, &s); err != nil {
		panic(err)
	}
	if err := includeRefs(s.AllOf, s); err != nil {
		panic(err)
	}
	// all refs have been recursively included so we can remove the root refs.
	s.AllOf = nil

	s1Bytes, err := json.Marshal(s)
	if err != nil {
		panic(err)
	}

	if err := os.WriteFile(args[2], s1Bytes, 0644); err != nil {
		panic(err)
	}
}

func includeRefs(refs []ref, s schema) error {
	for _, ref := range refs {
		var refBuff bytes.Buffer
		if strings.HasPrefix(ref.Ref, "http") {
			resp, err := http.Get(ref.Ref)
			if err != nil {
				return err
			}
			defer resp.Body.Close()
			_, err = refBuff.ReadFrom(resp.Body)
			if err != nil {
				return err
			}
		} else {
			f, err := os.Open(ref.Ref)
			if err != nil {
				return err
			}
			defer f.Close()
			_, err = refBuff.ReadFrom(f)
			if err != nil {
				return err
			}
		}

		var temp schema
		if err := json.Unmarshal(refBuff.Bytes(), &temp); err != nil {
			return err
		}
		if temp.AllOf != nil {
			includeRefs(temp.AllOf, temp)
		}
		maps.Copy(s.Properties, temp.Properties)
	}
	return nil
}

package main

import (
	"encoding/json"
	"log"
	"maps"
	"os"
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
		refBytes, err := os.ReadFile(ref.Ref)
		if err != nil {
			return err
		}
		var temp schema
		if err := json.Unmarshal(refBytes, &temp); err != nil {
			return err
		}
		if temp.AllOf != nil {
			includeRefs(temp.AllOf, temp)
		}
		maps.Copy(s.Properties, temp.Properties)
	}
	return nil
}

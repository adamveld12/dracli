package main

import (
	"encoding/json"
	"fmt"
	"os"
)

//Credential represents credentials for the credentials json
type Credential struct {
	Host      string
	Username  string
	AuthToken string
}

//LoadCredentials from a credentials json file
func LoadCredentials(path string) (*Credential, error) {
	fp := fmt.Sprintf("%s/credentials.json", path)
	f, err := os.Open(fp)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	credential := &Credential{}
	dec := json.NewDecoder(f)
	if err := dec.Decode(&credential); err != nil {
		return nil, err
	}

	return credential, nil
}

//SaveCredentials saves credentials to a json file
func SaveCredentials(path string, c Credential) error {
	f, err := os.Create(fmt.Sprintf("%s/credentials.json", path))
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "\t")
	if err := enc.Encode(c); err != nil {
		return err
	}
	return nil
}

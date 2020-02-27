package main

import (
	"io/ioutil"
	"log"
	"os"
	"path"

	yml "gopkg.in/yaml.v2"
)

type config struct {
	Debug bool
	Token string
}

var conf config

func readConf(f string) {
	if !path.IsAbs(f) && f[:1] == "~" {
		f = path.Join(os.Getenv("HOME"), f[1:])
	}

	bytes, err := ioutil.ReadFile(f)
	if err != nil {
		log.Fatalf("Read file %s is failed: %s", f, err)
	}

	err = yml.Unmarshal(bytes, &conf)
	if err != nil {
		log.Fatalf("Read config is failed: %s", err)
	}
}

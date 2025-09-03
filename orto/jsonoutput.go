package orto

import (
	"encoding/json"
	"log"
	"os"
)

type JsonOutput struct {
	absDestinationChangeSetJsonFile string
	encoder                         *json.Encoder
	file                            *os.File
	needsAComma                     bool
}

func (jsonOut *JsonOutput) start() {
	f, err := os.Create(jsonOut.absDestinationChangeSetJsonFile)
	if err != nil {
		log.Fatal(err)
	}
	jsonOut.file = f
	jsonOut.encoder = json.NewEncoder(jsonOut.file)
	jsonOut.writeRaw("{\"deletions\":[\n")
}

func (jsonOut *JsonOutput) encode(s any) {
	err := jsonOut.encoder.Encode(s)
	if err != nil {
		log.Fatal(err)
	}
}

func (jsonOut *JsonOutput) writeRaw(s string) {
	_, err := jsonOut.file.Write([]byte(s))
	if err != nil {
		log.Fatal(err)
	}
}
func (jsonOut *JsonOutput) close() {
	jsonOut.writeRaw("]}\n")
	jsonOut.file.Close()
}

// TODO: make this better:
func (jsonOut *JsonOutput) maybeAddComma() {
	if jsonOut.needsAComma {
		jsonOut.writeRaw(",\n")
	} else {
		jsonOut.needsAComma = true
	}
}

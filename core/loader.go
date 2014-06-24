package core

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"path"

	"github.com/idealeak/goserver/core/logger"
)

type Package interface {
	Name() string
	Init() error
	io.Closer
}

var packages = make(map[string]Package)

func RegistePackage(p Package) {
	packages[p.Name()] = p
}

func LoadPackages(configFile string) {
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		logger.Logger.Errorf("Error while reading config file %s: %s", configFile, err)
	}

	switch path.Ext(configFile) {
	case ".json":
		// Compact JSON to make it easier to extract JSON per package
		var buf bytes.Buffer
		err = json.Compact(&buf, data)
		if err != nil {
			logger.Logger.Errorf("Error in JSON config file %s: %s", configFile, err)
		}
		data = buf.Bytes()

		// Unmarshal packages /*in given order*/
		for _, pkg := range packages {
			// Extract JSON only for this package
			key := []byte(`"` + pkg.Name() + `":{`)
			begin := bytes.Index(data, key)
			if begin != -1 {
				begin += len(key) - 1
				end := 0
				braceCounter := 0
				for i := begin; i < len(data); i++ {
					switch data[i] {
					case '{':
						braceCounter++
					case '}':
						braceCounter--
					}
					if braceCounter == 0 {
						end = i + 1
						break
					}
				}

				err = json.Unmarshal(data[begin:end], pkg)
				if err != nil {
					logger.Logger.Errorf("Error while unmarshalling JSON from config file %s: %s", configFile, err)
				}
			}
			err := pkg.Init()
			if err != nil {
				logger.Logger.Errorf("Error while initializing package %s: %s", pkg.Name(), err)
			} else {
				logger.Logger.Info("module [", pkg.Name(), "] load success")
			}
		}

	default:
		panic("Unsupported config file: " + configFile)
	}
}

func ClosePackages() {
	for _, pkg := range packages {
		err := pkg.Close()
		if err != nil {
			logger.Logger.Errorf("Error while closing package %s: %s", pkg.Name(), err)
		}
	}
}

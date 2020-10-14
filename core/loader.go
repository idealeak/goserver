package core

import (
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
type ConfigFileEncryptorHook interface {
	IsCipherText([]byte) bool
	Encrypt([]byte) []byte
	Decrtypt([]byte) []byte
}

var packages = make(map[string]Package)
var packagesLoaded = make(map[string]bool)
var configFileEH ConfigFileEncryptorHook

func RegistePackage(p Package) {
	packages[p.Name()] = p
}

func IsPackageRegisted(name string) bool {
	if _, exist := packages[name]; exist {
		return true
	}
	return false
}

func IsPackageLoaded(name string) bool {
	if _, exist := packagesLoaded[name]; exist {
		return true
	}
	return false
}
func RegisterConfigEncryptor(h ConfigFileEncryptorHook) {
	configFileEH = h
}
func LoadPackages(configFile string) {
	//logger.Logger.Infof("Build time is: %s", BuildTime())
	switch path.Ext(configFile) {
	case ".json":
		fileBuff, err := ioutil.ReadFile(configFile)
		if err != nil {
			logger.Logger.Errorf("Error while reading config file %s: %s", configFile, err)
			break
		}
		if configFileEH != nil {
			if configFileEH.IsCipherText(fileBuff) {
				fileBuff = configFileEH.Decrtypt(fileBuff)
			}
		}
		var fileData interface{}
		err = json.Unmarshal(fileBuff, &fileData)
		if err != nil {
			break
		}
		fileMap := fileData.(map[string]interface{})
		for name, pkg := range packages {
			if moduleData, ok := fileMap[name]; ok {
				if data, ok := moduleData.(map[string]interface{}); ok {
					modelBuff, _ := json.Marshal(data)
					err = json.Unmarshal(modelBuff, &pkg)
					if err != nil {
						logger.Logger.Errorf("Error while unmarshalling JSON from config file %s: %s", configFile, err)
					} else {
						err = pkg.Init()
						if err != nil {
							logger.Logger.Errorf("Error while initializing package %s: %s", pkg.Name(), err)
						} else {
							packagesLoaded[pkg.Name()] = true
							logger.Logger.Infof("package [%16s] load success", pkg.Name())
						}
					}
				} else {
					logger.Logger.Errorf("Package %v init data unmarshal failed.", pkg.Name())
				}
			} else {
				logger.Logger.Errorf("Package %v init data not exist.", pkg.Name())
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

package i18n

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/howeyc/fsnotify"
	"github.com/idealeak/goserver/core"
	"github.com/idealeak/goserver/core/basic"
	"github.com/idealeak/goserver/core/logger"
)

var Config = Configuration{}

type Configuration struct {
	RootPath  string
	hashCodes map[string]string
	watcher   *fsnotify.Watcher
}

func (this *Configuration) Name() string {
	return "i18n"
}

func (this *Configuration) Init() error {
	var err error
	workDir, err := os.Getwd()
	if err != nil {
		return err
	}
	this.RootPath = filepath.Join(workDir, this.RootPath)
	this.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		logger.Logger.Warn(" fsnotify.NewWatcher err:", err)
		return err
	}

	// Process events
	go func() {
		defer func() {
			if err := recover(); err != nil {
				logger.Logger.Warn("watch data director modify goroutine err:", err)
			}
		}()
		for {
			select {
			case ev := <-this.watcher.Event:
				if ev != nil && ev.IsModify() && filepath.Ext(ev.Name) == ".json" {
					core.CoreObject().SendCommand(&fileModifiedCommand{fileName: ev.Name}, false)
					logger.Logger.Trace("fsnotify event:", ev)
				}

			case err := <-this.watcher.Error:
				logger.Logger.Warn("fsnotify error:", err)
			}
		}
		logger.Logger.Warn(this.RootPath, " watcher quit!")
	}()
	this.watcher.Watch(this.RootPath)

	this.hashCodes = make(map[string]string)
	//获得配置目录中所有的数据文件
	var files = []string{}
	filepath.Walk(this.RootPath, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() == false && filepath.Ext(info.Name()) == ".json" {
			files = append(files, info.Name())
		}
		return nil
	})
	//加载这些数据文件
	h := md5.New()
	for _, f := range files {
		logger.Logger.Trace("load file name ======", f)
		buf, err := ioutil.ReadFile(filepath.Join(this.RootPath, f))
		if err != nil {
			logger.Logger.Warn("i18n Config.Init ioutil.ReadFile error", err)
			return err
		}
		kv := make(map[string]string)
		err = json.Unmarshal(buf, &kv)
		if err != nil {
			logger.Logger.Warn("i18n Config.Init json.Unmarshal error", err)
			return err
		}
		nameAndExt := strings.SplitN(f, ".", 2)
		if len(nameAndExt) == 2 {
			lang := nameAndExt[0]
			loc := &locale{lang: lang, message: kv}
			if loc != nil {
				locales.Add(loc)
			}

			h.Reset()
			h.Write(buf)
			this.hashCodes[f] = hex.EncodeToString(h.Sum(nil))
		}
	}
	return nil
}

func (this *Configuration) Close() error {
	this.watcher.Close()
	return nil
}

type fileModifiedCommand struct {
	fileName string
}

func (fmc *fileModifiedCommand) Done(o *basic.Object) error {
	fn := filepath.Base(fmc.fileName)
	hashCode := Config.hashCodes[fn]
	buf, err := ioutil.ReadFile(filepath.Join(Config.RootPath, fn))
	if err != nil {
		logger.Logger.Warn("i18n fileModifiedCommand ioutil.ReadFile error", err)
		return err
	}
	if len(buf) == 0 {
		return nil
	}
	h := md5.New()
	h.Reset()
	h.Write(buf)
	newCode := hex.EncodeToString(h.Sum(nil))
	if newCode != hashCode {
		logger.Logger.Trace("modified file name ======", fn)
		kv := make(map[string]string)
		err = json.Unmarshal(buf, &kv)
		if err != nil {
			logger.Logger.Warn("i18n Config.Init json.Unmarshal error", err)
			return err
		}

		nameAndExt := strings.SplitN(fn, ".", 2)
		if len(nameAndExt) == 2 {
			lang := nameAndExt[0]
			loc, exist := locales.getLocale(lang)
			if exist {
				loc.message = kv
			} else {
				loc = &locale{lang: lang, message: kv}
				if loc != nil {
					locales.Add(loc)
				}
			}
			Config.hashCodes[fn] = newCode
		}
	}

	return nil
}

func init() {
	core.RegistePackage(&Config)
}

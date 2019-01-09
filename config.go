package main

import (
	"bufio"
	"cron/utils"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/getsentry/raven-go"
	"github.com/spf13/viper"
	"log"
	"os"
)

type Config struct {
	Shell                  *string
	User                   *string
	Timezone               *string
	Tasks                  []ConfigTask
	onConfigChangeHandlers []func(before *Config, after *Config)
	configPath             string
	configFile             string
}

type ConfigTask struct {
	Name    string
	Time    string
	User    *string
	Shell   *string
	Command string
	Output  *string
}

func (c *Config) init(path string, filename string) {

	c.configPath = c.formatPath(path)
	c.configFile = filename
	if _, err := os.Stat(fmt.Sprintf("%s/%s", c.configPath, c.configFile)); os.IsNotExist(err) {
		utils.PrintStdOut("Legacy file not found. Using new format configuration.")
		c.initViper()
	} else {
		c.initLegacy()
	}

}

func (c *Config) formatPath(path string) string {
	if len(path) == 1 {
		return path
	}
	if path[len(path)-1:] == "/" {
		return path[:len(path)-1]
	}
	return path
}

func (c *Config) initLegacy() {
	c.loadLegacy()
	utils.WatchFileChange(fmt.Sprintf("%s/%s", c.configPath, c.configFile), func(e fsnotify.Event) {
		before := *c

		for _, f := range c.onConfigChangeHandlers {
			f(&before, c)
		}
	})
}

func (c *Config) loadLegacy() {
	file, err := os.Open(fmt.Sprintf("%s/%s", c.configPath, c.configFile))
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line[:1] == "#" {
			continue
		}
		if line[:6] == "SHELL=" {
			shell := line[6:]
			c.Shell = &shell
		}
		if line[:8] == "CRON_TZ=" {
			tz := line[8:]
			c.Timezone = &tz
		}
		var lc LegacyConfig

		if task := lc.getFromDefaultWithUser(line); task != nil {
			c.Tasks = append(c.Tasks, *task)
			continue
		}

		if task := lc.getFromDefaultWithoutUser(line); task != nil {
			c.Tasks = append(c.Tasks, *task)
			continue
		}

		if task := lc.getFromExtensionsWithUser(line); task != nil {
			c.Tasks = append(c.Tasks, *task)
			continue
		}

		if task := lc.getFromExtensionsWithoutUser(line); task != nil {
			c.Tasks = append(c.Tasks, *task)
			continue
		}

		utils.PrintStdOut("Incorrect line in cron legacy config: %s", line)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func (c *Config) initViper() {
	viper.SetConfigName(c.configFile)
	viper.AddConfigPath(c.configPath)
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		before := *c
		if err := viper.Unmarshal(&c); err != nil {
			raven.CaptureErrorAndWait(err, nil)
			log.Fatalf("Fatal error unmarshal file: %s", err)
		}

		for _, f := range c.onConfigChangeHandlers {
			f(&before, c)
		}
	})

	if err := viper.ReadInConfig(); err != nil {
		raven.CaptureErrorAndWait(err, nil)
		log.Fatalf("Fatal error reding config file: %s", err)
	}

	if err := viper.Unmarshal(&c); err != nil {
		raven.CaptureErrorAndWait(err, nil)
		log.Fatalf("Fatal error unmarshal file: %s", err)
	}
}

func (c *Config) addOnConfigChangeHandler(f func(before *Config, after *Config)) {
	c.onConfigChangeHandlers = append(c.onConfigChangeHandlers, f)
}

func (c *Config) clearOnConfigChangeHandlers() {
	c.onConfigChangeHandlers = c.onConfigChangeHandlers[:0]
}

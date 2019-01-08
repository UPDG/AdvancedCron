package main

import (
	"cron/cron"
	"flag"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/getsentry/raven-go"
	"github.com/joho/godotenv"
	"github.com/logrusorgru/aurora"
	"github.com/spf13/viper"
	"log"
	"os"
	"os/exec"
	"time"
)

var c *cron.Cron
var config Config

func main() {

	c = cron.New()

	if err := godotenv.Load(); err != nil {
		log.Printf("Error loading .env file: %s", err)
	}

	if err := raven.SetDSN(os.Getenv("SENTRY_DSN")); err != nil {
		log.Printf("Sentry error: %s", err)
	}

	var configName = flag.String("config", "config", "filename without ext")
	var configPath = flag.String("path", ".", "path to config")
	flag.Parse()

	viper.SetConfigName(*configName)
	viper.AddConfigPath(*configPath)
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		log.Printf("Config updated. Reseting jobs.")
		loadConfig()
		runCron()
	})

	if err := viper.ReadInConfig(); err != nil {
		raven.CaptureErrorAndWait(err, nil)
		log.Fatalf("Fatal error config file: %s", err)
	}

	loadConfig()

	runCron()

	forever := make(chan bool)
	fmt.Printf(" [*] Running cron. To exit press CTRL+C\n")
	<-forever
}

func loadConfig() {
	if err := viper.Unmarshal(&config); err != nil {
		raven.CaptureErrorAndWait(err, nil)
		log.Fatalf("Fatal error config file: %s \n", err)
	}
}

func runCron() {
	c.Stop()
	c.CleanAll()
	for _, e := range config.Tasks {
		task := e
		err := c.AddFunc(task.Time, func() {
			logTime := time.Now().Local().Format("2006-01-02 15:04:05")

			timeBefore := time.Now()
			out, jobErr := exec.Command("/bin/sh", "-c", task.Command).Output()
			elapsedTime := time.Now().Sub(timeBefore)
			if jobErr != nil {
				raven.CaptureErrorAndWait(jobErr, map[string]string{"command": task.Command, "time": task.Time})
				fmt.Printf("%s Job '%s' finished in %fs - %s: %s\n", logTime, task.Name, elapsedTime.Seconds(), aurora.Red("Failed"), jobErr)
			} else {
				fmt.Printf("%s Job '%s' finished in %fs - %s\n", logTime, task.Name, elapsedTime.Seconds(), aurora.Green("Success"))
			}
			if task.Output != nil {
				f, err := os.OpenFile(*task.Output, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
				if err != nil {
					raven.CaptureErrorAndWait(err, nil)
					panic(err)
				}

				if _, err = f.WriteString(fmt.Sprintf("%s %s", logTime, string(out[:]))); err != nil {
					raven.CaptureErrorAndWait(err, nil)
					log.Printf("Cant write log of '%s' to file: %s \n", task.Name, err)
				}
				if err := f.Close(); err != nil {
					raven.CaptureErrorAndWait(err, nil)
					log.Printf("Cant close log file of '%s': %s \n", task.Name, err)
				}
			}
		}, task.Name)
		if err != nil {
			raven.CaptureErrorAndWait(err, nil)
			log.Printf("Error adding job to cron: %s \n", err)
		}
	}
	c.Start()
}

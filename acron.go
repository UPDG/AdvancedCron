package main

import (
	"cron/cron"
	"cron/utils"
	"flag"
	"fmt"
	"github.com/getsentry/raven-go"
	"github.com/joho/godotenv"
	"github.com/logrusorgru/aurora"
	"log"
	"os"
	"time"
)

func main() {
	var config Config

	c := cron.New()

	if err := godotenv.Load(); err != nil {
		log.Printf("Error loading .env file: %s (using system ENV)", err)
	}

	if err := raven.SetDSN(os.Getenv("SENTRY_DSN")); err != nil {
		log.Printf("Sentry error: %s", err)
	}

	var configName = flag.String("config", "config", "filename without ext")
	var configPath = flag.String("path", ".", "path to config")
	flag.Parse()

	config.init(*configPath, *configName)

	config.addOnConfigChangeHandler(func(before *Config, after *Config) {
		log.Printf("Config updated. Reseting jobs.")
		runCron(c, &config)
	})

	runCron(c, &config)

	forever := make(chan bool)
	utils.PrintStdOut("Cron loaded and running. To exit press CTRL+C")
	<-forever
}

func runCron(c *cron.Cron, config *Config) {
	c.Stop()
	c.CleanAll()

	if config.Timezone != nil {
		if tz, err := time.LoadLocation(*config.Timezone); err != nil {
			raven.CaptureErrorAndWait(err, nil)
			log.Printf("Cant find timezone: %s", err)
		} else {
			c.SetLocation(tz)
		}
	} else {
		c.SetLocation(nil)
	}

	for _, e := range config.Tasks {
		task := e

		if task.Time == "@reboot" {
			if err := c.AddOnceFunc("* * * * * *", taskFunction(config, &task), task.Name); err != nil {
				raven.CaptureErrorAndWait(err, nil)
				log.Printf("Error adding job '%s' to cron: %s", task.Name, err)
			}
			continue
		}

		if task.Time[:5] == "@once" {
			if err := c.AddOnceFunc(task.Time[6:], taskFunction(config, &task), task.Name); err != nil {
				raven.CaptureErrorAndWait(err, nil)
				log.Printf("Error adding job '%s' to cron: %s", task.Name, err)
			}
			continue
		}

		if err := c.AddFunc(task.Time, taskFunction(config, &task), task.Name); err != nil {
			raven.CaptureErrorAndWait(err, nil)
			log.Printf("Error adding job '%s' to cron: %s", task.Name, err)
		}

	}
	c.Start()
}

func taskFunction(config *Config, task *ConfigTask) func() {
	return func() {
		logTime := time.Now().Local().Format("2006-01-02 15:04:05")

		// Choose user
		user := config.User
		if task.Shell != nil {
			user = task.User
		}

		// Choose shell
		shell := config.Shell
		if task.Shell != nil {
			shell = task.Shell
		}

		timeBefore := time.Now()
		out, jobErr := utils.RunCMD(task.Command, user, shell)
		elapsedTime := time.Now().Sub(timeBefore)

		if jobErr != nil {
			raven.CaptureErrorAndWait(jobErr, map[string]string{"command": task.Command, "time": task.Time})
			utils.PrintStdOut("Job '%s' finished in %fs - %s: %s", task.Name, elapsedTime.Seconds(), aurora.Red("Failed"), jobErr)
		} else {
			utils.PrintStdOut("Job '%s' finished in %fs - %s", task.Name, elapsedTime.Seconds(), aurora.Green("Success"))
		}
		if task.Output != nil {
			f, err := os.OpenFile(*task.Output, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
			if err != nil {
				raven.CaptureErrorAndWait(err, nil)
				panic(err)
			}

			if _, err = f.WriteString(fmt.Sprintf("%s %s", logTime, string(out[:]))); err != nil {
				raven.CaptureErrorAndWait(err, nil)
				log.Printf("Cant write log of '%s' to file: %s", task.Name, err)
			}
			if err := f.Close(); err != nil {
				raven.CaptureErrorAndWait(err, nil)
				log.Printf("Cant close log file of '%s': %s", task.Name, err)
			}
		}
	}
}

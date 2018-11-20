package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/albertogviana/docker-firewall/config"
	"github.com/albertogviana/docker-firewall/firewall"
	"github.com/urfave/cli"
)

var pidFile = "/tmp/docker-firewall"
var configPath = "/etc/docker-firewall"

var (
	version   string
	gitCommit string
)

func main() {
	if os.Getenv("CONFIG_PATH") != "" {
		configPath = os.Getenv("CONFIG_PATH")
	}

	if version == "" {
		version = "not specified"
	}

	if gitCommit == "" {
		gitCommit = "not specified"
	}

	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Printf("version: %s\ngit commit: %s\n", c.App.Version, gitCommit)
	}

	app := cli.NewApp()
	app.Name = "docker-firewall"
	app.Usage = "Easy way to apply firewall rules to block docker services."
	app.Version = version
	app.Commands = []cli.Command{
		{
			Name:  "start",
			Usage: "start the service",
			Action: func(c *cli.Context) error {
				start()
				return nil
			},
		},
		{
			Name:  "stop",
			Usage: "stop the service",
			Action: func(c *cli.Context) error {
				stop()
				return nil
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func start() {
	log.Println("Starting docker-firewall")
	config, err := config.NewConfiguration(configPath)
	if err != nil {
		log.Fatal(err)
	}

	firewall, err := firewall.NewFirewall()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Applying rules")
	err = firewall.Apply(config.Config.Rules)
	if err != nil {
		stop()
		log.Fatal(err)
	}
	log.Println("Rules applied")

	err = writePidFile()
	if err != nil {
		stop()
		log.Fatal(err)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGQUIT)
	exitChan := make(chan int)

	go func() {
		for {
			s := <-signalChan
			log.Printf("Received signal: %s", s)

			switch s {
			// kill -SIGHUP XXXX
			case syscall.SIGHUP:
				log.Println("Reloading configuration")
				stop()
				start()

			// kill -SIGTERM XXXX
			case syscall.SIGTERM:
				log.Println("stop and core dump")
				stop()
				exitChan <- 0

			// kill -SIGQUIT XXXX
			case syscall.SIGQUIT:
				log.Println("Stopping the service")
				stop()
				exitChan <- 0

			default:
				log.Println("Unknown signal.")
				stop()
				exitChan <- 1
			}
		}
	}()

	for {
		time.Sleep(10 * time.Second)
		verify, err := firewall.Verify(config.Config.Rules)
		if err != nil {
			log.Printf("Something went wrong: %s", err)
			stop()
			exitChan <- 1
		}

		if !verify {
			log.Println("Applying rules again.")
			firewall.Apply(config.Config.Rules)
		}
	}

	code := <-exitChan
	os.Exit(code)
}

func stop() {
	firewall, err := firewall.NewFirewall()
	if err != nil {
		log.Fatal(err)
	}

	firewall.ClearRule()

	piddata, err := ioutil.ReadFile(pidFile)
	if err != nil {
		log.Fatal(err)
	}
	// Convert the file contents to an integer.
	pid, err := strconv.Atoi(string(piddata))
	if err != nil {
		log.Fatal(err)
	}

	err = syscall.Kill(pid, syscall.SIGTERM)
	if err != nil {
		log.Println(err)
	}

	os.Remove(pidFile)
}

func writePidFile() error {
	if _, err := os.Stat(pidFile); !os.IsNotExist(err) {
		piddata, err := ioutil.ReadFile(pidFile)
		if err != nil {
			return err
		}
		pid, err := strconv.Atoi(string(piddata))
		if err != nil {
			return err
		}
		process, err := os.FindProcess(pid)
		if err != nil {
			return err
		}
		// Send the process a signal zero kill.
		if err := process.Signal(syscall.Signal(0)); err == nil {
			return fmt.Errorf("pid already running: %d", pid)
		}
	}

	return ioutil.WriteFile(pidFile, []byte(fmt.Sprintf("%d", os.Getpid())), 0664)
}

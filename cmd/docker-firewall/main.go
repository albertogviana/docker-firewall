package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/albertogviana/docker-firewall/config"
	"github.com/albertogviana/docker-firewall/firewall"
	"github.com/urfave/cli"
)

var pidFile = "/tmp/docker-firewall"
var configPath = "/etc/docker-firewall"

func main() {
	if os.Getenv("CONFIG_PATH") != "" {
		configPath = os.Getenv("CONFIG_PATH")
	}

	app := cli.NewApp()
	app.Name = "docker-firewall"
	app.Usage = "Easy way to apply firewall rules to block docker services."
	app.Version = "0.0.0"
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

	// setup signal catching
	sigs := make(chan os.Signal, 1)

	// catch all signals since not explicitly listing
	signal.Notify(sigs)

	go func() {
		s := <-sigs
		log.Printf("Received signal: %s", s)
		stop()
		os.Exit(0)
	}()

	select {}
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

	err = syscall.Kill(pid, syscall.SIGKILL)
	if err != nil {
		log.Println("successful shutdown process")
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

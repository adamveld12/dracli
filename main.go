package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"
)

var (
	commands = map[string]func(map[string][]string) error{
		"login":  loginAction,
		"logout": logoutAction,
		"power":  powerStateAction,
		"query":  queryAction,
		"help":   helpAction,
	}

	queryHelp = []Attribute{
		PowerStatus,
		SystemDescription,
		SystemRevision,
		HostName,
		OSName,
		OSVersion,
		ServiceTag,
		ExpServiceCode,
		BiosVersion,
		FirmwareVersion,
		LCCFirmwareVersion,
		IPV4Enabled,
		IPV4Address,
		IPV6Enabled,
		IPV6LinkLocal,
		IPV6Address,
		IPV6SiteLocal,
		MacAddress,
		Batteries,
		FanRedundancy,
		Fans,
		Intrusion,
		PowerSupplyRedundancy,
		PowerSupplies,
		RMVRedundancy,
		RemovableStorage,
		Temperatures,
		Voltages,
		KVMEnabled,
		PowerBudgetData,
		EventLog,
		BootOnce,
		FirstBootDevice,
		VFKLicense,
		User,
		IDRACLog,
	}
)

func main() {
	c, err := ToCommand(os.Args[1:]...)
	if err != nil {
		log.Println(err)
		os.Exit(-1)
	}

	action, ok := commands[c.Name]
	if !ok {
		fmt.Printf("The command \"%s\" was not found.\n", c.Name)
		os.Exit(-1)
	}

	if err := action(c.Arguments); err != nil {
		fmt.Printf("The command exited with an error:\n%v\n", err)
		os.Exit(-1)
	}
}

func powerStateAction(args map[string][]string) error {
	c, err := NewFromCredentials(".")
	if err != nil {
		return err
	}
	rawPs, ok := args[""]
	if !ok || len(rawPs) == 0 {
		return errors.New("specify a power state (on|off|cold_reboot|warm_reboot)")
	}

	var ps PowerState
	switch rawPs[0] {
	case "on":
		ps = PowerOn
	case "off":
		ps = PowerOff
	case "cold_reboot":
		ps = ColdReboot
	case "warm_reboot":
		ps = WarmReboot
	case "nmi":
		ps = NonMaskingInterrupt
	case "graceful_shutdown":
		ps = GracefulShutdown
	default:
		return errors.New("specify a power state (on|off|cold_reboot|warm_reboot)")
	}

	res, err := c.SetPowerState(ps)
	if err != nil {
		return err
	}

	fmt.Printf("%s\n", res)

	return nil
}

func queryAction(args map[string][]string) error {
	c, err := NewFromCredentials(".")
	if err != nil {
		return err
	}

	queryParams, ok := args[""]
	if !ok {
		return errors.New("you should pass query parameters")
	}

	qps := []Attribute{}
	for _, qp := range queryParams {
		qps = append(qps, Attribute(qp))
	}

	output, err := c.Query(qps...)
	if err != nil {
		return err
	}

	fmt.Printf("%s", output)

	watch, watchOk := args["watch"]
	if watchOk && len(watch) > 0 {
		duration, err := time.ParseDuration(watch[0])
		if err != nil {
			return err
		}

		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			for {
				select {
				case <-time.After(duration):
					output, err := c.Query(qps...)
					if err != nil {
						return
					}

					fmt.Printf("%s", output)
				case <-ctx.Done():
					return
				}
			}
		}()

		notifChan := make(chan os.Signal)
		signal.Notify(notifChan, os.Interrupt)
		<-notifChan
		cancel()
	}

	return nil
}

func loginAction(args map[string][]string) error {
	u, ok1 := args["u"]
	p, ok2 := args["p"]
	h, ok3 := args["h"]

	if !ok1 || !ok2 || !ok3 {
		return errors.New("username (-u), password (-p), and a host (-h) must be defined")
	}

	username, password, host := u[0], p[0], h[0]

	credential, err := LoadCredentials(".")
	if (err == nil || os.IsExist(err)) && credential.Host == host {
		return errors.New("you are already logged in")
	}

	fmt.Printf("logging in to %s:%s@%s\n", username, password, host)
	client, err := NewClient(host, true)
	if err != nil {
		return err
	}

	authToken, err := client.Login(username, password)
	if err != nil {
		return err
	}

	if err := SaveCredentials(".", Credential{
		Host:      host,
		AuthToken: authToken,
	}); err != nil {
		return err
	}

	return nil
}

func logoutAction(args map[string][]string) error {
	_, err := LoadCredentials(".")
	if err == nil && os.IsNotExist(err) {
		return nil
	}

	if err := os.Remove("./credentials.json"); err != nil {
		return err
	}
	return nil
}

func helpAction(args map[string][]string) error {
	fmt.Println("login -u [username] -p [password] -h [host]: logs you in")
	fmt.Println("logout: logs you out")
	fmt.Println("power [on|off|nmi|graceful_shutdown|cold_reboot|warm_reboot]: manage power state of the server")
	fmt.Printf("query [-watch 1[s|m|h]] <attribute>,<attribute2>...: gets info about the server's various sensors and attributes\nPossible attributes:")
	for idx, q := range queryHelp {
		if idx%10 == 0 {
			fmt.Println()
		}
		fmt.Printf("%s ", q)
	}
	fmt.Println()
	return nil
}

type Command struct {
	Name      string
	Arguments map[string][]string
}

func ToCommand(args ...string) (Command, error) {
	if len(args) < 1 {
		return Command{}, nil

	}
	c := Command{
		Name:      args[0],
		Arguments: map[string][]string{},
	}

	for i := 1; i < len(args); i++ {
		arg := args[i]
		if arg[0] == '-' {
			argumentV := "true"
			if len(args) > i+1 && arg[1:] != "once" {
				argumentV = args[i+1]
				i++
			}

			c.Arguments[arg[1:]] = []string{argumentV}
		} else {
			v, ok := c.Arguments[""]
			if ok {
				v = append(v, arg)
				c.Arguments[""] = v
			} else {
				c.Arguments[""] = []string{arg}
			}
		}
	}

	return c, nil
}

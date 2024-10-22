package programs

import (
	"bytes"
	"fmt"
	"github.com/spf13/viper"
	"log"
	"os/exec"
)

type GoneRTT struct {
	Conf    *viper.Viper
	Running bool
}

func (g *GoneRTT) SetConfig(v *viper.Viper) {
	g.Conf = v
}

func (g *GoneRTT) IsRunning() bool {
	return g.Running
}

func (g *GoneRTT) Start() error {
	if g.Running {
		return fmt.Errorf("gone-rtt already started")
	}
	shell := exec.Command("docker", "run", "-d",
		"--name", g.Conf.GetString("GONE_RTT_ID"),
		"--network", g.Conf.GetString("NETWORK_NAMESPACE"),
		g.Conf.GetString("GONE_RTT_IMAGE"))

	var out bytes.Buffer
	var berr bytes.Buffer
	shell.Stdout = &out
	shell.Stderr = &berr

	if err := shell.Run(); err != nil {
		fmt.Println("GONE_RTT INFO:", log.Ltime, err, berr.String())
		return err
	}
	fmt.Print("GONE_RTT INFO: ", out.String())

	fmt.Println("GONE_RTT INFO:", log.Ltime, "Started", g.Conf.GetString("GONE_RTT_ID"))
	g.Running = true
	return nil

}

func (g *GoneRTT) Stop() error {
	if !g.Running {
		return fmt.Errorf("gone-rtt not Running")
	}
	shell := exec.Command("docker", "kill", g.Conf.GetString("GONE_RTT_ID"))

	_, err := shell.Output()
	if err != nil {
		fmt.Println("GONE_RTT INFO:", log.Ltime, err)
	}

	fmt.Println("GONE_RTT INFO:", log.Ltime, "Killed", g.Conf.GetString("GONE_RTT_ID"))

	shell = exec.Command("docker", "rm", g.Conf.GetString("GONE_RTT_ID"))

	_, err = shell.Output()
	if err != nil {
		fmt.Println("GONE_RTT INFO:", log.Ltime, err)
	}

	fmt.Println("GONE_RTT INFO:", log.Ltime, "Removed", g.Conf.GetString("GONE_RTT_ID"))
	g.Running = false
	return nil

}

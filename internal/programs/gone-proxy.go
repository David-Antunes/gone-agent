package programs

import (
	"bytes"
	"fmt"
	"github.com/spf13/viper"
	"log"
	"os/exec"
)

type GoneProxy struct {
	Conf    *viper.Viper
	Running bool
}

func (g *GoneProxy) SetConfig(v *viper.Viper) {
	g.Conf = v
}
func (g *GoneProxy) IsRunning() bool {
	return g.Running
}

func (g *GoneProxy) Start() error {
	if g.Running {
		return fmt.Errorf("gone-proxy already started")
	}
	shell := exec.Command("docker", "run", "-d", "--privileged",
		"--name", g.Conf.GetString("GONE_PROXY_ID"),
		"--network", "none",
		"-v", "/tmp:/tmp",
		"-v", "/var/run/docker:/var/run/docker",
		"--ulimit", "memlock=65535",
		"-e", "NETWORK="+g.Conf.GetString("NETWORK"),
		"-e", "PROXY_RTT_SOCKET="+g.Conf.GetString("PROXY_RTT_SOCKET"),
		"-e", "PROXY_SERVER="+g.Conf.GetString("PROXY_SERVER"),
		"-e", "TIMEOUT="+g.Conf.GetString("TIMEOUT"),
		"-e", "NUM_TESTS="+g.Conf.GetString("NUM_TESTS"),
		g.Conf.GetString("GONE_PROXY_IMAGE"))

	var out bytes.Buffer
	var berr bytes.Buffer
	shell.Stdout = &out
	shell.Stderr = &berr

	if err := shell.Run(); err != nil {
		fmt.Println("GONE_PROXY INFO:", log.Ltime, err, berr.String())
		return err
	}
	fmt.Print("GONE_PROXY INFO: ", out.String())

	fmt.Println("GONE_PROXY INFO:", log.Ltime, "Started", g.Conf.GetString("GONE_PROXY_ID"))
	g.Running = true
	return nil
}

func (g *GoneProxy) Stop() error {
	if !g.Running {
		return fmt.Errorf("gone-proxy already stopped")
	}
	shell := exec.Command("docker", "kill", g.Conf.GetString("GONE_PROXY_ID"))

	_, err := shell.Output()
	if err != nil {
		fmt.Println("GONE_PROXY INFO:", log.Ltime, err)
	}

	fmt.Println("GONE_PROXY INFO:", log.Ltime, "Killed", g.Conf.GetString("GONE_PROXY_ID"))

	shell = exec.Command("docker", "rm", g.Conf.GetString("GONE_PROXY_ID"))

	_, err = shell.Output()
	if err != nil {
		fmt.Println("GONE_PROXY INFO:", log.Ltime, err)
	}

	fmt.Println("GONE_PROXY INFO:", log.Ltime, "Removed", g.Conf.GetString("GONE_PROXY_ID"))
	g.Running = false
	return nil

}

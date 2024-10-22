package programs

import (
	"bytes"
	"fmt"
	"github.com/spf13/viper"
	"log"
	"os/exec"
)

type Gone struct {
	Conf    *viper.Viper
	Running bool
}

func (g *Gone) SetConfig(v *viper.Viper) {
	g.Conf = v
}

func (g *Gone) IsRunning() bool {
	return g.Running
}

func (g *Gone) Start() error {
	if g.Running {
		return fmt.Errorf("gone already Running")
	}

	shell := exec.Command("docker", "run", "-d", "--privileged",
		"--name", g.Conf.GetString("GONE_ID"),
		"-p", g.Conf.GetString("SERVER_PORT")+":"+g.Conf.GetString("SERVER_PORT"),
		"-p", g.Conf.GetString("SERVER_ROUTE_PORT")+":"+g.Conf.GetString("SERVER_ROUTE_PORT"),
		"-v", "/tmp:/tmp",
		"-v", "/var/run/docker:/var/run/docker",
		"-v", "/var/run/docker.sock:/var/run/docker.sock",
		"-v", "/proc:/proc",
		"-e", "GRAPHDB="+g.Conf.GetString("GRAPHDB"),
		"-e", "GRAPHDB_USER=\""+g.Conf.GetString("GRAPHDB_USER")+"\"",
		"-e", "GRAPHDB_PASSWORD=\""+g.Conf.GetString("GRAPHDB_PASSWORD")+"\"",
		"-e", "ID="+g.Conf.GetString("ID"),
		"-e", "NETWORK_NAMESPACE="+g.Conf.GetString("NETWORK_NAMESPACE"),
		"-e", "PRIMARY="+g.Conf.GetString("PRIMARY"),
		"-e", "PRIMARY_ROUTE_PORT="+g.Conf.GetString("PRIMARY_ROUTE_PORT"),
		"-e", "PRIMARY_SERVER_PORT="+g.Conf.GetString("PRIMARY_SERVER_PORT"),
		"-e", "PRIMARY_SERVER_IP="+g.Conf.GetString("PRIMARY_SERVER_IP"),
		"-e", "PROXY_RTT_SOCKET="+g.Conf.GetString("PROXY_RTT_SOCKET"),
		"-e", "PROXY_RTT_UPDATE_MS="+g.Conf.GetString("PROXY_RTT_UPDATE_MS"),
		"-e", "PROXY_SERVER="+g.Conf.GetString("PROXY_SERVER"),
		"-e", "SERVER_IP="+g.Conf.GetString("SERVER_IP"),
		"-e", "SERVER_PORT="+g.Conf.GetString("SERVER_PORT"),
		"-e", "SERVER_ROUTE_PORT="+g.Conf.GetString("SERVER_ROUTE_PORT"),
		"-e", "NUM_TESTS="+g.Conf.GetString("NUM_TESTS"),
		g.Conf.GetString("GONE_IMAGE"))

	var out bytes.Buffer
	var berr bytes.Buffer
	shell.Stdout = &out
	shell.Stderr = &berr

	if err := shell.Run(); err != nil {
		fmt.Println("GONE INFO: ", log.Ltime, err, berr.String())
		return err
	}
	fmt.Print("GONE INFO: ", out.String())

	fmt.Println("GONE INFO:", log.Ltime, "Started", g.Conf.GetString("GONE_ID"))
	g.Running = true
	return nil
}

func (g *Gone) Stop() error {
	if !g.Running {
		return fmt.Errorf("gone not running")
	}
	shell := exec.Command("docker", "kill", g.Conf.GetString("GONE_ID"))

	_, err := shell.Output()
	if err != nil {
		fmt.Println("GONE INFO:", log.Ltime, err)
	}

	fmt.Println("GONE INFO:", log.Ltime, "Killed", g.Conf.GetString("GONE_ID"))

	shell = exec.Command("docker", "rm", g.Conf.GetString("GONE_ID"))

	_, err = shell.Output()
	if err != nil {
		fmt.Println("GONE INFO:", log.Ltime, err)
	}

	fmt.Println("GONE INFO:", log.Ltime, "Removed", g.Conf.GetString("GONE_ID"))
	g.Running = false
	return nil
}

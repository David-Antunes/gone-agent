package main

import (
	"bytes"
	"github.com/David-Antunes/gone-agent/internal"
	"github.com/David-Antunes/gone-agent/internal/programs"
	"github.com/spf13/viper"
	"log"
	"os"
	"os/exec"
	"sort"
	"strings"
)

var agentLog = log.New(os.Stdout, "AGENT INFO: ", log.Ltime)

func setEnvVariables() {

	viper.SetConfigFile(".env")
	if err := viper.ReadInConfig(); err != nil {
		agentLog.Println(err)
	}
	viper.SetDefault("PRIMARY_IP", "192.168.1.1")
	viper.SetDefault("SERVER_IP", "192.168.1.1")
	viper.SetDefault("PORT", "3300")
	viper.SetDefault("NETWORK_NAMESPACE", "gone_net")
	viper.SetDefault("GONE_ID", "gone")
	viper.SetDefault("GONE_PROXY_ID", "proxy-gone")
	viper.SetDefault("GONE_RTT_ID", "gone-rtt")
	viper.SetDefault("GONE_IMAGE", "gone")
	viper.SetDefault("GONE_PROXY_IMAGE", "gone-proxy")
	viper.SetDefault("GONE_RTT_IMAGE", "gone-rtt")
	viper.SetConfigType("env")
	err := viper.WriteConfig()
	if err != nil {
		agentLog.Println(err)
	}
}

func printVariables(v *viper.Viper) {
	settings := v.AllSettings()
	sortedList := make([]string, 0, len(settings))
	for id, _ := range settings {
		sortedList = append(sortedList, id)
	}

	sort.Strings(sortedList)

	for _, id := range sortedList {
		agentLog.Println(id, settings[id])
	}

}

func main() {

	setEnvVariables()
	printVariables(viper.GetViper())
	validateEnvironment()

	goneConf, err := getConfig("/tmp/gone.env")
	if err != nil {
		agentLog.Println(err)
		os.Exit(1)
	}

	goneProxyConf, err := getConfig("/tmp/gone-proxy.env")
	if err != nil {
		agentLog.Println(err)
		os.Exit(1)
	}

	goneRttConf := viper.New()

	server := internal.NewServer(":" + viper.GetString("PORT"))
	goneConf.SetDefault("GONE_ID", viper.GetString("GONE_ID"))
	goneConf.SetDefault("GONE_IMAGE", viper.GetString("GONE_IMAGE"))
	goneProxyConf.SetDefault("GONE_PROXY_ID", viper.GetString("GONE_PROXY_ID"))
	goneProxyConf.SetDefault("NETWORK", viper.GetString("NETWORK"))
	goneProxyConf.SetDefault("GONE_PROXY_IMAGE", viper.GetString("GONE_PROXY_IMAGE"))
	goneRttConf.SetDefault("GONE_RTT_ID", viper.GetString("GONE_RTT_ID"))
	goneRttConf.SetDefault("NETWORK_NAMESPACE", viper.GetString("NETWORK_NAMESPACE"))
	goneRttConf.SetDefault("GONE_RTT_IMAGE", viper.GetString("GONE_RTT_IMAGE"))
	server.AddGone(&programs.Gone{Conf: goneConf, Running: true})
	server.AddGoneProxy(&programs.GoneProxy{Conf: goneProxyConf, Running: true})
	server.AddGoneRTT(&programs.GoneRTT{Conf: goneRttConf, Running: true})

	err = server.Serve()
	if err != nil {
		agentLog.Println(err)
		return
	}
}
func checkNS(id string) bool {
	// Find id of network to pass to proxy

	shell := exec.Command("docker", "inspect", id, "--format", "{{.ID}}")

	var b bytes.Buffer
	shell.Stdout = &b
	//out, err := shell.Output()
	if err := shell.Run(); err != nil {
		return false
	}
	ns := strings.Trim(b.String()[:12], " ")
	ns = strings.Trim(ns, "\n")
	//fmt.Println(ns)
	viper.SetDefault("NETWORK", ns[:12])
	return true
}

func validateEnvironment() {

	if !checkNS(viper.GetString("NETWORK_NAMESPACE")) {
		agentLog.Println("could not find appropriate network namespace")
		os.Exit(1)
	}

	if !checkGone(viper.GetString("GONE_ID")) {
		agentLog.Println("Could not find gone container")
		os.Exit(1)
	}
	if !checkProxy(viper.GetString("GONE_PROXY_ID")) {
		agentLog.Println("Could not find gone-proxy container")
		os.Exit(1)
	}
	if !checkRTT(viper.GetString("GONE_RTT_ID")) {
		agentLog.Println("Could not find gone-rtt container")
		os.Exit(1)
	}

}

func checkGone(id string) bool {
	shell := exec.Command("docker", "inspect", id)
	_, err := shell.Output()
	if err != nil {
		agentLog.Println(err)
		return false
	}
	return true
}

func checkProxy(id string) bool {
	shell := exec.Command("docker", "inspect", id)
	_, err := shell.Output()
	if err != nil {
		agentLog.Println(err)
		return false
	}
	return true
}

func checkRTT(id string) bool {
	shell := exec.Command("docker", "inspect", id)
	_, err := shell.Output()
	if err != nil {
		agentLog.Println(err)
		return false
	}
	return true
}

func getConfig(path string) (*viper.Viper, error) {
	v := viper.New()
	v.SetConfigFile(path)
	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}
	agentLog.Println()
	agentLog.Println("Loaded variables form", path)
	printVariables(v)

	return v, nil
}

package internal

import (
	"bytes"
	"encoding/json"
	"github.com/David-Antunes/gone-agent/internal/api"
	"github.com/David-Antunes/gone-agent/internal/programs"
	"github.com/spf13/viper"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"time"
)

var serverLog = log.New(os.Stdout, "SERVER INFO: ", log.Ltime)

type Server struct {
	sync.Mutex
	ips       map[string]string
	server    http.Server
	socket    net.Listener
	gone      *programs.Gone
	goneProxy *programs.GoneProxy
	goneRTT   *programs.GoneRTT
}

func NewServer(port string) *Server {
	socket, err := net.Listen("tcp", port)

	if err != nil {
		panic(err)
	}

	s := &Server{
		Mutex:     sync.Mutex{},
		ips:       make(map[string]string),
		server:    http.Server{},
		socket:    socket,
		gone:      nil,
		goneProxy: nil,
		goneRTT:   nil,
	}

	m := http.NewServeMux()
	m.HandleFunc("/ping", s.ping)
	m.HandleFunc("/restart", s.restart)
	m.HandleFunc("/register", s.register)
	m.HandleFunc("/start", s.start)
	m.HandleFunc("/stop", s.stop)

	s.server = http.Server{
		Handler: m,
	}

	s.contactPrimary(viper.GetString("PRIMARY_IP")+port, viper.GetString("SERVER_IP")+port)

	return s
}

func (s *Server) AddGone(p *programs.Gone) {
	s.gone = p
}

func (s *Server) AddGoneProxy(p *programs.GoneProxy) {
	s.goneProxy = p
}

func (s *Server) AddGoneRTT(p *programs.GoneRTT) {
	s.goneRTT = p
}

func (s *Server) Serve() error {
	if err := s.server.Serve(s.socket); err != nil {
		return err
	}
	return nil
}

func (s *Server) ping(w http.ResponseWriter, r *http.Request) {}

func (s *Server) restart(w http.ResponseWriter, r *http.Request) {
	s.Lock()

	s.BroadcastStop()

	if err := s.gone.Stop(); err != nil {
		serverLog.Println(err)
	}
	if err := s.goneProxy.Stop(); err != nil {
		serverLog.Println(err)
	}
	if err := s.goneRTT.Stop(); err != nil {
		serverLog.Println(err)
	}

	s.clearNS()
	if err := s.goneRTT.Start(); err != nil {
		serverLog.Println(err)
	}
	if err := s.goneProxy.Start(); err != nil {
		serverLog.Println(err)
	}
	time.Sleep(time.Second)
	if err := s.gone.Start(); err != nil {
		serverLog.Println(err)
	}
	time.Sleep(time.Second)
	s.BroadcastStart()

	s.Unlock()
}

func (s *Server) register(w http.ResponseWriter, r *http.Request) {
	req := &api.RegisterRequest{}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		serverLog.Println(err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}

	s.Lock()
	if _, ok := s.ips[req.Ip]; !ok {
		if resp, err := json.Marshal(api.RegisterResponse{Ips: s.ips}); err != nil {
			serverLog.Println(err)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		} else {
			if _, err = w.Write(resp); err != nil {
				serverLog.Println(err)
			}
		}
		s.ips[req.Ip] = req.Ip
		serverLog.Println("Added", req.Ip)
	} else {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}
	s.Unlock()
}

func (s *Server) start(w http.ResponseWriter, r *http.Request) {
	result := true
	s.Lock()

	if err := s.goneRTT.Start(); err != nil {
		serverLog.Println(err)
		result = false
	}
	time.Sleep(time.Second)
	if err := s.goneProxy.Start(); err != nil {
		serverLog.Println(err)
		result = false
	}
	time.Sleep(time.Second)
	if err := s.gone.Start(); err != nil {
		serverLog.Println(err)
		result = false
	}
	time.Sleep(time.Second)

	s.Unlock()
	if !result {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}
}

func (s *Server) stop(w http.ResponseWriter, r *http.Request) {
	result := true
	s.Lock()

	if err := s.gone.Stop(); err != nil {
		serverLog.Println(err)
		result = false
	}
	if err := s.goneProxy.Stop(); err != nil {
		serverLog.Println(err)
		result = false
	}
	if err := s.goneRTT.Stop(); err != nil {
		serverLog.Println(err)
		result = false
	}

	s.clearNS()
	s.Unlock()

	if !result {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}
}

func (s *Server) Broadcast(path string) {
	for _, ip := range s.ips {

		req, err := http.NewRequest(http.MethodGet, "http://"+ip+path, nil)
		serverLog.Println(ip, path)
		if err != nil {
			serverLog.Println(err)
		}
		_, err = http.DefaultClient.Do(req)
		if err != nil {
			serverLog.Println(err)
		}
	}
}

func (s *Server) BroadcastStart() {
	s.Broadcast("/start")
}

func (s *Server) BroadcastStop() {
	s.Broadcast("/stop")
}

func (s *Server) clearNS() {
	shell := exec.Command("docker", "unpause", "$(docker ps --filter \"network="+viper.GetString("NETWORK_NAMESPACE")+"\")")

	if _, err := shell.Output(); err != nil {
		//serverLog.Println(err)
	}

	shell = exec.Command("docker", "kill", "$(docker ps -q --filter \"network="+viper.GetString("NETWORK_NAMESPACE")+"\")")

	if _, err := shell.Output(); err != nil {
		//serverLog.Println(err)
	}

	shell = exec.Command("docker", "rm", "$(docker ps -q --filter \"network="+viper.GetString("NETWORK_NAMESPACE")+"\")")

	if _, err := shell.Output(); err != nil {
		//serverLog.Println(err)
	}
}

func (s *Server) contactPrimary(primary string, ip string) {
	serverLog.Println(primary, ip)
	if primary == ip {
		serverLog.Println("PRIMARY")
		serverLog.Println("waiting for nodes")
		return
	}

	var body []byte
	var err error
	if body, err = json.Marshal(&api.RegisterRequest{
		Ip: ip,
	}); err != nil {
		panic(err)
	}

	var req *http.Request
	if req, err = http.NewRequest(http.MethodPost, "http://"+primary+"/register", bytes.NewReader(body)); err != nil {
		panic(err)
	}
	r, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}

	resp := &api.RegisterResponse{}

	d := json.NewDecoder(r.Body)

	if err = d.Decode(resp); err != nil {
		panic(err)
	}
	s.ips[primary] = primary
	serverLog.Println(resp)

	for _, node := range resp.Ips {
		s.ips[node] = node
		if req, err = http.NewRequest(http.MethodPost, "http://"+node+"/register", bytes.NewReader(body)); err != nil {
			panic(err)
		}
		_, err = http.DefaultClient.Do(req)
		if err != nil {
			panic(err)
		}
	}

}

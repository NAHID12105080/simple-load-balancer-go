package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type Server interface {
	Address() string
	isAlive() bool
	Serve(res http.ResponseWriter, req *http.Request)
}

type simpleServer struct {
	address string
	proxy   *httputil.ReverseProxy
}

type loadBalancer struct {
	port             string
	roundRobinIndex  int
	servers          []Server
}

func NewLoadBalancer(port string, servers []Server) *loadBalancer {
	return &loadBalancer{
		port:            port,
		roundRobinIndex: 0,
		servers:         servers,
	}
}

func (lb *loadBalancer) AddServer(server Server) {
	lb.servers = append(lb.servers, server)
}

func (lb *loadBalancer) getNextAvailableServer() Server {
	if len(lb.servers) == 0 {
		panic("No servers available")
	}

	for {
		server := lb.servers[lb.roundRobinIndex%len(lb.servers)]
		lb.roundRobinIndex++
		if server.isAlive() {
			return server
		}
	}
}

func (lb *loadBalancer) serverProxy(rw http.ResponseWriter, req *http.Request) {
	targetServer := lb.getNextAvailableServer()
	fmt.Println("Proxying request to server:", targetServer.Address())
	targetServer.Serve(rw, req)
}

func handleError(err error) {
	if err != nil {
		panic(err)
	}
}

func newSimpleServer(address string) *simpleServer {
	serverUrl, err := url.Parse(address)
	handleError(err)
	return &simpleServer{
		address: address,
		proxy:   httputil.NewSingleHostReverseProxy(serverUrl),
	}
}

func (s *simpleServer) Address() string {
	return s.address
}

func (s *simpleServer) isAlive() bool {
	// Always return true in this example. Add health checks as needed.
	return true
}

func (s *simpleServer) Serve(rw http.ResponseWriter, req *http.Request) {
	s.proxy.ServeHTTP(rw, req)
}

func main() {
	servers := []Server{
		newSimpleServer("https://www.facebook.com"),
		newSimpleServer("https://www.google.com"),
		newSimpleServer("https://www.youtube.com"),
	}

	lb := NewLoadBalancer("8080", servers)
	handleRedirect := func(rw http.ResponseWriter, req *http.Request) {
		lb.serverProxy(rw, req)
	}

	http.HandleFunc("/", handleRedirect)
	fmt.Println("Load Balancer is running on port 8080")
	err := http.ListenAndServe(":8080", nil)
	handleError(err)
}

package main

import (
	
	"fmt"
	"net/http"
	"net/http/httputil"

)
type Server interface {
	Address() string
	isAlive() bool
	Serve(res http.ResponseWriter, req *http.Request)
	
}

type simpleServer struct {
	address string
	proxy *httputil.ReverseProxy

}

type loadBalancer struct {
	port string
	roundRobinIndex int
	servers []*simpleServer
}

func NewLoadBalancer(port string) *loadBalancer{
	return &loadBalancer{
		port: port,
		roundRobinIndex: 0,
		servers: make([]*simpleServer,0),
	}
}

func (lb *loadBalancer) AddServer(server *simpleServer){}

func (lb *loadBalancer) getNextAvailableServer() Server{
	server:= lb.servers[lb.roundRobinIndex]
	for !server.isAlive() {
		lb.roundRobinIndex++
		server=lb.servers[lb.roundRobinCount%len(lb.servers)]
	}
	lb.roundRobinIndex++
	return server


}

func (lb *loadBalancer) serverProxy(rw http.ResponseWriter, req *http.Request){
	targetServer:= lb.getNextAvailableServer()
	fmt.Println("Proxying request to server: ", targetServer.address)
	targetServer.Serve(rw,req) 
}

func handleError(err error){
	if err != nil{
		panic(err)
		
	}
}

func newSimpleServer(address string ) *simpleServer{
	serverUrl,err= url.Parse(address)
	handleError(err)
	return &simpleServer{
		address: address,
		proxy: httputil.NewSingleHostReverseProxy(serverUrl),
	}
}

func Address(s *simpleServer) string{
	return s.address
}

func (s *simpleServer) isAlive() bool {
	return true
}

func (s *simpleServer) Serve(rw http.ResponseWriter, req *http.Request){
	s.proxy.ServeHTTP(rw,req)
}

func main(){
	servers:= []Server{
		newSimpleServer("https://www.facebook.com"),
		newSimpleServer("https://www.google.com"),
		newSimpleServer("https://www.youtube.com"),

	}

	lb:= NewLoadBalancer("8080",servers)
	handleRedirect:= func(rw http.ResponseWriter, req *http.Request){
		lb.serverProxy(rw,req)
	}
	http.HandleFunc("/",handleRedirect)
	fmt.println("Load Balancer is running on port 8080")
	http.ListenAndServe(":8080",nil)

	
}
package main

import "flag"

var port int
var pool string

func init() {
	flag.IntVar(&port, "port", 1234, "Port to listen on")	
	flag.StringVar(&pool, "pool", "127.0.0.1:1234", "pool")
	flag.Parse()
}

package main

import "github.com/sapcc/pod-readiness/pod"

func main() {
	p := pod.New()
	p.StartServer()
}

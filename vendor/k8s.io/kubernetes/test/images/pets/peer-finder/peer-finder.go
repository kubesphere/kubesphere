/*
Copyright 2014 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// A small utility program to lookup hostnames of endpoints in a service.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/util/sets"
)

const (
	pollPeriod = 1 * time.Second
)

var (
	onChange  = flag.String("on-change", "", "Script to run on change, must accept a new line separated list of peers via stdin.")
	onStart   = flag.String("on-start", "", "Script to run on start, must accept a new line separated list of peers via stdin.")
	svc       = flag.String("service", "", "Governing service responsible for the DNS records of the domain this pod is in.")
	namespace = flag.String("ns", "", "The namespace this pod is running in. If unspecified, the POD_NAMESPACE env var is used.")
	domain    = flag.String("domain", "cluster.local", "The Cluster Domain which is used by the Cluster.")
)

func lookup(svcName string) (sets.String, error) {
	endpoints := sets.NewString()
	_, srvRecords, err := net.LookupSRV("", "", svcName)
	if err != nil {
		return endpoints, err
	}
	for _, srvRecord := range srvRecords {
		// The SRV records ends in a "." for the root domain
		ep := fmt.Sprintf("%v", srvRecord.Target[:len(srvRecord.Target)-1])
		endpoints.Insert(ep)
	}
	return endpoints, nil
}

func shellOut(sendStdin, script string) {
	log.Printf("execing: %v with stdin: %v", script, sendStdin)

	cmd := exec.Command(script)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatalf("Failed to get stdin pipe: %v", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalf("Failed to get stdout pipe: %v", err)
	}

	cmd.Start()

	stdin.Write([]byte(sendStdin))
	stdin.Close()

	out, err := ioutil.ReadAll(stdout)
	if err != nil {
		log.Fatalf("Failed to execute %v: %v, err: %v", script, string(out), err)
	}

	log.Print(string(out))
}

func main() {
	flag.Parse()

	ns := *namespace
	if ns == "" {
		ns = os.Getenv("POD_NAMESPACE")
	}
	if *svc == "" || ns == "" || (*onChange == "" && *onStart == "") {
		log.Fatalf("Incomplete args, require -on-change and/or -on-start, -service and -ns or an env var for POD_NAMESPACE.")
	}

	hostname, err := os.Hostname()
	if err != nil {
		log.Fatalf("Failed to get hostname: %s", err)
	}

	svcLocalSuffix := strings.Join([]string{"svc", *domain}, ".")
	myName := strings.Join([]string{hostname, *svc, ns, svcLocalSuffix}, ".")
	script := *onStart
	if script == "" {
		script = *onChange
		log.Printf("No on-start supplied, on-change %v will be applied on start.", script)
	}
	for newPeers, peers := sets.NewString(), sets.NewString(); script != ""; time.Sleep(pollPeriod) {
		newPeers, err = lookup(*svc)
		if err != nil {
			log.Printf("%v", err)
			continue
		}
		if newPeers.Equal(peers) || !newPeers.Has(myName) {
			continue
		}
		peerList := newPeers.List()
		sort.Strings(peerList)
		log.Printf("Peer list updated\nwas %v\nnow %v", peers.List(), newPeers.List())
		shellOut(strings.Join(peerList, "\n"), script)
		peers = newPeers
		script = *onChange
	}
	// TODO: Exit if there's no on-change?
	log.Printf("Peer finder exiting")
}

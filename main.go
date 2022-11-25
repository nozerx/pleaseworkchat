package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
)


func createNode() {
	host, err := libp2p.New(libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"))
	if err != nil {
		fmt.Println("Error in establishing a node")
	} else {
		fmt.Println("Succesfull in establishing the node", host.ID().String())
	}
	ctx := context.Background()
	kad_dht, err := dht.New(ctx, host)
	if err != nil {
		fmt.Println("Error while creating a new dht")
	} else {
		fmt.Println("Succesful in creating a new dht")
	}
	err = kad_dht.Bootstrap(ctx)
	if err != nil {
		fmt.Println("Error while getting the dht to bootstrap state")
	} else {
		fmt.Println("Successfull in getting the dht to bootstrap state")
	}
	var waitGroup sync.WaitGroup
	var (
		totalbootstrapnodes     = 0
		connectedbootstrapnodes = 0
	)
	for _, peerAddr := range dht.GetDefaultBootstrapPeerAddrInfos() {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			err := host.Connect(ctx, peerAddr)
			if err != nil {
				fmt.Println("Error while connecting to bootstrap node")
				totalbootstrapnodes++
			} else {
				fmt.Println("Connected to bootstrap node")
				connectedbootstrapnodes++
				totalbootstrapnodes++
			}

		}()
		waitGroup.Wait()
		routingDiscovery := drouting.NewRoutingDiscovery()

	}
}

func main(
	createNode()
)
// func initDHT()
// func advertise()
// func discoverPeers()

package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	dutil "github.com/libp2p/go-libp2p/p2p/discovery/util"
)

const service string = "chatapp/rezon"

func createNode() (host.Host, context.Context) {
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

	}
	waitGroup.Wait()
	fmt.Println(len(host.Network().Peers()))
	return host, ctx
}

func main() {
	host, ctx := createNode()
	sub, topic := newPubSub(ctx, host, service)
	go discoverPeers(ctx, host)
	go getMessageFromSubscription(ctx, sub)
	streamHandler(ctx, topic)
}

func initDHT(ctx context.Context, host host.Host) *dht.IpfsDHT {
	kad_dht, err := dht.New(ctx, host)
	if err != nil {
		fmt.Println("Error while creating a new dht")
	} else {
		fmt.Println("Successfully created a new dht")
	}
	return kad_dht
}

func discoverPeers(ctx context.Context, host host.Host) {
	kad_dht := initDHT(ctx, host)
	routingDiscovery := drouting.NewRoutingDiscovery(kad_dht)
	dutil.Advertise(ctx, routingDiscovery, "service1")
	fmt.Println(len(host.Network().Peers()))

	connectedList := []peer.AddrInfo{}
	for len(connectedList) < 5 {
		peerChan, err := routingDiscovery.FindPeers(ctx, "service1")
		fmt.Println(len(host.Network().Peers()))

		if err != nil {
			fmt.Println("Error while finding peers after advertising")
		} else {
			fmt.Println("Successfully found peers after advertising")
		}
		fmt.Println("Connected to ", len(connectedList), "out of 5 nodes")
		fmt.Println(connectedList)
		loopSkip := false
		for peerAddr := range peerChan {
			if peerAddr.ID == host.ID() {
				continue
			}
			for _, prAddr := range connectedList {
				if prAddr.ID == peerAddr.ID {
					loopSkip = true
				}
			}
			if loopSkip {
				continue
			}
			err := host.Connect(ctx, peerAddr)
			if err != nil {
				// fmt.Println("Error while connecting to the peer :", peerAddr.ID, peerAddr.Addrs)
			} else {
				fmt.Println("Successfully connected to the peer :", peerAddr.ID, peerAddr.Addrs)
				connectedList = append(connectedList, peerAddr)
			}
		}
	}
	fmt.Println("Announcement: Connected to these peers")
	fmt.Println(connectedList)

}

func newPubSub(ctx context.Context, host host.Host, service string) (*pubsub.Subscription, *pubsub.Topic) {
	pubSubService, err := pubsub.NewGossipSub(ctx, host)
	if err != nil {
		fmt.Println("Error while establishing a pubsub service")
	} else {
		fmt.Println("Successfull in establishing a pubsub service")
	}
	topic, err := pubSubService.Join(service)
	if err != nil {
		fmt.Println("Error whie joining the", service, "service")
	} else {
		fmt.Println("Successfully joined the ", service, "service")
	}
	subscription, err := topic.Subscribe()
	if err != nil {
		fmt.Println("Error while subscribing to ", topic, "service")
	} else {
		fmt.Println("Successfully subscribed to ", topic, "service")
	}
	return subscription, topic
}

func getMessageFromSubscription(ctx context.Context, sub *pubsub.Subscription) {
	for {
		m, err := sub.Next(ctx)
		if err != nil {
			fmt.Println("Error while reading from subsciption", sub.Topic())
		} else {
			fmt.Println("Successfully read from subscription", sub.Topic())
		}
		fmt.Println(m.ReceivedFrom, " : ->", string(m.Message.Data))
	}
}

func streamHandler(ctx context.Context, topic *pubsub.Topic) {
	reader := bufio.NewReader(os.Stdin)
	for {
		s, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error while reading a string")
		}
		err = topic.Publish(ctx, []byte(s))
		if err != nil {
			fmt.Println("Error while publishing the message :>", s)
		}

	}
}

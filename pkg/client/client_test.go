package client

import (
	"testing"
	"time"

	"github.com/Indra-Labs/indra/pkg/ifc"
	"github.com/Indra-Labs/indra/pkg/key/address"
	"github.com/Indra-Labs/indra/pkg/key/prv"
	"github.com/Indra-Labs/indra/pkg/key/pub"
	"github.com/Indra-Labs/indra/pkg/key/signer"
	"github.com/Indra-Labs/indra/pkg/node"
	"github.com/Indra-Labs/indra/pkg/nonce"
	"github.com/Indra-Labs/indra/pkg/session"
	"github.com/Indra-Labs/indra/pkg/sha256"
	"github.com/Indra-Labs/indra/pkg/slice"
	"github.com/Indra-Labs/indra/pkg/testutils"
	"github.com/Indra-Labs/indra/pkg/transport"
	"github.com/Indra-Labs/indra/pkg/wire"
	"github.com/Indra-Labs/indra/pkg/wire/confirm"
	log2 "github.com/cybriq/proc/pkg/log"
	"github.com/cybriq/qu"
)

func TestPing(t *testing.T) {
	log2.CodeLoc = true
	// log2.SetLogLevel(log2.Trace)
	const nTotal = 4
	var clients [nTotal]*Client
	var nodes [nTotal]*node.Node
	var transports [nTotal]ifc.Transport
	var e error
	for i := range transports {
		transports[i] = transport.NewSim(nTotal)
	}
	for i := range nodes {
		var hdrPrv *prv.Key
		if hdrPrv, e = prv.GenerateKey(); check(e) {
			t.Error(e)
			t.FailNow()
		}
		hdrPub := pub.Derive(hdrPrv)
		addr := slice.GenerateRandomAddrPortIPv6()
		nodes[i], _ = node.New(addr, hdrPub, hdrPrv, transports[i])
		if clients[i], e = New(transports[i], hdrPrv, nodes[i], nil); check(e) {
			t.Error(e)
			t.FailNow()
		}
		clients[i].AddrPort = nodes[i].AddrPort
	}
	// add each node to each other's Nodes except itself.
	for i := range nodes {
		for j := range nodes {
			if i == j {
				continue
			}
			clients[i].Nodes = append(clients[i].Nodes, nodes[j])
		}
	}
	// Start up the clients.
	for _, v := range clients {
		go v.Start()
	}
	pn := nonce.NewID()
	var ks *signer.KeySet
	if _, ks, e = signer.New(); check(e) {
		t.Error(e)
		t.FailNow()
	}
	var hop [nTotal - 1]*node.Node
	for i := range clients[0].Nodes {
		hop[i] = clients[0].Nodes[i]
	}
	os := wire.Ping(pn, clients[0].Node, hop, ks)
	// log.I.S(os)
	quit := qu.T()
	log.I.S("sending ping with ID", os[len(os)-1].(*confirm.OnionSkin))
	clients[0].RegisterConfirmation(func(cf *confirm.OnionSkin) {
		log.I.S("received ping confirmation ID", cf)
		quit.Q()
	}, os[len(os)-1].(*confirm.OnionSkin))
	o := os.Assemble()
	b := wire.EncodeOnion(o)
	hop[0].Send(b)
	// go func() {
	// 	time.Sleep(time.Second)
	// 	quit.Q()
	// 	t.Error("ping got stuck")
	// }()
	<-quit.Wait()
	for _, v := range clients {
		v.Shutdown()
	}
}

func TestSendKeys(t *testing.T) {
	log2.CodeLoc = true
	// log2.SetLogLevel(log2.Trace)
	const nTotal = 6
	var clients [nTotal]*Client
	var nodes [nTotal]*node.Node
	var transports [nTotal]ifc.Transport
	var e error
	for i := range transports {
		transports[i] = transport.NewSim(nTotal)
	}
	for i := range nodes {
		var hdrPrv *prv.Key
		if hdrPrv, e = prv.GenerateKey(); check(e) {
			t.Error(e)
			t.FailNow()
		}
		hdrPub := pub.Derive(hdrPrv)
		addr := slice.GenerateRandomAddrPortIPv4()
		nodes[i], _ = node.New(addr, hdrPub, hdrPrv, transports[i])
		if clients[i], e = New(transports[i], hdrPrv, nodes[i], nil); check(e) {
			t.Error(e)
			t.FailNow()
		}
		clients[i].AddrPort = nodes[i].AddrPort
	}
	// add each node to each other's Nodes except itself.
	for i := range nodes {
		for j := range nodes {
			if i == j {
				continue
			}
			clients[i].Nodes = append(clients[i].Nodes, nodes[j])
		}
	}
	// Start up the clients.
	for _, v := range clients {
		go v.Start()
	}
	pn := nonce.NewID()
	var ks *signer.KeySet
	if _, ks, e = signer.New(); check(e) {
		t.Error(e)
		t.FailNow()
	}
	var hop [nTotal - 1]*node.Node
	for i := range clients[0].Nodes {
		hop[i] = clients[0].Nodes[i]
	}
	var hdr, pld *pub.Key
	if _, _, hdr, pld, e = testutils.GenerateTestKeyPairs(); check(e) {
		t.Error(e)
		t.FailNow()
	}
	os := wire.SendKeys(pn, hdr, pld, clients[0].Node, hop, ks)
	// log.I.S(os)
	quit := qu.T()
	log.I.S("sending sendkeys with ID", os[len(os)-1].(*confirm.OnionSkin))
	clients[0].RegisterConfirmation(func(cf *confirm.OnionSkin) {
		log.I.S("received sendkeys confirmation ID", cf)
		quit.Q()
	}, os[len(os)-1].(*confirm.OnionSkin))
	o := os.Assemble()
	b := wire.EncodeOnion(o)
	hop[0].Send(b)
	// go func() {
	// 	time.Sleep(time.Second * 2)
	// 	quit.Q()
	// 	t.Error("sendkeys got stuck")
	// }()
	<-quit.Wait()
	for _, v := range clients {
		v.Shutdown()
	}
}

func TestSendPurchase(t *testing.T) {
	log2.CodeLoc = true
	// log2.SetLogLevel(log2.Trace)
	const nTotal = 6
	var clients [nTotal]*Client
	var nodes [nTotal]*node.Node
	var transports [nTotal]ifc.Transport
	var e error
	for i := range transports {
		transports[i] = transport.NewSim(nTotal)
	}
	for i := range nodes {
		var hdrPrv *prv.Key
		if hdrPrv, e = prv.GenerateKey(); check(e) {
			t.Error(e)
			t.FailNow()
		}
		hdrPub := pub.Derive(hdrPrv)
		addr := slice.GenerateRandomAddrPortIPv4()
		nodes[i], _ = node.New(addr, hdrPub, hdrPrv, transports[i])
		if clients[i], e = New(transports[i], hdrPrv, nodes[i], nil); check(e) {
			t.Error(e)
			t.FailNow()
		}
		clients[i].AddrPort = nodes[i].AddrPort
	}
	// add each node to each other's Nodes except itself.
	for i := range nodes {
		for j := range nodes {
			if i == j {
				continue
			}
			clients[i].Nodes = append(clients[i].Nodes, nodes[j])
		}
	}
	var ks *signer.KeySet
	if _, ks, e = signer.New(); check(e) {
		t.Error(e)
		t.FailNow()
	}
	var sess [3]*session.Session
	for i := range sess {
		sess[i] = session.NewSession(nonce.NewID(), 203230230,
			time.Hour, ks)
	}
	clients[4].ReceiveCache.Add(address.NewReceiver(sess[0].HeaderPrv))
	clients[5].ReceiveCache.Add(address.NewReceiver(sess[1].HeaderPrv))
	clients[0].ReceiveCache.Add(address.NewReceiver(sess[2].HeaderPrv))
	clients[4].Sessions = clients[4].Sessions.Add(sess[0])
	clients[5].Sessions = clients[5].Sessions.Add(sess[1])
	clients[0].Sessions = clients[0].Sessions.Add(sess[2])

	// Start up the clients.
	for _, v := range clients {
		go v.Start()
	}
	var hop [nTotal - 1]*node.Node
	for i := range clients[0].Nodes {
		hop[i] = clients[0].Nodes[i]
	}
	const nBytes = 2342342
	id := nonce.NewID()
	os := wire.SendPurchase(id, nBytes, clients[0].Node, hop, sess, ks)
	clients[0].PendingSessions = append(clients[0].PendingSessions, id)
	o := os.Assemble()
	b := wire.EncodeOnion(o)
	hop[0].Send(b)
	go func() {
		time.Sleep(time.Second * 2)
		clients[0].Q()
		t.Error("sendpurchase got stuck")
	}()
	<-clients[0].Wait()
	for _, v := range clients {
		v.Shutdown()
	}
}

func TestSendExit(t *testing.T) {
	log2.CodeLoc = true
	// log2.SetLogLevel(log2.Trace)
	const nTotal = 6
	var clients [nTotal]*Client
	var nodes [nTotal]*node.Node
	var transports [nTotal]ifc.Transport
	var e error
	for i := range transports {
		transports[i] = transport.NewSim(nTotal)
	}
	for i := range nodes {
		var hdrPrv *prv.Key
		if hdrPrv, e = prv.GenerateKey(); check(e) {
			t.Error(e)
			t.FailNow()
		}
		hdrPub := pub.Derive(hdrPrv)
		addr := slice.GenerateRandomAddrPortIPv4()
		nodes[i], _ = node.New(addr, hdrPub, hdrPrv, transports[i])
		if clients[i], e = New(transports[i], hdrPrv, nodes[i], nil); check(e) {
			t.Error(e)
			t.FailNow()
		}
		clients[i].AddrPort = nodes[i].AddrPort
	}
	// add each node to each other's Nodes except itself.
	for i := range nodes {
		for j := range nodes {
			if i == j {
				continue
			}
			clients[i].Nodes = append(clients[i].Nodes, nodes[j])
		}
	}
	var ks *signer.KeySet
	if _, ks, e = signer.New(); check(e) {
		t.Error(e)
		t.FailNow()
	}
	var sess [3]*session.Session
	for i := range sess {
		sess[i] = session.NewSession(nonce.NewID(), 203230230,
			time.Hour, ks)
	}
	clients[4].ReceiveCache.Add(address.NewReceiver(sess[0].HeaderPrv))
	clients[5].ReceiveCache.Add(address.NewReceiver(sess[1].HeaderPrv))
	clients[0].ReceiveCache.Add(address.NewReceiver(sess[2].HeaderPrv))
	clients[4].Sessions = clients[4].Sessions.Add(sess[0])
	clients[5].Sessions = clients[5].Sessions.Add(sess[1])
	clients[0].Sessions = clients[0].Sessions.Add(sess[2])
	// set up forwarding port service
	const port = 3455
	clients[3].Services = append(clients[3].Services, &node.Service{
		Port:      port,
		Transport: transport.NewSim(0),
	})
	// Start up the clients.
	for _, v := range clients {
		go v.Start()
	}
	var hop [nTotal - 1]*node.Node
	for i := range clients[0].Nodes {
		hop[i] = clients[0].Nodes[i]
	}
	// id := nonce.NewID()
	var message slice.Bytes
	var hash sha256.Hash
	if message, hash, e = testutils.GenerateTestMessage(32); check(e) {
		t.Error(e)
		t.FailNow()
	}
	quit := qu.T()
	// log.I.S(hash, message.ToBytes())
	os := wire.SendExit(message, port, clients[0].Node, hop, sess, ks)
	clients[0].ExitHooks = clients[0].ExitHooks.Add(hash, func() {
		log.I.S("finished")
		quit.Q()
	})
	// clients[0].PendingSessions = append(clients[0].PendingSessions, id)
	o := os.Assemble()
	b := wire.EncodeOnion(o)
	hop[0].Send(b)
	go func() {
		time.Sleep(time.Second * 6)
		quit.Q()
		t.Error("SendExit got stuck")
	}()
	bb := <-clients[3].Services[0].Receive()
	log.I.S(bb.ToBytes())
	if e = clients[3].SendTo(port, bb); check(e) {
		t.Error("fail send")
	}
	log.I.S("response sent")
	<-quit.Wait()
	for _, v := range clients {
		v.Shutdown()
	}
}

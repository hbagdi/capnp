package main

import (
	"crypto/sha1"
	"fmt"
	"hash"
	"net"

	"context"
	"github.com/hbagdi/capnp/rpc/hashes"
	"zombiezen.com/go/capnproto2/rpc"
)

func server(c net.Conn, message string) error {
	main := hashes.HashFactory_ServerToClient(hashFactory{})
	conn := rpc.NewConn(rpc.StreamTransport(c), rpc.MainInterface(main.Client))

	ctx := context.Background()
	hf := hashes.HashFactory{Client: conn.Bootstrap(ctx)}

	err := rpcToOtherSide(hf, message)
	if err != nil {
		fmt.Println("server err:", err)
	}

	// Wait for connection to abort.
	//err = conn.Wait()
	return err
}

func rpcToOtherSide(hf hashes.HashFactory, message string) error {
	ctx := context.Background()
	// Now we can call methods on hf, and they will be sent over c.
	s := hf.NewSha1(ctx, func(p hashes.HashFactory_newSha1_Params) error {
		return nil
	}).Hash()
	// s refers to a remote Hash.  Method calls are delivered in order.
	s.Write(ctx, func(p hashes.Hash_write_Params) error {
		err := p.SetData([]byte("Hello, "))
		return err
	})
	s.Write(ctx, func(p hashes.Hash_write_Params) error {
		err := p.SetData([]byte("World!"))
		return err
	})
	// Get the sum, waiting for the result.
	result, err := s.Sum(ctx, func(p hashes.Hash_sum_Params) error {
		return nil
	}).Struct()
	if err != nil {
		return err
	}

	// Display the result.
	sha1Val, err := result.Hash()
	if err != nil {
		return err
	}
	fmt.Printf("%s: %x\n", message, sha1Val)
	return nil
}

func main() {
	c1, c2 := net.Pipe()
	go server(c1, "server calling client")
	server(c2, "client calling server")
}

// hashFactory is a local implementation of HashFactory.
type hashFactory struct{}

func (hf hashFactory) NewSha1(call hashes.HashFactory_newSha1) error {
	// Create a new locally implemented Hash capability.
	hs := hashes.Hash_ServerToClient(hashServer{sha1.New()})
	// Notice that methods can return other interfaces.
	return call.Results.SetHash(hs)
}

// hashServer is a local implementation of Hash.
type hashServer struct {
	h hash.Hash
}

func (hs hashServer) Write(call hashes.Hash_write) error {
	data, err := call.Params.Data()
	if err != nil {
		return err
	}
	_, err = hs.h.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func (hs hashServer) Sum(call hashes.Hash_sum) error {
	s := hs.h.Sum(nil)
	return call.Results.SetHash(s)
}

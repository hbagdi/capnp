package main

import (
	"fmt"
	"os"

	"github.com/hbagdi/capnp/serialization/books"
	"zombiezen.com/go/capnproto2"
)

func main() {
	// Read the message from stdin.
	msg, err := capnp.NewDecoder(os.Stdin).Decode()
	if err != nil {
		panic(err)
	}

	// Extract the root struct from the message.
	book, err := books.ReadRootBook(msg)
	if err != nil {
		panic(err)
	}

	// Access fields from the struct.
	title, err := book.Title()
	if err != nil {
		panic(err)
	}
	pageCount := book.PageCount()
	fmt.Printf("%q has %d pages\n", title, pageCount)
}

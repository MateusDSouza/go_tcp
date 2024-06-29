package main

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

type FileServer struct{}

// This function is responsible to create the server that will receive a file. We use the TCP protocol
// The server will operate in the port 3000
// In case of error, we have the Logging for it as a default procedure
// We let the server running in a read loop, so he is always listening in the case bytes are sended
func (fs *FileServer) start() {
	ln, err := net.Listen("tcp", ":3000")
	if err != nil {
		log.Fatal(err)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go fs.readLoop(conn)
	}

}

// In this readLoop we have the Server with a buffer of 2048 bytes.
// Once it receives a file, it creates the buffer as an array of bytes
// As we still dont know the size of the file, we use the variable N for this

func (fs *FileServer) readLoop(conn net.Conn) {
	buf := new(bytes.Buffer)
	for {
		var size int64
		binary.Read(conn, binary.LittleEndian, &size)
		n, err := io.CopyN(buf, conn, size)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(buf.Bytes())
		fmt.Printf("%d bytes received over the network\n", n)
	}
}

// This function has the goal of mimic the file sender.
// It first do a dial to confirm the connection to the server and then, send the file.
// The send of the file is done by writting it to the server in the giver port.
func sendFile(size int) error {
	file := make([]byte, size)
	_, err := io.ReadFull(rand.Reader, file)
	if err != nil {
		return err
	}

	conn, err := net.Dial("tcp", ":3000")
	if err != nil {
		return err
	}
	binary.Write(conn, binary.LittleEndian, int64(size))
	n, err := io.CopyN(conn, bytes.NewReader(file), int64(size))
	if err != nil {
		return err
	}
	fmt.Printf("%d bytes sended\n", n)
	return nil
}

// In the main function we test our cases.
func main() {
	// First case, send a bigger file than the buffer of the server
	// We can see that the file is divided into two chuncks, one with the maximum size of the buffer
	// and another with the rest of the file
	go func() {
		time.Sleep(4 * time.Second)
		sendFile(50000)
	}()
	server := FileServer{}
	server.start()
	// This experiment shows that it is very hard on the server side to handle this size of files.
	// Especially when we are dealing with multiple sources sending different files and we have a limit memory server.
	// The best alternative for this is use a streaming method of receiving files, so that we can have different sources
	// sending file simultaneously. The idea is to split the memory of the server in different parts in order to receive
	//stream of data.
}

// To allow the sender to send bytes, we can use the function NewReader from bytes, so we can send the bytes,
// chunck by chunk
// By changing the sender and reader with the io.Copy, the reading keeps going until reach an end of file.
// In this test, however we do not have an end of file, so the reader keep an open communication waiting for the end of file.
//In order to avoid this problem we can use the io.copyn, so instead of waiting for an end of file, we can say the server to read until the last byte.

// Ok, the copyN can solve our problem. But, then a new problem appear: How can I know how many bytes I should read?????? By the server side....
// Well, our communication between the sender and the receiver, now should have the size parameter.
// The size parameter it is a wait to comunicate in the network how many bytes it should expect before start the communication.

// We use the binary read and write, to communicate with the server before start streaming the file.
// We pass the size of the file, and only after it receives the size, it starts the streaming until the last byte.

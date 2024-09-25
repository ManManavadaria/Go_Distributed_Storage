package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/ManManavadaria/Go_Distributed_Storage/p2p"
)

type FileServer struct {
	FileServerOpts

	mu    sync.Mutex
	peers map[string]p2p.Peer
	Store
	QuitCh chan struct{}
}
type FileServerOpts struct {
	ListenAddr        string
	StorageRoot       string
	PathTransformFunc PathTransformFunc
	Transport         p2p.Transport
	TCPTransportOpts  p2p.TCPTransportOpts
	BootstrapedNodes  []string
}

func NewFileServer(opts FileServerOpts) *FileServer {
	root := opts.StorageRoot[1:]

	storeOpts := &StoreOpts{
		root,
		opts.PathTransformFunc,
	}

	return &FileServer{
		FileServerOpts: opts,
		Store:          *NewStore(storeOpts),
		QuitCh:         make(chan struct{}),
		peers:          make(map[string]p2p.Peer),
		mu:             sync.Mutex{},
	}
}

type Message struct {
	From    string
	Payload any
}
type MessageStoreFile struct {
	Key string
}

type DataMessage struct {
	Key  string
	Data []byte
}

func (s *FileServer) BroadCast(msg *Message) error {
	peers := []io.Writer{}

	for _, peer := range s.peers {
		peers = append(peers, peer)
	}

	mw := io.MultiWriter(peers...)

	// Return an error if gob encoding fails
	if err := gob.NewEncoder(mw).Encode(msg); err != nil {
		return fmt.Errorf("broadcast error: %w", err)
	}
	return nil
}

func (s *FileServer) StoreData(key string, r io.Reader) error {
	msg := &Message{
		Payload: MessageStoreFile{
			Key: key,
		},
	}

	buf := new(bytes.Buffer)

	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
		log.Fatal(err)
		return err
	}

	for _, peer := range s.peers {
		if err := peer.Send(buf.Bytes()); err != nil {
			return err
		}
	}

	time.Sleep(3 * time.Second)

	payload := []byte("this large file")
	for _, peer := range s.peers {
		if err := peer.Send(payload); err != nil {
			return err
		}
	}

	return nil
}

func (fs *FileServer) Start() error {
	fs.Transport.ListenAndAccept()

	fs.bootStarpNetwork()
	fs.Loop()

	return nil
}

func (f *FileServer) bootStarpNetwork() error {
	for _, addr := range f.BootstrapedNodes {
		if len(f.BootstrapedNodes) == 0 {
			continue
		}

		go func(addr string) {
			if err := f.Transport.Dial(addr); err != nil {
				log.Println(err)
			}
		}(addr)
	}
	return nil
}

func (f *FileServer) OnPeer(p p2p.Peer) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.peers[p.RemoteAddr().String()] = p

	log.Println("Connected with peer : ", p.RemoteAddr())
	return nil
}

func (f *FileServer) Loop() {

	defer func() {
		fmt.Println("File server stopped")
		f.Transport.Close()
	}()

	for {
		select {
		case rpc := <-f.Transport.Consume():
			var message Message
			if err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&message); err != nil {
				fmt.Println("decode error:", err)
			}

			if err := f.handleMessage(rpc.From.String(), &message); err != nil {
				log.Println("Error : ", err)
				return
			}

		case <-f.QuitCh:
			return
		}
	}
}

func (f *FileServer) handleMessage(from string, msg *Message) error {
	switch v := msg.Payload.(type) {
	case MessageStoreFile:
		return f.handleMessageStoreFile(from, v)
	}

	return nil
}

func (f *FileServer) handleMessageStoreFile(from string, msg MessageStoreFile) error {
	peer, ok := f.peers[from]
	if !ok {
		return fmt.Errorf("Peer (%s) could not be found in the peer map", from)
	}

	if err := f.Store.Write(msg.Key, peer); err != nil {
		log.Println(err)
		return err
	}
	peer.(*p2p.TCPPeer).Wg.Done()
	return nil
}

func (f *FileServer) Stop() {
	close(f.QuitCh)
	return
}

func init() {
	gob.Register(Message{})
	gob.Register(DataMessage{})
	gob.Register(MessageStoreFile{})
}

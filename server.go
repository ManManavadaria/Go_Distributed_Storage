package main

import (
	"fmt"
	"log"
	"sync"

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
	storeOpts := &StoreOpts{
		opts.StorageRoot,
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
		case msg := <-f.Transport.Consume():
			fmt.Println(msg)

		case <-f.QuitCh:
			return
		}
	}
}

func (f *FileServer) Stop() {
	close(f.QuitCh)
	return
}

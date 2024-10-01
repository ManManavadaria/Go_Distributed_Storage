package main

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/ManManavadaria/Go_Distributed_Storage/p2p"
)

var (
	Wg = sync.WaitGroup{}
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
	EncKey            []byte
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
	Key  string
	Size int
}

type DataMessage struct {
	Key  string
	Data []byte
}

func (s *FileServer) stream(msg *Message) error {
	peers := []io.Writer{}

	for _, peer := range s.peers {
		peers = append(peers, peer)
	}

	mw := io.MultiWriter(peers...)

	if err := gob.NewEncoder(mw).Encode(msg); err != nil {
		return fmt.Errorf("broadcast error: %w", err)
	}
	return nil
}

func (s *FileServer) broadCast(msg *Message) error {
	msgBuf := new(bytes.Buffer)

	if err := gob.NewEncoder(msgBuf).Encode(msg); err != nil {
		log.Fatal(err)
		return err
	}

	for _, peer := range s.peers {
		peer.Send([]byte{p2p.IncomingMessage})
		if err := peer.Send(msgBuf.Bytes()); err != nil {
			return err
		}
	}
	return nil
}

type MessageGetFile struct {
	Key string
}

func (s *FileServer) Get(key string) (int64, io.Reader, error) {
	if s.Store.Has(key) {
		fmt.Printf("[%s] Serving file (%s) from the local disk\n", s.Transport.Addr(), key)
		return s.Store.Read(key)
	} else {
		fmt.Printf("[%s] Doesn't exist file (%s) locally, Fetching from the network...\n", s.Transport.Addr(), key)
	}

	msg := Message{
		Payload: MessageGetFile{
			Key: hashKey(key),
		},
	}

	if err := s.broadCast(&msg); err != nil {
		return 0, nil, err
	}

	time.Sleep(time.Millisecond * 500)

	for _, peer := range s.peers {
		var fileSize int64
		binary.Read(peer, binary.LittleEndian, &fileSize)

		_, err := s.Store.writeDecrypt(s.EncKey, key, io.LimitReader(peer, fileSize))
		if err != nil {
			return 0, nil, err
		}

	}
	return s.Store.Read(key)
}

func (s *FileServer) store(key string, r io.Reader) error {

	var (
		fileBuffer = new(bytes.Buffer)
		tee        = io.TeeReader(r, fileBuffer)
	)

	size, err := s.Store.Write(key, tee)
	if err != nil {
		return err
	}

	msg := &Message{
		Payload: MessageStoreFile{
			Key:  hashKey(key),
			Size: int(size) + 16,
		},
	}

	if err := s.broadCast(msg); err != nil {
		return err
	}

	time.Sleep(3 * time.Second)

	peers := []io.Writer{}
	for _, peer := range s.peers {
		peers = append(peers, peer)
	}
	mw := io.MultiWriter(peers...)
	mw.Write([]byte{p2p.IncomingStream})
	n, err := copyEncrypt(s.EncKey, fileBuffer, mw)
	if err != nil {
		return err
	}

	fmt.Printf("[%s] Received and written (%d) bytes to disk\n", s.Transport.Addr(), n)

	return nil
}

type MessageRemoveFile struct {
	Key string
}

func (s *FileServer) Remove(key string) error {
	s.Store.Delete(key)

	fmt.Printf("[%v] File (%s) removed from the local disk \n", s.Transport.Addr(), key)

	msg := Message{
		Payload: MessageRemoveFile{
			Key: hashKey(key),
		},
	}

	Wg.Add(len(s.peers))
	if err := s.broadCast(&msg); err != nil {
		return err
	}

	Wg.Wait()
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

// Loop is the main loop of the server which listens for incoming messages and handles them
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
				log.Println("handle message error  : ", err)
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

	case MessageGetFile:
		return f.handleMessageGetFile(from, v)

	case MessageRemoveFile:
		return f.handleMessageRemoveFile(from, v)
	}
	return nil
}

func (f *FileServer) handleMessageGetFile(from string, msg MessageGetFile) error {
	if !f.Store.Has(msg.Key) {
		return fmt.Errorf("[%s] file serving request of (%s) but does'nt exist on disk \n", f.Transport.Addr(), msg.Key)
	}

	fmt.Printf("[%s] serving file (%s) over the network\n", f.Transport.Addr(), msg.Key)

	fileSize, r, err := f.Store.Read(msg.Key)
	if err != nil {
		return err
	}

	if rc, ok := r.(io.ReadCloser); ok {
		defer rc.Close()
	}

	peer, ok := f.peers[from]
	if !ok {
		return fmt.Errorf("Peer (%s) is not in map", from)
	}

	peer.Send([]byte{p2p.IncomingStream})
	binary.Write(peer, binary.LittleEndian, fileSize)
	n, err := io.Copy(peer, r)
	if err != nil {
		return err
	}

	fmt.Printf("[%s] Written %d bytes over the network to %s\n", f.Transport.Addr(), n, from)
	return nil
}

func (f *FileServer) handleMessageStoreFile(from string, msg MessageStoreFile) error {
	peer, ok := f.peers[from]
	if !ok {
		log.Fatal(fmt.Errorf("Peer (%s) could not be found in the peer map", from))
		return fmt.Errorf("Peer (%s) could not be found in the peer map", from)
	}

	fmt.Println("writing file to peer ===> ", f.Transport.Addr())
	n, err := f.Store.Write(msg.Key, io.LimitReader(peer, int64(msg.Size)))
	if err != nil {
		log.Fatal(err)
		return err
	}

	fmt.Printf("[%s] Written %v bytes to disk\n", f.Transport.Addr(), n)

	peer.CloseStream()
	return nil
}

func (f *FileServer) handleMessageRemoveFile(from string, msg MessageRemoveFile) error {
	peer, ok := f.peers[from]
	if !ok {
		return fmt.Errorf("Peer (%s) could not be found in the peer map\n", from)
	}

	err := f.Store.Delete(msg.Key)
	if err != nil {
		log.Println(err)
		return err
	}
	Wg.Done()
	fmt.Printf("[%v] Removed (%v) file from the network storage\n", peer.LocalAddr(), msg.Key)

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
	gob.Register(MessageGetFile{})
	gob.Register(MessageRemoveFile{})
}

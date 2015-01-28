package impl

import (
	"encoding/hex"
	"fmt"
	"github.com/jbenet/go-ipfs/Godeps/_workspace/src/code.google.com/p/go.net/context"
	mh "github.com/jbenet/go-ipfs/Godeps/_workspace/src/github.com/jbenet/go-multihash"
	"github.com/jbenet/go-ipfs/blocks"
	"github.com/jbenet/go-ipfs/core"
	"github.com/jbenet/go-ipfs/core/coreunix"
	// "github.com/jbenet/go-ipfs/core/corerepo"
	mdag "github.com/jbenet/go-ipfs/merkledag"
	// uio "github.com/jbenet/go-ipfs/unixfs/io"
	fsrepo "github.com/jbenet/go-ipfs/repo/fsrepo"
	"github.com/jbenet/go-ipfs/util"
	"github.com/jbenet/go-ipfs/repo/config"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

type IMap map[string]interface{}

var StreamSize = 1024

type Ipfs struct {
	node *core.IpfsNode
	cfg  *config.Config
	root string
}

func NewIpfs() *Ipfs{
	return &Ipfs{}
}

// TODO: UDP socket won't close
// https://github.com/jbenet/go-ipfs/issues/389
func (ipfs *Ipfs) Init(rootDir string) error {
	
	var cfg *config.Config 
	var err error
	
	cfg, err = fsrepo.ConfigAt(rootDir)
	
	if err != nil {
		fmt.Println("Repo not found, initializing")
		cfg, err = initWithDefaults(rootDir)
		if err != nil {
			return err
		}
	}
	
	ipfs.cfg = cfg
	ipfs.root = rootDir
	
	// TODO add settings later.
	util.SetLogLevel("*", "debug")
	fmt.Println("IPFS: init done")
	return nil
}

// NOTE: Init is in the init file
func (ipfs *Ipfs) Start() error {
	ctx := context.Background()
	r := fsrepo.At(ipfs.root)
	r.Open()
	n, err := core.NewIPFSNode(ctx, core.Online(r))
	if err != nil {
		return err
	}
	ipfs.node = n
	return nil
}

// TODO Implement
func (ipfs *Ipfs) Shutdown() error {
	
	return nil
}

func (ipfs *Ipfs) GetBlock(hash string) ([]byte, error) {
	h, err := hex.DecodeString(hash)
	if err != nil {
		return nil, err
	}
	k := util.Key(h)
	// TODO consider this
	ctx, _ := context.WithTimeout(context.TODO(), time.Second*5)
	fmt.Printf("IPFS STUFF: node: %v\n", ipfs.node)
	fmt.Printf("IPFS STUFF: Blocks: %v\n", ipfs.node.Blocks)
	b, err := ipfs.node.Blocks.GetBlock(ctx, k)
	if err != nil {
		return nil, fmt.Errorf("block get: %v", err)
	}
	return b.Data, nil
}

func (ipfs *Ipfs) GetFile(hash string) ([]byte, error) {
	h, err := hexPath2B58(hash)
	if err != nil {
		return nil, err
	}
	// buf := bytes.NewBuffer(nil)
	reader, err := coreunix.Cat(ipfs.node, h) //cmds.Cat(ipfs.node, []string{h}, nil, buf)
	if err != nil {
		return nil, err
	}
	
	return ioutil.ReadAll(reader) 
}

/*
func (ipfs *Ipfs) GetStream(hash string) (chan []byte, error) {
	fpath, err := hexPath2B58(hash)
	if err != nil {
		return nil, err
	}
	dagnode, err := ipfs.node.Resolver.ResolvePath(fpath)
	if err != nil {
		return nil, fmt.Errorf("catFile error: %v", err)
	}
	read, err := uio.NewDagReader(dagnode, ipfs.node.DAG)
	if err != nil {
		return nil, fmt.Errorf("cat error: %v", err)
	}
	ch := make(chan []byte)
	var n int
	go func() {
		for err != io.EOF {
			b := make([]byte, StreamSize)
			// read from reader 1024 bytes at a time
			n, err = read.Read(b)
			if err != nil && err != io.EOF {
				//return nil, err
				break
				// how to handle these errors?!
			}
			// broadcast on channel
			ch <- b[:n]
		}
		close(ch)
	}()
	return ch, nil
}
*/

// TODO: depth
func (ipfs *Ipfs) GetTree(hash string, depth int) (IMap, error) {
	fpath, err := hexPath2B58(hash)
	if err != nil {
		return nil, err
	}
	nd, err1 := ipfs.node.Resolver.ResolvePath(fpath)
	if err1 != nil {
		return nil, err1
	}
	mhash, err2 := nd.Multihash()
	if err2 != nil {
		return nil, err2
	}
	tree := getTreeNode("", hex.EncodeToString(mhash))
	err3 := grabRefs(ipfs.node, nd, tree)
	return tree, err3
}

func (ipfs *Ipfs) AddBlock(data []byte) (string, error) {
	b := blocks.NewBlock(data)
	k, err := ipfs.node.Blocks.AddBlock(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString([]byte(k)), nil
}

func (ipfs *Ipfs) AddBlockString(data string) (string, error) {
	return ipfs.AddBlock([]byte(data))
}

func (ipfs *Ipfs) AddFile(fpath string) (string, error) {
	file, err := os.Open(fpath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	k, err := coreunix.Add(ipfs.node,file)
	return "0x" + hex.EncodeToString([]byte(k)), err
}

func (ipfs *Ipfs) AddTree(fpath string, depth int) (string, error) {
	/*
	ff, err := os.Open(fpath)
	if err != nil {
		return "", err
	}
	f, err := openPath(ff, fpath)
	if err != nil {
		return "", err
	}
	added := &cmds.AddOutput{}
	nd, err := addDir(ipfs.node, f, added)
	if err != nil {
		return "", err
	}
	h, err := nd.Multihash()
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(h), nil
	*/
	return coreunix.AddR(ipfs.node,fpath);
}

/*
// Key manager functions.
// Note in ipfs (in contrast with a blockchain), one is much less likely
// to change keys, as there are accrued benefits to sticking with a single key,
// and there is no notion of "transactions"
// An ipfs ID is simply the multihash of the publickey
func (ipfs *Ipfs) ActiveAddress() string {
	return hex.EncodeToString(ipfs.node.Identity.ID())
}

// Ipfs node's only have one address
func (ipfs *Ipfs) Address(n int) (string, error) {
	return ipfs.ActiveAddress(), nil
}
func (ipfs *Ipfs) SetAddress(addr string) error {
	return fmt.Errorf("It is not possible to set the ipfs node address without restarting.")
}
func (ipfs *Ipfs) SetAddressN(n int) error {
	return fmt.Errorf("It is not possible to set the ipfs node address without restarting.")
}

// We don't create new addresses on the fly
func (ipfs *Ipfs) NewAddress(set bool) (string, error) {
	return "", fmt.Errorf("It is not possible to create new addresses during runtime.")
}

// we only have one ipfs address
func (ipfs *Ipfs) AddressCount() int {
	return 1
}
*/

func HexToB58(s string) (string, error) {
	var b []byte
	if len(s) > 2 {
		if s[:2] == "0x" {
			s = s[2:]
		}
	}
	b, err := hex.DecodeString(s)
	if err != nil {
		return "", err
	}
	bmh := mh.Multihash(b)
	return bmh.B58String(), nil //b58.Encode(b), nil
}

// should this return 0x prefixed?
func B58ToHex(s string) (string, error) {
	r, err := mh.FromB58String(s) //b58.Decode(s) with panic recovery
	if err != nil {
		return "", err
	}
	h := hex.EncodeToString(r)
	return "0x" + h, nil
}

// convert path beginning with 32 byte hex string to path beginning with base58 encoded
func hexPath2B58(p string) (string, error) {
	var err error
	p = strings.TrimLeft(p, "/") // trim leading slash
	spl := strings.Split(p, "/") // split path
	leadingHash := spl[0]
	spl[0], err = HexToB58(leadingHash) // convert leading hash to base58
	if err != nil {
		return "", err
	}
	if len(spl) > 1 {
		return strings.Join(spl, "/"), nil
	}
	return spl[0], nil
}
func getTreeNode(name, hash string) IMap {
	obj := make(IMap)
	obj["Nodes"] = make([]IMap, 0)
	obj["Name"] = name
	obj["Hash"] = hash
	return obj
}
func grabRefs(n *core.IpfsNode, nd *mdag.Node, tree IMap) error {
	for _, link := range nd.Links {
		h := link.Hash
		newNode := getTreeNode(link.Name, h.B58String())
		nd, err := n.DAG.Get(util.Key(h))
		if err != nil {
			//log.Errorf("error: cannot retrieve %s (%s)", h.B58String(), err)
			return err
		}
		err = grabRefs(n, nd, newNode)
		if err != nil {
			return err
		}
		nds := tree["Nodes"].([]IMap)
		nds = append(nds, newNode)
	}
	return nil
}

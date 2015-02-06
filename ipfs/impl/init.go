package impl


import (
	"encoding/base64"
	"fmt"
	"os"
	context "github.com/jbenet/go-ipfs/Godeps/_workspace/src/code.google.com/p/go.net/context"
	core "github.com/jbenet/go-ipfs/core"
	ipns "github.com/jbenet/go-ipfs/fuse/ipns"
	ci "github.com/jbenet/go-ipfs/p2p/crypto"
	peer "github.com/jbenet/go-ipfs/p2p/peer"
	config "github.com/jbenet/go-ipfs/repo/config"
	fsrepo "github.com/jbenet/go-ipfs/repo/fsrepo"
	debugerror "github.com/jbenet/go-ipfs/util/debugerror"
)

const nBitsForKeypairDefault = 4096
/*
var ErisPeerServer config.BootstrapPeer = config.BootstrapPeer{
	PeerID: "QmNexwW6SdVgwEdk9XYTfPHph4EEgmcrD4wNsnwFj9YqnS",
	Address: "/ip4/92.243.15.73/tcp/4001",
}
*/

func initWithDefaults(repoRoot string) (*config.Config, error) {
	return doInit(repoRoot, false, nBitsForKeypairDefault)
}

func doInit(repoRoot string, force bool, nBitsForKeypair int) (*config.Config, error) {

	fmt.Printf("initializing ipfs node at %s\n", repoRoot)

	conf, err := config.Init(os.Stdout, nBitsForKeypair)
	if err != nil {
		return nil, err
	}

	if fsrepo.IsInitialized(repoRoot) && !force {
		fmt.Println("Repo already exists")
		return nil, nil
	}

	if err := fsrepo.Init(repoRoot, conf); err != nil {
		return nil, err
	}
	
	if fsrepo.IsInitialized(repoRoot) {
		if err := fsrepo.Remove(repoRoot); err != nil {
			return nil, err
		}
	}
	
	if err := fsrepo.Init(repoRoot, conf); err != nil {
		return nil, err
	}

	err = initializeIpnsKeyspace(repoRoot)
	if err != nil {
		return nil, err
	}
	
	return conf, nil
}

func initializeIpnsKeyspace(repoRoot string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	r := fsrepo.At(repoRoot)
	if err := r.Open(); err != nil { // NB: repo is owned by the node
		return err
	}

	nd, err := core.NewIPFSNode(ctx, core.Offline(r))
	if err != nil {
		return err
	}
	defer nd.Close()

	err = nd.SetupOfflineRouting()
	if err != nil {
		return err
	}

	return ipns.InitializeKeyspace(nd, nd.PrivateKey)
}

func datastoreConfig() (*config.Datastore, error) {
	dspath, err := config.DataStorePath("")
	if err != nil {
		return nil, err
	}
	return &config.Datastore{
		Path: dspath,
		Type: "leveldb",
	}, nil
}

// identityConfig initializes a new identity.
func identityConfig(nbits int) (config.Identity, error) {
	// TODO guard higher up
	ident := config.Identity{}
	if nbits < 1024 {
		return ident, debugerror.New("Bitsize less than 1024 is considered unsafe.")
	}

	fmt.Printf("generating %v-bit RSA keypair...", nbits)
	sk, pk, err := ci.GenerateKeyPair(ci.RSA, nbits)
	if err != nil {
		return ident, err
	}
	fmt.Printf("done\n")

	// currently storing key unencrypted. in the future we need to encrypt it.
	// TODO(security)
	skbytes, err := sk.Bytes()
	
	if err != nil {
		return ident, err
	}
	
	ident.PrivKey = base64.StdEncoding.EncodeToString(skbytes)

	id, err := peer.IDFromPublicKey(pk)
	if err != nil {
		return ident, err
	}
	ident.PeerID = id.Pretty()
	fmt.Printf("peer identity: %s\n", ident.PeerID)
	return ident, nil
}
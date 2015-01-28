package ipfs

import (
	"fmt"
	"github.com/eris-ltd/decerver-modules/ipfs/impl"
	"github.com/eris-ltd/modules/types"
	"github.com/eris-ltd/decerver/interfaces/files"
	"github.com/eris-ltd/decerver/interfaces/modules"
	"github.com/eris-ltd/decerver/interfaces/scripting"
	"io/ioutil"
	"path"
)

// implements decerver-interface module.
type (
	// This is the module.
	IpfsModule struct {
		ipfs    *impl.Ipfs
		ipfsApi *IpfsApi
		config  *IpfsDecerverConfig
		fileIO  files.FileIO
	}

	// This is the api.
	IpfsApi struct {
		ipfs *impl.Ipfs
	}
)

// This is the configuration file Decerver uses.
type IpfsDecerverConfig struct {
	RootDir string `json:"root_directory"`
}

func getDefaultConfig(rootDir string) *IpfsDecerverConfig {
	return &IpfsDecerverConfig{rootDir}
}

func NewIpfsModule() *IpfsModule {
	ipfs := impl.NewIpfs()
	mod := &IpfsModule{}
	mod.ipfs = ipfs
	mod.ipfsApi = &IpfsApi{ipfs}
	return mod
}

func (mod *IpfsModule) Register(mapi modules.DecerverModuleApi) error {
	mod.fileIO = mapi.FileIO()
	mapi.RegisterRuntimeObject("ipfs", mod.ipfsApi)
	return nil
}

func (mod *IpfsModule) Init() error {
	fmt.Println("IPFS: initializing")
	// Now we load (or create) the config file for decerver stuff.
	ipfsModDir := path.Join(mod.fileIO.Modules(), "ipfs")
	// ipfsConf := path.Join(ipfsModDir,"config")
	var err error
	configFile := &IpfsDecerverConfig{}
	err = mod.fileIO.UnmarshalJsonFromFile(ipfsModDir, "config", configFile)
	if err != nil {
		rootDir := path.Join(mod.fileIO.Filesystems(), "ipfs")
		mod.fileIO.CreateDirectory(rootDir)
		fmt.Println("Ipfs: config error - resorting to defaults: " + err.Error())
		configFile = getDefaultConfig(rootDir)
		oErr := mod.fileIO.MarshalJsonToFile(ipfsModDir, "config", configFile)
		if oErr != nil {
			fmt.Println("Config not saved: " + oErr.Error())
		}
	}
	mod.config = configFile
	// Now we go on to load ipfs using the data from the config file.
	return mod.ipfs.Init(mod.config.RootDir)
}

func (mod *IpfsModule) Start() error {
	mod.ipfs.Start()
	return nil
}

func (mod *IpfsModule) Shutdown() error {
	return mod.ipfs.Shutdown()
}

// TODO figure out when this would actually be used.
func (mod *IpfsModule) Restart() error {
	err := mod.Shutdown()
	if err != nil {
		return nil
	}
	return mod.Start()
}

func (mod *IpfsModule) SetProperty(name string, data interface{}) {
}

func (mod *IpfsModule) Property(name string) interface{} {
	return nil
}

func (mod *IpfsModule) Name() string {
	return "ipfs"
}

func (mod *IpfsModule) Subscribe(name string, event string, target string) chan types.Event {
	return nil
}

func (mod *IpfsModule) UnSubscribe(name string) {

}

func (api *IpfsApi) GetBlock(hash string) scripting.SObject {
	data, err := api.ipfs.GetBlock(hash)
	if err != nil {
		return scripting.JsReturnValErr(err)
	}
	return scripting.JsReturnValNoErr(string(data))
}

func (api *IpfsApi) GetFile(hash string) scripting.SObject {
	reader, err := api.ipfs.GetFile(hash)
	if err != nil {
		return scripting.JsReturnValErr(err)
	}
	bts, _ := ioutil.ReadAll(reader)
	return scripting.JsReturnValNoErr(string(bts))
}

func (api *IpfsApi) GetTree(hash string, depth int) scripting.SObject {
	tree, err := api.ipfs.GetTree(hash, depth)
	return scripting.JsReturnVal(tree, err)
}

// Deprecated. Use PushBlock instead.
func (api *IpfsApi) PushBlockString(block string) scripting.SObject {
	fmt.Println("IPFS Module: Use of deprecated function 'PushBlockString'.")
	block, err := api.ipfs.AddBlockString(block)
	return scripting.JsReturnVal(block, err)
}

func (api *IpfsApi) PushBlock(block string) scripting.SObject {
	block, err := api.ipfs.AddBlockString(block)
	return scripting.JsReturnVal(block, err)
}

func (api *IpfsApi) PushFile(fpath string) scripting.SObject {
	f, err := api.ipfs.AddFile(fpath)
	return scripting.JsReturnVal(f, err)
}

func (api *IpfsApi) PushTree(fpath string, depth int) scripting.SObject {
	tree, err := api.ipfs.AddTree(fpath, depth)
	return scripting.JsReturnVal(tree, err)
}

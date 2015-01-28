package main

// This is a temporary "test" file.
import (
	"fmt"
	"log"
	"time"
	//"encoding/hex"
	"github.com/eris-ltd/decerver-modules/ipfs/impl"
	"github.com/jbenet/go-ipfs/util"
	"os/user"
)

func main() {
	
	i := impl.NewIpfs()
	// Root path goes here.
	err := i.Init("/user/androlo/ipfstest")
	if err != nil {
		log.Fatal(err)
	}
	
	err = i.Start()
	
	if err != nil {
		log.Fatal(err)
	}
	
	util.SetLogLevel("*", "warning")
	
	//c := "QmVHdqmE5x55kZaavWUmscLmieusDdZhQBP5mjZHwMB3U9"
	//c := "Qmb8zwr341xu5uUWwxvVKbZs1ZbjJRJJ965tnV9HDeVUkH"
	
	time.Sleep(time.Second * 2)
	c := "QmVq6uMzsKg7x5mDEyLS5p5xiyTQ49LR8kFk1wnFDhodzz"
	
	fmt.Println("Testing block add/get.") 
	
	fmt.Println("c :" + c)
	h, _ := impl.B58ToHex(c)
	
	fmt.Println("c Hash: " + h)
	
	a, _ := i.AddBlock([]byte(h))
	fmt.Println("a Hash: " + "0x" + string(a))
	
	ah, _ := i.GetBlock(a)
	fmt.Println("a holds: " + "0x" + string(ah))
	
	cah, _ := impl.HexToB58(string(ah))
	fmt.Println("a Base58: " + string(cah))
	
	fmt.Printf("a base58 == c: %t\n", string(cah) == c)
	
	afHash, _ := i.AddFile("./testfiles/test0.txt")
	
	fmt.Println("Hash of added file: " + afHash)
	
	afFile, _ := i.GetFile(afHash)
	
	fmt.Println("Content of added file: " + string(afFile))
	/*
	atHash, _ := i.AddTree("./testfiles")
	
	fmt.Println("Hash of added directory: " + atHash)
	
	atFile, _ := i.GetFile(atHash)
	
	fmt.Println("Content of added dir: " + string(atFile))
	*/
	
	err = i.Unpin(afHash)
	if err != nil {
		fmt.Println("Error with unpin: " + err.Error())
	}
	
	err = i.GC()
	if err != nil {
		fmt.Println("Error with gc: " + err.Error())
	}
	
	/*
	g, _ := i.Get("tree", ipfs.B58ToHex("QmaKxiCScMY6BG1eq228F2fDJmjxZ53MJ8MtEyEJZr3v44"))
	t := g.(modules.FsNode)
	printTree(&t)

	   ch, _ := i.Get("stream", h)
	   for r := range ch.(chan []byte){
	       fmt.Println(string(r))
	   }

	   a := hex.EncodeToString([]byte("fuck you"))
	   fmt.Println("#####")
	   k, _ := i.Push("block", a)
	   fmt.Println(k)
	   aa, err := i.Get("block", k)
	   if err != nil{
	       fmt.Println(err)
	   }
	   fmt.Println(string(aa.([]byte)))
	   
	time.Sleep(time.Second * 5)
	fmt.Println("calling get file...")
	j := i.GetFile(h)
	a := j["Data"]
	e := j["Err"]
	fmt.Println(a.(string), e)
	i.Shutdown()
	*/
}

/*
func printTree(t *modules.FsNode) {
	fmt.Println(t.Name)
	for _, tt := range t.Nodes {
		printTree(tt)
	}
}
*/
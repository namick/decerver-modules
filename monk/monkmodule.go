package monkjs

import (
	"fmt"
	"github.com/eris-ltd/decerver/interfaces/modules"
	"github.com/eris-ltd/decerver/interfaces/scripting"
	"github.com/eris-ltd/modules/types"
	"github.com/eris-ltd/thelonious/monk"
)

type TempProps struct {
	ChainId    string
	RemoteHost string
	RemotePort int
	RootDir    string
}

// implements decerver-interfaces Module
type MonkModule struct {
	monk *monk.MonkModule
	temp *TempProps
	mapi *MonkApi
}

type MonkApi struct {
	monk *monk.MonkModule
}

func NewMonkModule() *MonkModule {
	monk := monk.NewMonk(nil)
	mapi := &MonkApi{monk}
	return &MonkModule{
		monk : monk, 
		temp : &TempProps{}, 
		mapi : mapi, 
	}
}

// Register the module.
func (mm *MonkModule) Register(dma modules.DecerverModuleApi) error {
	dma.RegisterRuntimeObject("monk", mm.mapi)
	dma.RegisterRuntimeScript(eslScript)
	return nil
}

// Initialize the module (does nothing for monk)
func (mm *MonkModule) Init() error {
	return nil
}

// Start the module (does nothing for monk)
func (mm *MonkModule) Start() error {
	return nil // mjs.mm.Start()
}

// Shut down the module.
func (mm *MonkModule) Shutdown() error {
	
	return mm.monk.Shutdown()
}

func (mm *MonkModule) Restart() error {
	err := mm.Shutdown()
	
	if err != nil {
		return nil
	}
	
	mm.monk = monk.NewMonk(nil)
	mm.mapi.monk = mm.monk
	
	// Inject the config:
	mm.monk.SetProperty("RootDir", mm.temp.RootDir)
	mm.monk.SetProperty("ChainId", mm.temp.ChainId)
	mm.monk.SetProperty("RemoteHost", mm.temp.RemoteHost)
	mm.monk.SetProperty("RemotePort", mm.temp.RemotePort)
	
	mm.monk.Init()
	
	err2 := mm.monk.Start()
	
	mm.temp.ChainId = ""
	mm.temp.RemoteHost = ""
	mm.temp.RemotePort = 0
	mm.temp.RootDir = ""

	return err2
}

func (mm *MonkModule) SetProperty(name string, data interface{}) {
	if name == "ChainId" {
		dt, dtok := data.(string)
		if !dtok {
			fmt.Println("Setting property 'ChainId' to an undefined value. Should be 'string'.")
			return
		}
		mm.temp.ChainId = dt
	} else if name == "RemoteHost" {
		dt2, dtok2 := data.(string)
		if !dtok2 {
			fmt.Println("Setting property 'RemoteHost' to an undefined value. Should be 'string'.")
			return
		}
		mm.temp.RemoteHost = dt2
	} else if name == "RemotePort" {
		dt3, dtok3 := data.(int)
		if !dtok3 {
			fmt.Println("Setting property 'RemotePort' to an undefined value. Should be 'int'.")
			return
		}
		mm.temp.RemotePort = dt3
	} else if name == "RootDir" {
		dt4, dtok4 := data.(string)
		if !dtok4 {
			fmt.Println("Setting property 'RootDir' to an undefined value. Should be 'string'.")
			return
		}
		mm.temp.RootDir = dt4
	} else {
		fmt.Println("Setting undefined property.")
	}
}

func (mm *MonkModule) Property(name string) interface{} {
	return nil
}

// ReadConfig and WriteConfig implemented in config.go

// What module is this?
func (mm *MonkModule) Name() string {
	return "monk"
}

// TODO update this.
func (mm *MonkModule) Subscribe(name, event, target string) chan types.Event {
	return mm.monk.Subscribe(name, event, target)
}

func (mm *MonkModule) UnSubscribe(name string) {
	mm.monk.UnSubscribe(name)
}

/*
   Wrapper so module satisfies Blockchain
*/

// TODO not many errors is returned by the monk object.

func (mapi *MonkApi) WorldState() scripting.SObject {
	ws := mapi.monk.WorldState()
	return scripting.JsReturnValNoErr(ws)
}

func (mapi *MonkApi) State() scripting.SObject {
	return scripting.JsReturnValNoErr(mapi.monk.State())
}

func (mapi *MonkApi) Storage(target string) scripting.SObject {
	return scripting.JsReturnValNoErr(mapi.monk.Storage(target))
}

func (mapi *MonkApi) Account(target string) scripting.SObject {
	return scripting.JsReturnValNoErr(mapi.monk.Account(target))
}

func (mapi *MonkApi) StorageAt(target, storage string) scripting.SObject {
	ret := mapi.monk.StorageAt(target, storage)
	if ret == "" || ret == "0x" {
		ret = "0x0"
	} else {
		ret = "0x" + ret
	}
	return scripting.JsReturnValNoErr(ret)
}

func (mapi *MonkApi) BlockCount() scripting.SObject {
	return scripting.JsReturnValNoErr(mapi.monk.BlockCount())
}

func (mapi *MonkApi) LatestBlock() scripting.SObject {
	return scripting.JsReturnValNoErr(mapi.monk.LatestBlock())
}

func (mapi *MonkApi) Block(hash string) scripting.SObject {
	return scripting.JsReturnValNoErr(mapi.monk.Block(hash))
}

func (mapi *MonkApi) IsScript(target string) scripting.SObject {
	return scripting.JsReturnValNoErr(mapi.monk.IsScript(target))
}

func (mapi *MonkApi) Tx(addr, amt string) scripting.SObject {
	hash, err := mapi.monk.Tx(addr, amt)
	var ret scripting.SObject
	if err == nil {
		ret = make(scripting.SObject)
		ret["Hash"] = hash
		ret["Address"] = ""
		ret["Error"] = ""
	}
	return scripting.JsReturnVal(ret, err)
}

func (mapi *MonkApi) Msg(addr string, data []interface{}) scripting.SObject {
	fmt.Printf("MESSAGE DATA: %v\n", data)
	indata := make([]string, 0)

	if data != nil && len(data) > 0 {
		for _, d := range data {
			str, ok := d.(string)
			if !ok {
				return scripting.JsReturnValErr(fmt.Errorf("Msg indata is not an array of strings"))
			}
			indata = append(indata, str)
		}
	}
	hash, err := mapi.monk.Msg(addr, indata)
	fmt.Println("HASH: " + hash)
	ret := make(scripting.SObject)
	if err == nil {
		ret["Hash"] = "0x" + hash // Might as well
		ret["Address"] = ""
		ret["Error"] = ""
	} else {
		ret["Hash"] = ""
		ret["Address"] = ""
		ret["Error"] = err.Error()
	}
	return scripting.JsReturnVal(ret, err)
}

/*
func (mapi *MonkApi) Script(file, lang string) scripting.SObject {
	addr, err := mapi.monk.Script(file, lang)
	var ret scripting.SObject
	if err == nil {
		ret = make(scripting.SObject)
		ret["Hash"] = ""
		ret["Address"] = addr
		ret["Error"] = ""
	}
	return scripting.JsReturnVal(ret, err)
}
*/
func (mapi *MonkApi) Commit() scripting.SObject {
	mapi.monk.Commit()
	return scripting.JsReturnValNoErr(nil)
}

func (mapi *MonkApi) AutoCommit(toggle bool) scripting.SObject {
	mapi.monk.AutoCommit(toggle)
	return scripting.JsReturnValNoErr(nil)
}

func (mapi *MonkApi) IsAutocommit() scripting.SObject {
	return scripting.JsReturnValNoErr(mapi.monk.IsAutocommit())
}

/*
   Module should also satisfy KeyManager
*/

func (mapi *MonkApi) ActiveAddress() scripting.SObject {
	return scripting.JsReturnValNoErr(mapi.monk.ActiveAddress())
}

func (mapi *MonkApi) Addresses() scripting.SObject {
	count := mapi.monk.AddressCount()
	addresses := make(scripting.SObject)
	array := make([]string, count)

	for i := 0; i < count; i++ {
		addr, _ := mapi.monk.Address(i)
		array[i] = addr
	}
	addresses["Addresses"] = array
	return scripting.JsReturnValNoErr(addresses)
}

func (mapi *MonkApi) SetAddress(addr string) scripting.SObject {
	err := mapi.monk.SetAddress(addr)
	if err != nil {
		return scripting.JsReturnValErr(err)
	} else {
		// No error means success.
		return scripting.JsReturnValNoErr(nil)
	}
}

// TODO Js runtime returns weird numbers.
func (mapi *MonkApi) SetAddressN(n interface{}) scripting.SObject {
	// TODO Safe conversion
	switch v := n.(type) {
	case int:
		mapi.monk.SetAddressN(v)
	case uint:
		mapi.monk.SetAddressN(int(v))
	case int64:
		mapi.monk.SetAddressN(int(v))
	case uint64:
		mapi.monk.SetAddressN(int(v))
	case float32:
		mapi.monk.SetAddressN(int(v))
	case float64:
		mapi.monk.SetAddressN(int(v))
	default:
		return scripting.JsReturnValErr(fmt.Errorf("Value is not a proper number: %v\n", n))
	}
	return scripting.JsReturnValNoErr(nil)
}

func (mapi *MonkApi) NewAddress(set bool) scripting.SObject {
	return scripting.JsReturnValNoErr(mapi.monk.NewAddress(set))
}

func (mapi *MonkApi) AddressCount() scripting.SObject {
	return scripting.JsReturnValNoErr(mapi.monk.AddressCount())
}

var eslScript string = `

var StdVarOffset = "0x1";

var NSBase = Exp("0x100","31");

var esl = {};

esl.SA = function(acc,addr) {
	return monk.StorageAt(acc,addr).Data;
};

esl.array = {

	//Constants
	"ESizeOffset" : "0",

	"MaxEOffset" : "0",
	"StartOffset" : "1",

	//Structure
	"CTS" : function(name, key){
		return Add(esl.stdvar.Vari(name), Add(Mul(Mod(key, Exp("0x100", "20")), Exp("0x100", "3")), Exp("0x100","2")));
	},

	"CTK" : function(slot){
		return Mod(Div(slot, Exp("0x100","3")), Exp("0x100","20"));
	},

	"ESizeslot" : function(name){
		return Add(esl.stdvar.VariBase(name), this.ESizeOffset);
	},
	
	"MaxESlot" : function(name, key){
		return Add(this.CTS(name, key),this.MaxEOffset);
	},

	"StartSlot" : function(name, key){
		return Add(this.CTS(name, key),this.StartOffset);
	},

	//Gets
	"ESize" : function(addr, name){
		return esl.SA(addr, this.EsizeSlot(name));
	},

	"MaxE" : function(addr, name, key){
		return esl.SA(addr, this.MaxESlot(name, key));
	},
	
	"Element" : function(addr, name, key, index){
		var Esize = this.ESize(addr, name);
		if(this.MaxE(addr, name, key) > index){
			return "0";
		}

		if(Esize == "0x100"){
			return esl.SA(addr, Add(index, this.StartOffset));
		}else{
			var eps = Div("0x100",Esize);
			var pos = Mod(index, eps);
			var row = Add(Mod(Div(index, eps),"0xFFFF"), this.StartOffset);

			var sval = esl.SA(addr, row);
			return Mod(Div(sval, Exp(Esize, pos)), Exp("2", Esize)); 
		}
	},
};

esl.kv = {

	"CTS" : function(name, key){
		return Add(esl.stdvar.Vari(name), Add(Mul(Mod(key, Exp("0x100", "20")), Exp("0x100", "3")), Exp("0x100","2")));
	},
	
	"CTK" : function(slot){
		return Mod(Div(slot, Exp("0x100","3")), Exp("0x100","20"));
	},
	
	"Value" : function(addr, name, key){
		return esl.SA(addr, this.CTS(name, key));
	},
};

esl.ll = {

	//Constants
	"TailSlotOffset"  : "0",
	"HeadSlotOffset"  : "1",
	"LenSlotOffset"   : "2",

	"LLSlotSize" 	  : "3",

	"EntryMainOffset" : "0",
	"EntryPrevOffset" : "1",
	"EntryNextOffset" : "2",

	//Structure
	"CTS" : function(name, key){
		return Add(esl.stdvar.Vari(name), Add(Mul(Mod(key, Exp("0x100", "20")), Exp("0x100", "3")), Exp("0x100","2")));
	},
	
	"CTK" : function(slot){
		return Mod(Div(slot, Exp("0x100","3")), Exp("0x100","20"));
	},

	// Structure
	"TailSlot" : function(name){
		return Add(esl.stdvar.VariBase(name), this.TailSlotOffset);
	},
	
	"HeadSlot" : function(name){
		return Add(esl.stdvar.VariBase(name), this.HeadSlotOffset);
	},
	
	"LenSlot" : function(name){
		return Add(esl.stdvar.VariBase(name), this.LenSlotOffset);
	},

	"MainSlot" : function(name, key){
		return Add(this.CTS(name, key), this.EntryMainOffset);
	},
	
	"PrevSlot" : function(name, key){
		return Add(this.CTS(name,key), this.EntryPrevOffset);
	},
	
	"NextSlot" : function(name, key){
		return Add(this.CTS(name,key), this.EntryNextOffset);
	},

	//Gets
	"TailAddr" : function(addr, name){
		var tail = esl.SA(addr, this.TailSlot(name));
		if(IsZero(tail)){
			return null;
		}
		else{
			return tail;
		}
	},
	
	"HeadAddr" : function(addr, name){
		var head = esl.SA(addr, this.HeadSlot(name));
		if(IsZero(head)){
			return null;
		}
		else{
			return head;
		}
	},
	
	"Tail" : function(addr, name){
		var tail = esl.SA(addr, this.TailSlot(name));
		if(IsZero(tail)){
			return null;
		}
		else{
			return this.CTK(tail);
		}
	},
	
	"Head" : function(addr, name){
		var head = esl.SA(addr, this.HeadSlot(name));
		if(IsZero(head)){
			return null;
		}
		else{
			return this.CTK(head);
		}
	},
	
	"Len"  : function(addr, name){
		return esl.SA(addr, this.LenSlot(name));
	},

	"Main" : function(addr, name, key){
		return esl.SA(addr, this.MainSlot(name, key));
	},
	
	"PrevAddr" : function(addr, name, key){
		var prev = esl.SA(addr, this.PrevSlot(name, key));
		if(IsZero(prev)){
			return null;
		}
		else{
			return prev;
		}
	},
	
	"NextAddr" : function(addr, name, key){
		var next = esl.SA(addr, this.NextSlot(name, key));
		if(IsZero(next)){
			return null;
		}
		else{
			return next;
		}
	},
	
	"Prev" : function(addr, name, key){
		var prev = esl.SA(addr, this.PrevSlot(name, key));
		if(IsZero(prev)){
			return null;
		}
		else{
			return this.CTK(prev);
		}	
	},
	
	"Next" : function(addr, name, key){
		var next = esl.SA(addr, this.NextSlot(name, key));
		if(IsZero(next)){
			return null;
		}
		else{
			return this.CTK(next);
		}
	},

	//Gets the whole list. Note the separate function which gets the keys
	"GetList" : function(addr, name, num){
		var list = [];
		var current = this.Tail(addr, name);
		

		if(typeof(num)=="undefined"){
       		while(current !== null){
				list.push(this.Main(addr, name, current));
				current = this.Next(addr, name, current);
			}

       }else{
       		var c = num;
       		while(current !== null && (c > 0)){
       			list.push(this.Main(addr, name, current));
				current = this.Next(addr, name, current);
	           c = c - 1;
	       }
       }

		return list;
	},

	"GetKeys" : function(addr, name, num){
		var keys = [];
		var current = this.Tail(addr, name);
		
		if(typeof(num)=="undefined"){
       		while(current !== null){
				list.push(current);
				current = this.Next(addr, name, current);
			}
       }else{
       		var c = num;
       		while(current !== null && (c > 0)){
       			list.push(current);
				current = this.Next(addr, name, current);
	           c = c - 1;
	       }
       }
		return keys;
	},
	
	"GetPairs" : function(addr, name, num){
       var list = new Array();
       var current = this.Tail(addr, name);
       
        if(typeof(num)=="undefined"){
       		while(current !== null){
       			var pair = {};
	           pair.Key = current;
	           pair.Value = this.Main(addr, name, current);
	           list.push(pair);
	           current = this.Next(addr, name, current);
       		}
       }else{
       		var c = num;
       		while(current !== null && (c > 0)){
       			var pair = {};
	           pair.Key = current;
	           pair.Value = this.Main(addr, name, current);
	           list.push(pair);
	           current = this.Next(addr, name, current);
	           c = c - 1;
	       }
       }
       return list;
   },

   "GetListRev" : function(addr, name, num){
		var list = [];
		var current = this.Head(addr, name);
		if(typeof(num)=="undefined"){
       		while(current !== null){
	       		list.push(this.Main(addr, name, current));
				current = this.Prev(addr, name, current);
			}
       }else{
       		var c = num;
       		while(current !== null && (c > 0)){
       			list.push(this.Main(addr, name, current));
				current = this.Prev(addr, name, current);
	           c = c - 1;
	       }
       }

		return list;
	},

	"GetKeysRev" : function(addr, name, num){
		var keys = [];
		var current = this.Head(addr, name);

		if(typeof(num)=="undefined"){
       		while(current !== null){
       			list.push(current);
				current = this.Prev(addr, name, current);
			}
       }else{
       		var c = num;
       		while(current !== null && (c > 0)){
       			list.push(current);
				current = this.Prev(addr, name, current);
	            c = c - 1;
	       }
       }
		return keys;
	},
	
	"GetPairsRev" : function(addr, name, num){
       var list = new Array();
       var current = this.Head(addr, name);
       if(typeof(num)=="undefined"){
       		while(current !== null){
	           var pair = {};
	           pair.Key = current;
	           pair.Value = this.Main(addr, name, current);
	           list.push(pair);
	           current = this.Prev(addr, name, current);
	       }
       }else{
       		var c = num;
       		while(current !== null && (c > 0)){
	           var pair = {};
	           pair.Key = current;
	           pair.Value = this.Main(addr, name, current);
	           list.push(pair);
	           current = this.Prev(addr, name, current);
	           c = c - 1;
	       }
       }
       return list;
   },
};

esl.single = {
	
	//Structure
	"ValueSlot" : function(name){
		return esl.stdvar.VariBase(name);
	},
	
	//Gets
	"Value" : function(addr, name){
		slotaddr = this.ValueSlot(name);
		return esl.SA(addr, this.ValueSlot(name));
	},
};

esl.double = {
	
	//Structure
	"ValueSlot" : function(name){
		return esl.stdvar.VariBase(name);
	},
	
	"ValueSlot2" : function(name){
		return Add(esl.stdvar.VariBase(name),"1");
	},
	
	//Gets
	"Value" : function(addr, name){
		var values = [];
		values.push(esl.SA(addr, this.ValueSlot(name)));
		values.push(esl.SA(addr, this.ValueSlot2(name)));
		return values;
	},
};


esl.stdvar = {
	
	//Constants
	"StdVarOffset" 	: "0x1",
	"VarSlotSize" 	: "0x5",
	
	"TypeOffset"	: "0x0",
	"NameOffset"	: "0x1",
	"AddPermOffset"	: "0x2",
	"RmPermOffset"	: "0x3",
	"ModPermOffset"	: "0x4",
	
	//Functions?
	"Vari" 	: function(name){
		var sha3 = SHA3(name);
		var fact = Div(sha3, Exp("0x100", "24") );
		var addr = Add(NSBase, Mul(fact,Exp("0x100", "23")) );
		return addr;
	},
	
	"VarBase" : function(base){
		return Add(base, this.VarSlotSize);
	},
	
	"VariBase" : function(varname){
		return this.VarBase(this.Vari(varname));
	},
	
	//Data Slots
	"VarTypeSlot"	: function(name){
		return Add(this.Vari(name),TypeOffset);
	},
	
	"VarNameSlot"	: function(name){
		return Add(this.Vari(name), NameOffset);
	},
	
	"VarAddPermSlot"	: function(name){
		return Add(this.Vari(name), AddPermOffset);
	},
	
	"VarRmPermSlot" 	: function(name){
		return Add(this.Vari(name), RmPermOffset);
	},
	
	"VarModPermSlot"	: function(name){
		return Add(this.Vari(name), ModPermOffset);
	},
	
	//Getting Variable stuff
	"Type" 	: function(addr, name){
		return esl.SA(addr,this.VarTypeSlot(name));
	},
	
	"Name" 	: function(addr, name){
		return esl.SA(addr,this.VarNameSlot(name));
	},
	
	"Addperm" 	: function(addr, varname){
		return esl.SA(addr,this.VarAddPermSlot(name));
	},
	
	"Rmperm" 	: function(addr, varname){
		return esl.SA(addr,this.VarRmPermSlot(name));
	},
	
	"Modperm" 	: function(addr, varname){
		return esl.SA(addr,this.VarModPermSlot(name));
	},
} 
`

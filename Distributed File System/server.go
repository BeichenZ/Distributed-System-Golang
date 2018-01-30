package main

import (
	//"crypto/md5"
	//"encoding/hex"

	// TODO
	"encoding/json"
	"fmt"
	//"log"
	//"math/rand"
	"net"
	"os"
	"net/rpc"
	//"./dfslib"
	"./dfslib/shared"
	"errors"
	"strings"
	//"time"
)
//Implement shared.DFSService Interface
type DFSServiceObj int
//Register a new client and return a
func (t *DFSServiceObj) RegisterNewClient(args *shared.RNCArgs,reply *shared.RNCReply) error{
	//Assumption by Assignment, Number of Clients is capped at 16. No client can be deleted
	availableIndex := -1
	for index,element := range clientList {
		if element.occupied == false {
			availableIndex = index
			break //since no user can be deleted.The first empty one cannot be repetitive
		}else {
			if strings.Compare(element.localIP,args.LocalIP)==0 && strings.Compare(element.localPath,args.LocalPath)==0 {
				availableIndex = index
				break //if this user is previous registered			
			} 
		} 
		
		
	}
	if availableIndex == -1 {
		return errors.New("Application:Client Count is more than 16.")
	}
	//Populate Local Client Info
	clientList[availableIndex].localIP = args.LocalIP
	clientList[availableIndex].localPath = args.LocalPath
	clientList[availableIndex].occupied = true
	clientList[availableIndex].ID = availableIndex+1
	clientList[availableIndex].fileMap = make(map[string]SingleDFSFileInfo)
	//Pass Back Remote Client Info
	(*reply).ID = availableIndex+1
	return nil
}

func (t *DFSServiceObj)GlobalFileExists(args *shared.OneStringMsg, reply *shared.ExistsMsg) error{
	//To-Do: Disconnected Error
	for _,element := range clientList {
		if _,exists := element.fileMap[args.Msg];exists{
			fmt.Println("Find File Name",args.Msg,"under ID:",element.ID)
			(*reply).Exists = true
			return nil
		}
	}
	//if cannot find
	(*reply).Exists = false
	return nil
}

//Data Structure Type
type SingleClientInfo struct {
	occupied bool
	localIP string
	localPath string
	ID int
	fileMap map[string]SingleDFSFileInfo
}
type SingleDFSFileInfo struct{
	chunkVersionArray [256]int //version number 0 is the default version
	
}
//Global Data Storage Shared by Multiple RPC calls and Main
var clientList [16]SingleClientInfo
//Main Method
func main() {
	//Define Used Data Structure
	
	if len(os.Args) != 2 {
		fmt.Println("Invalid Number of Command Line Argument")
		os.Exit(1)
	}

	tcpAddr_Server := os.Args[1]
	//Server's Overall Structure is referenced from https://coderwall.com/p/wohavg/creating-a-simple-tcp-server-in-go
	//RPC Reference:https://parthdesai.me/articles/2016/05/20/go-rpc-server/
	//Listen for incoming connections
	Iconn,err := net.Listen("tcp",tcpAddr_Server)
	Check_ServerError(err)
	defer Iconn.Close()
	fmt.Println("Listening on" + tcpAddr_Server)
	
	//Register DFSService RPC Server
	DFSService_Instance := new(DFSServiceObj)
	DFSService_rpcServer := rpc.NewServer()
	registerRPC_DFSService(DFSService_rpcServer,DFSService_Instance)
	DFSService_rpcServer.Accept(Iconn)
	
 

}
//Wrappers For Registering RPC Services
func registerRPC_Arith(server *rpc.Server,arith shared.Arith){
	server.RegisterName("Arith_Interface",arith)
}
func registerRPC_DFSService(server *rpc.Server,dfsService shared.DFSService){
	server.RegisterName("DFSService",dfsService)
}

//Separate Thread to handle the request
func handleRequest(conn net.Conn){
	buf := make([]byte,1024)
	reqLen,err := conn.Read(buf)
	Check_NonFatalError(err)
	var receivedMsg shared.OneStringMsg
	err = json.Unmarshal(buf[:reqLen],&receivedMsg)
	Check_NonFatalError(err)
	fmt.Println("message from client:",receivedMsg.Msg)
	conn.Write([]byte("Message Received."))
	conn.Close()
}

//Check for Server's error that leads to Server Shut-Down
func Check_ServerError(err error) {
	if err != nil {
		fmt.Println("Error Ocurred:", err)
		os.Exit(0)
	}
}
func Check_NonFatalError(err error){
	if err !=nil {
		fmt.Println("Non-Fatal Error:",err)	
	}
}

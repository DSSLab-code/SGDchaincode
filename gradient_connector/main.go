/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	//"reflect"
	"context"
	"log"
	"net"
	"strconv"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
	"encoding/json"
	pb "github.com/DSSLab-code/distributed-SGD"
	"google.golang.org/grpc"
	uuid "github.com/satori/go.uuid"

)


// Gradient describes basic details of what makes up gradients
type Gradients struct {
	DmValue              string           `json:"dmvalue"`
	DcValue              string           `json:"dcvalue"`
	EpochId		           string           `json:"epochid"`
	Revoked              bool             `json:"revoked"`
}

// QueryResult structure used for handling result of query
type QueryResult struct {
	Key    string   `json:"key"`
	Record *Gradients `json:"record"`
}


const (
	port = ":50051"
)


// server is used to implement helloworld.GreeterServer.
type server struct {
	pb.UnimplementedGradientsServer
}


func (s *server) SendDcDm(ctx context.Context, in *pb.DcDmRequest) (*pb.DcDmReply, error) {

	log.Printf("Received: %v， %v， %v", in.GetDm(), in.GetDc(), in.GetEpochID())

	os.Setenv("DISCOVERY_AS_LOCALHOST", "true")
	wallet, err := gateway.NewFileSystemWallet("wallet")
	if err != nil {
		fmt.Printf("Failed to create wallet: %s\n", err)
		os.Exit(1)
	}

	if !wallet.Exists("appUser1") {
		err = populateWallet(wallet)
		if err != nil {
			fmt.Printf("Failed to populate wallet contents: %s\n", err)
			os.Exit(1)
		}
	}

	ccpPath := filepath.Join(
		"..",
		"..",
		"test-network",
		"organizations",
		"peerOrganizations",
		"org1.example.com",
		"connection-org1.yaml",
	)

	gw, err := gateway.Connect(
		gateway.WithConfig(config.FromFile(filepath.Clean(ccpPath))),
		gateway.WithIdentity(wallet, "appUser"),
	)
	if err != nil {
		fmt.Printf("Failed to connect to gateway: %s\n", err)
		os.Exit(1)
	}
	defer gw.Close()

	network, err := gw.GetNetwork("mychannel")
	if err != nil {
		fmt.Printf("Failed to get network: %s\n", err)
		os.Exit(1)
	}
	contract := network.GetContract("gradients")

	fmt.Println("creating Record .....")

	result, err := contract.SubmitTransaction("createRecord", uuid.NewV4().String(), in.GetDm(), in.GetDc(), in.GetEpochID())
	if err != nil {
		fmt.Printf("Failed to submit transaction: %s\n", err)
		os.Exit(1)
	}

	print(string(result))

	fmt.Println("......create a gradients record successully.")

	result, err = contract.EvaluateTransaction("queryRecordByEpochID", in.GetEpochID()) // query all record
	if err != nil {
		fmt.Printf("Failed to evaluate transaction: %s\n", err)
		os.Exit(1)
	}
	fmt.Println(string(result))

	var obj[] QueryResult

	if err := json.Unmarshal(result, &obj); err != nil {
		fmt.Printf("Failed to Ummarshal: %s\n", err)
	}


	dmvalue := ""
  dcvalue := ""
	count := 0
	for _, record := range obj {

		    dmvalue += string(record.Record.DmValue) + ",  "
				dcvalue += string(record.Record.DcValue) + ",  "
				count += 1
  }


	return &pb.DcDmReply{Weights: dmvalue, Bias: dcvalue, Com: strconv.Itoa(count)}, nil

}





func main() {

	lis, err := net.Listen("tcp", port)
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
		s := grpc.NewServer()
		pb.RegisterGradientsServer(s, &server{})
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}

}



func populateWallet(wallet *gateway.Wallet) error {
	credPath := filepath.Join(
		"..",
		"..",
		"test-network",
		"organizations",
		"peerOrganizations",
		"org1.example.com",
		"users",
		"User1@org1.example.com",
		"msp",
	)

	certPath := filepath.Join(credPath, "signcerts", "cert.pem")
	// read the certificate pem
	cert, err := ioutil.ReadFile(filepath.Clean(certPath))
	if err != nil {
		return err
	}

	keyDir := filepath.Join(credPath, "keystore")
	// there's a single file in this dir containing the private key
	files, err := ioutil.ReadDir(keyDir)
	if err != nil {
		return err
	}
	if len(files) != 1 {
		return errors.New("keystore folder should have contain one file")
	}
	keyPath := filepath.Join(keyDir, files[0].Name())
	key, err := ioutil.ReadFile(filepath.Clean(keyPath))
	if err != nil {
		return err
	}

	identity := gateway.NewX509Identity("Org1MSP", string(cert), string(key))

	err = wallet.Put("appUser", identity)
	if err != nil {
		return err
	}
	return nil
}

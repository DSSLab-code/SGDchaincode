package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	//"github.com/hyperledger/fabric-contract-api-go/metadata"
	//"github.com/hyperledger/fabric-contract-api-go/serializer"
)

// SmartContract provides functions for managing gradients
type SmartContract struct {
	contractapi.Contract
}


type AdditionalTerms map[string]interface{}

// Gradient describes basic details of what makes up gradients
type Gradients struct {
	DmValue              string           `json:"dmvalue"`
	DcValue              string           `json:"dcvalue"`
	EpochId							 string           `json:"epochid"`
	Revoked              bool             `json:"revoked"`
}

// QueryResult structure used for handling result of query
type QueryResult struct {
	Key    string   `json:"key"`
	Record *Gradients `json:"record"`
}

// InitLedger adds a base set of gradients to the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {

	records := []Gradients{
		{DmValue : "0.20", DcValue : "0.20", EpochId : "-1"},
		{DmValue : "0.21", DcValue : "0.21", EpochId : "-1"},
		{DmValue : "0.22", DcValue : "0.22", EpochId : "-1"},
		{DmValue : "0.23", DcValue : "0.23", EpochId : "-1"},
		{DmValue : "0.24", DcValue : "0.24", EpochId : "-1"},
	}

	for i, record := range records {
		actual, _ := json.Marshal(record)
		// put the record into the chaincodeQuery
		if err := ctx.GetStub().PutState("REC"+strconv.Itoa(i), actual); err != nil {
			return fmt.Errorf("Failed to put to world state. %s", err.Error())
		}
	}
	return nil
}

// CreateRecord adds a new gradient to the world state with given details
func (s *SmartContract) CreateRecord(ctx contractapi.TransactionContextInterface, recordID string, dmVaule string, dcValue string, epochID string) error {

		record := Gradients{
					DmValue: dmVaule,
		      DcValue: dcValue,
					EpochId: epochID,
		}

		GradientsAsBytes, _ := json.Marshal(record)
		return ctx.GetStub().PutState(recordID, GradientsAsBytes)
}

// Query a gradient with a contractID
func (s *SmartContract) QueryRecord(ctx contractapi.TransactionContextInterface, recordID string) (*Gradients, error) {

	recordAsBytes, err := ctx.GetStub().GetState(recordID)

	if err != nil {
		return nil, fmt.Errorf("Failed to read from world state. %s", err.Error())
	}

	if recordAsBytes == nil {
		return nil, fmt.Errorf("%s does not exist", recordID)
	}

	record := new(Gradients)
	if err = json.Unmarshal(recordAsBytes, record); err != nil {
		return nil, fmt.Errorf("error unmarshaling transaction: %+v", err)
	}

	return record, nil
}

//revoke a contract with a contractID
func (s *SmartContract) RevokeGradients(ctx contractapi.TransactionContextInterface, recordID string) error {
	recordAsBytes, err := ctx.GetStub().GetState(recordID)

	if err != nil {
		return fmt.Errorf("Failed to read from world state. %s", err.Error())
	}

	if recordAsBytes == nil {
		return fmt.Errorf("%s does not exist", recordID)
	}

	record := new(Gradients)
	if err = json.Unmarshal(recordAsBytes, record); err != nil {
		return fmt.Errorf("error unmarshaling transaction: %+v", err)
	}

	record.Revoked = true

	GradientsAsBytes, _ := json.Marshal(record)
	return ctx.GetStub().PutState(recordID, GradientsAsBytes)
}


func constructQueryResponseFromIterator(resultsIterator shim.StateQueryIteratorInterface) ([]QueryResult, error) {
	results := []QueryResult{}

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()

		if err != nil {
			return nil, err
		}

		record := new(Gradients)
		if err = json.Unmarshal(queryResponse.Value, record); err != nil {
			return nil, fmt.Errorf("error unmarshaling transaction: %+v", err)
		}

		queryResult := QueryResult{Key: queryResponse.Key, Record: record}
		results = append(results, queryResult)
	}

	return results, nil
}

/*
func constructQueryResponseFromIteratorforEpochOnly(resultsIterator shim.StateQueryIteratorInterface) (float64, float64, int, error) {

	dm := float64(0)
	dc := float64(0)
	m := float64(0)
	c := float64(0)
	count := 0
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()

		if err != nil {
			return 0, 0, 0, err
		}

		record := new(Gradients)
		if err = json.Unmarshal(queryResponse.Value, record); err != nil {
			return 0, 0, 0, fmt.Errorf("error unmarshaling transaction: %+v", err)
		}

		m, _ = strconv.ParseFloat(record.DmValue, 32)
		c, _ = strconv.ParseFloat(record.DcValue, 32)
		dm += m
		dc += c

		count ++
	}

	return dm, dc, count, nil
}
*/

func (s *SmartContract) QueryAllRecords(ctx contractapi.TransactionContextInterface) ([]QueryResult, error) {
	startKey := ""
	endKey := ""

	resultsIterator, err := ctx.GetStub().GetStateByRange(startKey, endKey)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	return constructQueryResponseFromIterator(resultsIterator)
}

func (s *SmartContract) QueryRecordByEpochID(ctx contractapi.TransactionContextInterface, epochID string) ([]QueryResult, error){
	queryString := fmt.Sprintf(`{"selector":{"epochid":"%s"}}`, epochID)
	return getQueryResultForQueryString(ctx, queryString)
}

func getQueryResultForQueryString(ctx contractapi.TransactionContextInterface, queryString string) ([]QueryResult, error) {
	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	return constructQueryResponseFromIterator(resultsIterator)
}


func main() {

	chaincode, err := contractapi.NewChaincode(new(SmartContract))

	if err != nil {
		fmt.Printf("Error create gradient chaincode: %s", err.Error())
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting gradient chaincode: %s", err.Error())
	}
}

package chaincode

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

type State uint

const (
	// ISSUED state for when a loan asset has been issued
	ISSUED State = iota + 1
	// PENDING state for when a loan asset is pending
	PENDING
	// TRADING state for when a loan asset is trading
	TRADING
	// REDEEMED state for when a loan asset has been redeemed
	REDEEMED
)

func (state State) String() string {
	names := []string{"ISSUED", "PENDING", "TRADING", "REDEEMED"}

	if state < ISSUED || state > REDEEMED {
		return "UNKNOWN"
	}

	return names[state-1]
}

type Asset struct {
	ID      		 string `json:"assetID"`
	Lender           string `json:"lender"`
	Borrower         string `json:"borrower"`

	StartDate        int    `json:"startDate"`
	Amount           int    `json:"amount"`
	EndDate          int    `json:"endDate"`

	BorrowerAddress  string   `json:"senderAddress"`
	InvestorAddress  string   `json:"investorAddress"`
	PaymentHashes    []string `json:"paymentHashes"`

	state            State  `metadata:"currentState"`
}

func (asset *Asset) GetState() State {
	return asset.state
}

// SetIssued returns the state to issued
func (asset *Asset) SetIssued() {
	asset.state = ISSUED
}

// SetTrading sets the state to trading
func (asset *Asset) SetTrading() {
	asset.state = TRADING
}

// SetRedeemed sets the state to redeemed
func (asset *Asset) SetRedeemed() {
	asset.state = REDEEMED
}

// SetRedeemed sets the state to redeemed
func (asset *Asset) SetPending() {
	asset.state = PENDING
}


func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	assets := []Asset{
		{ID: "asset1", StartDate: 20210101, EndDate: 20220101, Amount: 300},
		{ID: "asset2", StartDate: 20210101, EndDate: 20220101, Amount: 400},
		{ID: "asset3", StartDate: 20210101, EndDate: 20220101, Amount: 500},
		{ID: "asset4", StartDate: 20210101, EndDate: 20220101, Amount: 600},
		{ID: "asset5", StartDate: 20210101, EndDate: 20220101, Amount: 700},
		{ID: "asset6", StartDate: 20210101, EndDate: 20220101, Amount: 800},
	}

	for _, asset := range assets {

		asset.SetIssued()

		client, err := submittingClientIdentity(ctx)
		if err != nil {
			return err
		}

		asset.Lender = client

		assetJSON, err := json.Marshal(asset)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(asset.ID, assetJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}
	}

	return nil
}

func (s *SmartContract) CreateAsset(ctx contractapi.TransactionContextInterface, id string, start int, end int, amount int) error {
	exists, err := s.AssetExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the asset %s already exists", id)
	}

	asset := Asset{
		ID:             id,
		StartDate:      start,
		EndDate:        end,
		Amount:         amount,
	}

	asset.SetIssued()

	client, err := submittingClientIdentity(ctx)
	if err != nil {
		return err
	}

	asset.Lender = client

	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, assetJSON)
}

func (s *SmartContract) ReadAsset(ctx contractapi.TransactionContextInterface, id string) (*Asset, error) {
	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if assetJSON == nil {
		return nil, fmt.Errorf("the asset %s does not exist", id)
	}

	var asset Asset
	err = json.Unmarshal(assetJSON, &asset)
	if err != nil {
		return nil, err
	}

	return &asset, nil
}

// func (s *SmartContract) UpdateAsset(ctx contractapi.TransactionContextInterface, id string, color string, size int, owner string, appraisedValue int) error {
// 	exists, err := s.AssetExists(ctx, id)
// 	if err != nil {
// 		return err
// 	}
// 	if !exists {
// 		return fmt.Errorf("the asset %s does not exist", id)
// 	}

// 	asset := Asset{
// 		ID:             id,
// 		Color:          color,
// 		Size:           size,
// 		Owner:          owner,
// 		AppraisedValue: appraisedValue,
// 	}
// 	assetJSON, err := json.Marshal(asset)
// 	if err != nil {
// 		return err
// 	}

// 	return ctx.GetStub().PutState(id, assetJSON)
// }

func (s *SmartContract) DeleteAsset(ctx contractapi.TransactionContextInterface, id string) error {
	exists, err := s.AssetExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the asset %s does not exist", id)
	}

	return ctx.GetStub().DelState(id)
}

func (s *SmartContract) AssetExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return assetJSON != nil, nil
}

func (s *SmartContract) TransferAsset(ctx contractapi.TransactionContextInterface, id string, newOwner string) error {
	asset, err := s.ReadAsset(ctx, id)
	if err != nil {
		return err
	}

	asset.Owner = newOwner
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, assetJSON)
}

func (s *SmartContract) GetAllAssets(ctx contractapi.TransactionContextInterface) ([]*Asset, error) {

	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var assets []*Asset
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var asset Asset
		err = json.Unmarshal(queryResponse.Value, &asset)
		if err != nil {
			return nil, err
		}
		assets = append(assets, &asset)
	}

	return assets, nil
}

func submittingClientIdentity(ctx contractapi.TransactionContextInterface) (string, error) {
	b64ID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return "", fmt.Errorf("Failed to read clientID: %v", err)
	}
	decodeID, err := base64.StdEncoding.DecodeString(b64ID)
	if err != nil {
		return "", fmt.Errorf("failed to base64 decode clientID: %v", err)
	}
	return string(decodeID), nil
}
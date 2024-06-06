package utils

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/stretchr/testify/assert"
	"math/big"
	"strings"
	"testing"
)

const (
	InitPauseModuleID  = "0x832a729e"
	NullAddress        = "0x0000000000000000000000000000000000000000"
	ModuleStartAddress = "0x0000000000000000000000000000000000000001"
)

func TestCalculateInterfaceId(t *testing.T) {
	type args struct {
		contractABI string
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "",
			args: args{contractABI: `
[{"inputs":[],"stateMutability":"nonpayable","type":"constructor"},{"inputs":[],"name":"ApprovalCallerNotOwnerNorApproved","type":"error"},{"inputs":[],"name":"ApprovalQueryForNonexistentToken","type":"error"},{"inputs":[],"name":"BalanceQueryForZeroAddress","type":"error"},{"inputs":[{"internalType":"uint256","name":"newMaxSupply","type":"uint256"}],"name":"CannotExceedMaxSupplyOfUint64","type":"error"},{"inputs":[],"name":"ContractDisabled","type":"error"},{"inputs":[],"name":"EthTransferFailed","type":"error"},{"inputs":[{"internalType":"uint256","name":"basisPoints","type":"uint256"}],"name":"InvalidRoyaltyBasisPoints","type":"error"},{"inputs":[],"name":"LockedBaseURI","type":"error"},{"inputs":[],"name":"MintERC2309QuantityExceedsLimit","type":"error"},{"inputs":[{"internalType":"uint256","name":"total","type":"uint256"},{"internalType":"uint256","name":"maxSupply","type":"uint256"}],"name":"MintQuantityExceedsMaxSupply","type":"error"},{"inputs":[],"name":"MintToZeroAddress","type":"error"},{"inputs":[],"name":"MintZeroQuantity","type":"error"},{"inputs":[],"name":"OnlyAllowedSeaDrop","type":"error"},{"inputs":[],"name":"OnlyOwner","type":"error"},{"inputs":[{"internalType":"address","name":"operator","type":"address"}],"name":"OperatorNotAllowed","type":"error"},{"inputs":[],"name":"OwnerQueryForNonexistentToken","type":"error"},{"inputs":[],"name":"OwnershipNotInitializedForExtraData","type":"error"},{"inputs":[],"name":"ProvenanceHashCannotBeSetAfterMintStarted","type":"error"},{"inputs":[],"name":"RoyaltyAddressCannotBeZeroAddress","type":"error"},{"inputs":[],"name":"SignersMismatch","type":"error"},{"inputs":[],"name":"TokenGatedMismatch","type":"error"},{"inputs":[],"name":"TransferCallerNotOwnerNorApproved","type":"error"},{"inputs":[],"name":"TransferFromIncorrectOwner","type":"error"},{"inputs":[],"name":"TransferToNonERC721ReceiverImplementer","type":"error"},{"inputs":[],"name":"TransferToZeroAddress","type":"error"},{"inputs":[],"name":"URIQueryForNonexistentToken","type":"error"},{"inputs":[],"name":"VerifierNotSet","type":"error"},{"inputs":[],"name":"ZeroAddress","type":"error"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address[]","name":"allowedSeaDrop","type":"address[]"}],"name":"AllowedSeaDropUpdated","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"owner","type":"address"},{"indexed":true,"internalType":"address","name":"approved","type":"address"},{"indexed":true,"internalType":"uint256","name":"tokenId","type":"uint256"}],"name":"Approval","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"owner","type":"address"},{"indexed":true,"internalType":"address","name":"operator","type":"address"},{"indexed":false,"internalType":"bool","name":"approved","type":"bool"}],"name":"ApprovalForAll","type":"event"},{"anonymous":false,"inputs":[],"name":"BaseURILocked","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"uint256","name":"_fromTokenId","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"_toTokenId","type":"uint256"}],"name":"BatchMetadataUpdate","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"uint256","name":"fromTokenId","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"toTokenId","type":"uint256"},{"indexed":true,"internalType":"address","name":"from","type":"address"},{"indexed":true,"internalType":"address","name":"to","type":"address"}],"name":"ConsecutiveTransfer","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"string","name":"newContractURI","type":"string"}],"name":"ContractURIUpdated","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"bool","name":"newDisabledState","type":"bool"}],"name":"IsDisabledSet","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"uint256","name":"newMaxSupply","type":"uint256"}],"name":"MaxSupplyUpdated","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"previousOwner","type":"address"},{"indexed":true,"internalType":"address","name":"newOwner","type":"address"}],"name":"OwnershipTransferStarted","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"previousOwner","type":"address"},{"indexed":true,"internalType":"address","name":"newOwner","type":"address"}],"name":"OwnershipTransferred","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"newPrimeAddress","type":"address"}],"name":"PrimeAddressSet","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"bytes32","name":"previousHash","type":"bytes32"},{"indexed":false,"internalType":"bytes32","name":"newHash","type":"bytes32"}],"name":"ProvenanceHashUpdated","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"receiver","type":"address"},{"indexed":false,"internalType":"uint256","name":"bps","type":"uint256"}],"name":"RoyaltyInfoUpdated","type":"event"},{"anonymous":false,"inputs":[],"name":"SeaDropTokenDeployed","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"from","type":"address"},{"indexed":true,"internalType":"address","name":"to","type":"address"},{"indexed":true,"internalType":"uint256","name":"tokenId","type":"uint256"}],"name":"Transfer","type":"event"},{"inputs":[],"name":"OPERATOR_FILTER_REGISTRY","outputs":[{"internalType":"contract IOperatorFilterRegistry","name":"","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"_baseURILocked","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"_contractURI","outputs":[{"internalType":"string","name":"","type":"string"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"_disabled","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"_maxSupply","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"_prime","outputs":[{"internalType":"contract IERC20","name":"","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"_provenanceHash","outputs":[{"internalType":"bytes32","name":"","type":"bytes32"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint256","name":"","type":"uint256"}],"name":"_routerEndpoints","outputs":[{"internalType":"address","name":"nftReceiver","type":"address"},{"internalType":"address","name":"ethReceiver","type":"address"},{"internalType":"address","name":"primeReceiver","type":"address"},{"internalType":"address","name":"verifier","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"_royaltyInfo","outputs":[{"internalType":"address","name":"royaltyAddress","type":"address"},{"internalType":"uint96","name":"royaltyBps","type":"uint96"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"_tokenBaseURI","outputs":[{"internalType":"string","name":"","type":"string"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"acceptOwnership","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"operator","type":"address"},{"internalType":"uint256","name":"tokenId","type":"uint256"}],"name":"approve","outputs":[],"stateMutability":"payable","type":"function"},{"inputs":[{"internalType":"address","name":"owner","type":"address"}],"name":"balanceOf","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"baseURI","outputs":[{"internalType":"string","name":"","type":"string"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"contractURI","outputs":[{"internalType":"string","name":"","type":"string"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint256","name":"fromTokenId","type":"uint256"},{"internalType":"uint256","name":"toTokenId","type":"uint256"}],"name":"emitBatchMetadataUpdate","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"uint256","name":"tokenId","type":"uint256"}],"name":"getApproved","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"minter","type":"address"}],"name":"getMintStats","outputs":[{"internalType":"uint256","name":"minterNumMinted","type":"uint256"},{"internalType":"uint256","name":"currentTotalSupply","type":"uint256"},{"internalType":"uint256","name":"maxSupply","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint256","name":"_id","type":"uint256"},{"internalType":"uint256[]","name":"_tokenIds","type":"uint256[]"},{"internalType":"uint256","name":"_primeValue","type":"uint256"},{"internalType":"bytes","name":"_data","type":"bytes"}],"name":"invoke","outputs":[],"stateMutability":"payable","type":"function"},{"inputs":[{"internalType":"address","name":"owner","type":"address"},{"internalType":"address","name":"operator","type":"address"}],"name":"isApprovedForAll","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"isFilterDisabled","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"lockBaseURI","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"maxSupply","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"minter","type":"address"},{"internalType":"uint256","name":"quantity","type":"uint256"}],"name":"mintSeaDrop","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"components":[{"internalType":"uint256","name":"maxSupply","type":"uint256"},{"internalType":"string","name":"baseURI","type":"string"},{"internalType":"string","name":"contractURI","type":"string"},{"internalType":"address","name":"seaDropImpl","type":"address"},{"components":[{"internalType":"uint80","name":"mintPrice","type":"uint80"},{"internalType":"uint48","name":"startTime","type":"uint48"},{"internalType":"uint48","name":"endTime","type":"uint48"},{"internalType":"uint16","name":"maxTotalMintableByWallet","type":"uint16"},{"internalType":"uint16","name":"feeBps","type":"uint16"},{"internalType":"bool","name":"restrictFeeRecipients","type":"bool"}],"internalType":"struct PublicDrop","name":"publicDrop","type":"tuple"},{"internalType":"string","name":"dropURI","type":"string"},{"components":[{"internalType":"bytes32","name":"merkleRoot","type":"bytes32"},{"internalType":"string[]","name":"publicKeyURIs","type":"string[]"},{"internalType":"string","name":"allowListURI","type":"string"}],"internalType":"struct AllowListData","name":"allowListData","type":"tuple"},{"internalType":"address","name":"creatorPayoutAddress","type":"address"},{"internalType":"bytes32","name":"provenanceHash","type":"bytes32"},{"internalType":"address[]","name":"allowedFeeRecipients","type":"address[]"},{"internalType":"address[]","name":"disallowedFeeRecipients","type":"address[]"},{"internalType":"address[]","name":"allowedPayers","type":"address[]"},{"internalType":"address[]","name":"disallowedPayers","type":"address[]"},{"internalType":"address[]","name":"tokenGatedAllowedNftTokens","type":"address[]"},{"components":[{"internalType":"uint80","name":"mintPrice","type":"uint80"},{"internalType":"uint16","name":"maxTotalMintableByWallet","type":"uint16"},{"internalType":"uint48","name":"startTime","type":"uint48"},{"internalType":"uint48","name":"endTime","type":"uint48"},{"internalType":"uint8","name":"dropStageIndex","type":"uint8"},{"internalType":"uint32","name":"maxTokenSupplyForStage","type":"uint32"},{"internalType":"uint16","name":"feeBps","type":"uint16"},{"internalType":"bool","name":"restrictFeeRecipients","type":"bool"}],"internalType":"struct TokenGatedDropStage[]","name":"tokenGatedDropStages","type":"tuple[]"},{"internalType":"address[]","name":"disallowedTokenGatedAllowedNftTokens","type":"address[]"},{"internalType":"address[]","name":"signers","type":"address[]"},{"components":[{"internalType":"uint80","name":"minMintPrice","type":"uint80"},{"internalType":"uint24","name":"maxMaxTotalMintableByWallet","type":"uint24"},{"internalType":"uint40","name":"minStartTime","type":"uint40"},{"internalType":"uint40","name":"maxEndTime","type":"uint40"},{"internalType":"uint40","name":"maxMaxTokenSupplyForStage","type":"uint40"},{"internalType":"uint16","name":"minFeeBps","type":"uint16"},{"internalType":"uint16","name":"maxFeeBps","type":"uint16"}],"internalType":"struct SignedMintValidationParams[]","name":"signedMintValidationParams","type":"tuple[]"},{"internalType":"address[]","name":"disallowedSigners","type":"address[]"}],"internalType":"struct ERC721SeaDropStructsErrorsAndEvents.MultiConfigureStruct","name":"config","type":"tuple"}],"name":"multiConfigure","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"name","outputs":[{"internalType":"string","name":"","type":"string"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"owner","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint256","name":"tokenId","type":"uint256"}],"name":"ownerOf","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"pendingOwner","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"provenanceHash","outputs":[{"internalType":"bytes32","name":"","type":"bytes32"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"renounceOwnership","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"royaltyAddress","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"royaltyBasisPoints","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint256","name":"","type":"uint256"},{"internalType":"uint256","name":"_salePrice","type":"uint256"}],"name":"royaltyInfo","outputs":[{"internalType":"address","name":"receiver","type":"address"},{"internalType":"uint256","name":"royaltyAmount","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"from","type":"address"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"tokenId","type":"uint256"}],"name":"safeTransferFrom","outputs":[],"stateMutability":"payable","type":"function"},{"inputs":[{"internalType":"address","name":"from","type":"address"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"tokenId","type":"uint256"},{"internalType":"bytes","name":"data","type":"bytes"}],"name":"safeTransferFrom","outputs":[],"stateMutability":"payable","type":"function"},{"inputs":[{"internalType":"address","name":"operator","type":"address"},{"internalType":"bool","name":"approved","type":"bool"}],"name":"setApprovalForAll","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"string","name":"newBaseURI","type":"string"}],"name":"setBaseURI","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"string","name":"newContractURI","type":"string"}],"name":"setContractURI","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"bool","name":"disabled","type":"bool"}],"name":"setDisabled","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"prime","type":"address"}],"name":"setPrime","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"bytes32","name":"newProvenanceHash","type":"bytes32"}],"name":"setProvenanceHash","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"uint256","name":"_id","type":"uint256"},{"internalType":"address","name":"_nftReceiver","type":"address"},{"internalType":"address","name":"_ethReceiver","type":"address"},{"internalType":"address","name":"_primeReceiver","type":"address"},{"internalType":"address","name":"_verifier","type":"address"}],"name":"setReceiver","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"components":[{"internalType":"address","name":"royaltyAddress","type":"address"},{"internalType":"uint96","name":"royaltyBps","type":"uint96"}],"internalType":"struct IParallelAvatarInvoke.RoyaltyInfo","name":"newInfo","type":"tuple"}],"name":"setRoyaltyInfo","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"bytes4","name":"interfaceId","type":"bytes4"}],"name":"supportsInterface","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"symbol","outputs":[{"internalType":"string","name":"","type":"string"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint256","name":"tokenId","type":"uint256"}],"name":"tokenURI","outputs":[{"internalType":"string","name":"","type":"string"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"totalSupply","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"from","type":"address"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"tokenId","type":"uint256"}],"name":"transferFrom","outputs":[],"stateMutability":"payable","type":"function"},{"inputs":[{"internalType":"address","name":"newOwner","type":"address"}],"name":"transferOwnership","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"seaDropImpl","type":"address"},{"components":[{"internalType":"bytes32","name":"merkleRoot","type":"bytes32"},{"internalType":"string[]","name":"publicKeyURIs","type":"string[]"},{"internalType":"string","name":"allowListURI","type":"string"}],"internalType":"struct AllowListData","name":"allowListData","type":"tuple"}],"name":"updateAllowList","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"seaDropImpl","type":"address"},{"internalType":"address","name":"feeRecipient","type":"address"},{"internalType":"bool","name":"allowed","type":"bool"}],"name":"updateAllowedFeeRecipient","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address[]","name":"allowedSeaDrop","type":"address[]"}],"name":"updateAllowedSeaDrop","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"seaDropImpl","type":"address"},{"internalType":"address","name":"payoutAddress","type":"address"}],"name":"updateCreatorPayoutAddress","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"seaDropImpl","type":"address"},{"internalType":"string","name":"dropURI","type":"string"}],"name":"updateDropURI","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"seaDropImpl","type":"address"},{"internalType":"address","name":"payer","type":"address"},{"internalType":"bool","name":"allowed","type":"bool"}],"name":"updatePayer","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"seaDropImpl","type":"address"},{"components":[{"internalType":"uint80","name":"mintPrice","type":"uint80"},{"internalType":"uint48","name":"startTime","type":"uint48"},{"internalType":"uint48","name":"endTime","type":"uint48"},{"internalType":"uint16","name":"maxTotalMintableByWallet","type":"uint16"},{"internalType":"uint16","name":"feeBps","type":"uint16"},{"internalType":"bool","name":"restrictFeeRecipients","type":"bool"}],"internalType":"struct PublicDrop","name":"publicDrop","type":"tuple"}],"name":"updatePublicDrop","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"seaDropImpl","type":"address"},{"internalType":"address","name":"signer","type":"address"},{"components":[{"internalType":"uint80","name":"minMintPrice","type":"uint80"},{"internalType":"uint24","name":"maxMaxTotalMintableByWallet","type":"uint24"},{"internalType":"uint40","name":"minStartTime","type":"uint40"},{"internalType":"uint40","name":"maxEndTime","type":"uint40"},{"internalType":"uint40","name":"maxMaxTokenSupplyForStage","type":"uint40"},{"internalType":"uint16","name":"minFeeBps","type":"uint16"},{"internalType":"uint16","name":"maxFeeBps","type":"uint16"}],"internalType":"struct SignedMintValidationParams","name":"signedMintValidationParams","type":"tuple"}],"name":"updateSignedMintValidationParams","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"seaDropImpl","type":"address"},{"internalType":"address","name":"allowedNftToken","type":"address"},{"components":[{"internalType":"uint80","name":"mintPrice","type":"uint80"},{"internalType":"uint16","name":"maxTotalMintableByWallet","type":"uint16"},{"internalType":"uint48","name":"startTime","type":"uint48"},{"internalType":"uint48","name":"endTime","type":"uint48"},{"internalType":"uint8","name":"dropStageIndex","type":"uint8"},{"internalType":"uint32","name":"maxTokenSupplyForStage","type":"uint32"},{"internalType":"uint16","name":"feeBps","type":"uint16"},{"internalType":"bool","name":"restrictFeeRecipients","type":"bool"}],"internalType":"struct TokenGatedDropStage","name":"dropStage","type":"tuple"}],"name":"updateTokenGatedDrop","outputs":[],"stateMutability":"nonpayable","type":"function"}]
`},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cABI, _ := abi.JSON(strings.NewReader(tt.args.contractABI))
			selects := make([][]byte, 0)
			for _, method := range cABI.Methods {
				selects = append(selects, method.ID)
			}

			for _, event := range cABI.Events {
				fmt.Println(event.ID.Bytes())
			}

			result := make([]byte, 4)
			for i := 0; i < 4; i++ {
				result[i] = selects[0][i]
				for j := 1; j < len(selects); j++ {
					result[i] = result[i] ^ selects[j][i]
				}
			}
			fmt.Println(common.Bytes2Hex(result[:]))
		})
	}
}

func TestCodeAt(t *testing.T) {
	type args struct {
		ctx      context.Context
		c        *rpc.Client
		contract string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr error
	}{
		{
			name: "",
			args: args{
				ctx:      context.Background(),
				c:        ETHClient,
				contract: "0xa68eea7850fb8fa9bcb003441258de8056166aa3",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CodeAt(tt.args.ctx, tt.args.c, tt.args.contract)
			assert.Nil(t, err)
			fmt.Println(got)
		})
	}
}

func TestCall(t *testing.T) {
	type args struct {
		ctx      context.Context
		c        *rpc.Client
		contract string
		abi      string
	}
	tests := []struct {
		name    string
		args    args
		want    any
		wantErr error
	}{
		//{
		//	name: "owner",
		//	args: args{
		//		ctx:      context.Background(),
		//		c:        MerlinClient,
		//		contract: "0xe29a6620dac789b8a76e9b9ec8fe9b7cf2b663d5",
		//		abi:      "[{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
		//	},
		//},
		{
			name: "owner",
			args: args{
				ctx:      context.Background(),
				c:        BscClient,
				contract: "0xa281c299f9e85d15f2c743c727ee34551e13d37e",
				abi:      `[{"inputs": [],"name": "getOwners","outputs": [{"internalType": "address[]","name": "","type": "address[]"}],"stateMutability": "view","type": "function"}]`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decodeABI, err := DecodeABI(tt.args.abi)
			assert.Nil(t, err)
			got, err := Call(tt.args.ctx, tt.args.c, tt.args.contract, decodeABI.Methods["getOwners"])
			assert.Nil(t, err)
			if addrs, can := got.([]common.Address); can {
				for _, v := range addrs {
					fmt.Println(v.String())
				}
			}
		})
	}
}

func TestCallWithInput(t *testing.T) {
	type args struct {
		ctx      context.Context
		c        *rpc.Client
		contract string
		abi      string
	}
	tests := []struct {
		name    string
		args    args
		want    any
		wantErr error
	}{
		//{
		//	name: "getModulesPaginated",
		//	args: args{
		//		ctx:      context.Background(),
		//		c:        BscClient,
		//		contract: "0xa281c299f9e85d15f2c743c727ee34551e13d37e",
		//		abi:      `[{"inputs":[{"internalType":"address","name":"start","type":"address"},{"internalType":"uint256","name":"pageSize","type":"uint256"}],"name":"getModulesPaginated","outputs":[{"internalType":"address[]","name":"array","type":"address[]"},{"internalType":"address","name":"next","type":"address"}],"stateMutability":"view","type":"function"}]`,
		//	},
		//},
		{
			name: "getAmountsIn",
			args: args{
				ctx:      context.Background(),
				c:        BscClient,
				contract: "0x7a250d5630b4cf539739df2c5dacb4c659f2488d",
				abi:      `[{"inputs":[{"internalType":"address","name":"_factory","type":"address"},{"internalType":"address","name":"_WETH","type":"address"}],"stateMutability":"nonpayable","type":"constructor"},{"inputs":[],"name":"WETH","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"tokenA","type":"address"},{"internalType":"address","name":"tokenB","type":"address"},{"internalType":"uint256","name":"amountADesired","type":"uint256"},{"internalType":"uint256","name":"amountBDesired","type":"uint256"},{"internalType":"uint256","name":"amountAMin","type":"uint256"},{"internalType":"uint256","name":"amountBMin","type":"uint256"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"deadline","type":"uint256"}],"name":"addLiquidity","outputs":[{"internalType":"uint256","name":"amountA","type":"uint256"},{"internalType":"uint256","name":"amountB","type":"uint256"},{"internalType":"uint256","name":"liquidity","type":"uint256"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"token","type":"address"},{"internalType":"uint256","name":"amountTokenDesired","type":"uint256"},{"internalType":"uint256","name":"amountTokenMin","type":"uint256"},{"internalType":"uint256","name":"amountETHMin","type":"uint256"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"deadline","type":"uint256"}],"name":"addLiquidityETH","outputs":[{"internalType":"uint256","name":"amountToken","type":"uint256"},{"internalType":"uint256","name":"amountETH","type":"uint256"},{"internalType":"uint256","name":"liquidity","type":"uint256"}],"stateMutability":"payable","type":"function"},{"inputs":[],"name":"factory","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint256","name":"amountOut","type":"uint256"},{"internalType":"uint256","name":"reserveIn","type":"uint256"},{"internalType":"uint256","name":"reserveOut","type":"uint256"}],"name":"getAmountIn","outputs":[{"internalType":"uint256","name":"amountIn","type":"uint256"}],"stateMutability":"pure","type":"function"},{"inputs":[{"internalType":"uint256","name":"amountIn","type":"uint256"},{"internalType":"uint256","name":"reserveIn","type":"uint256"},{"internalType":"uint256","name":"reserveOut","type":"uint256"}],"name":"getAmountOut","outputs":[{"internalType":"uint256","name":"amountOut","type":"uint256"}],"stateMutability":"pure","type":"function"},{"inputs":[{"internalType":"uint256","name":"amountOut","type":"uint256"},{"internalType":"address[]","name":"path","type":"address[]"}],"name":"getAmountsIn","outputs":[{"internalType":"uint256[]","name":"amounts","type":"uint256[]"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint256","name":"amountIn","type":"uint256"},{"internalType":"address[]","name":"path","type":"address[]"}],"name":"getAmountsOut","outputs":[{"internalType":"uint256[]","name":"amounts","type":"uint256[]"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint256","name":"amountA","type":"uint256"},{"internalType":"uint256","name":"reserveA","type":"uint256"},{"internalType":"uint256","name":"reserveB","type":"uint256"}],"name":"quote","outputs":[{"internalType":"uint256","name":"amountB","type":"uint256"}],"stateMutability":"pure","type":"function"},{"inputs":[{"internalType":"address","name":"tokenA","type":"address"},{"internalType":"address","name":"tokenB","type":"address"},{"internalType":"uint256","name":"liquidity","type":"uint256"},{"internalType":"uint256","name":"amountAMin","type":"uint256"},{"internalType":"uint256","name":"amountBMin","type":"uint256"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"deadline","type":"uint256"}],"name":"removeLiquidity","outputs":[{"internalType":"uint256","name":"amountA","type":"uint256"},{"internalType":"uint256","name":"amountB","type":"uint256"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"token","type":"address"},{"internalType":"uint256","name":"liquidity","type":"uint256"},{"internalType":"uint256","name":"amountTokenMin","type":"uint256"},{"internalType":"uint256","name":"amountETHMin","type":"uint256"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"deadline","type":"uint256"}],"name":"removeLiquidityETH","outputs":[{"internalType":"uint256","name":"amountToken","type":"uint256"},{"internalType":"uint256","name":"amountETH","type":"uint256"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"token","type":"address"},{"internalType":"uint256","name":"liquidity","type":"uint256"},{"internalType":"uint256","name":"amountTokenMin","type":"uint256"},{"internalType":"uint256","name":"amountETHMin","type":"uint256"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"deadline","type":"uint256"}],"name":"removeLiquidityETHSupportingFeeOnTransferTokens","outputs":[{"internalType":"uint256","name":"amountETH","type":"uint256"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"token","type":"address"},{"internalType":"uint256","name":"liquidity","type":"uint256"},{"internalType":"uint256","name":"amountTokenMin","type":"uint256"},{"internalType":"uint256","name":"amountETHMin","type":"uint256"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"deadline","type":"uint256"},{"internalType":"bool","name":"approveMax","type":"bool"},{"internalType":"uint8","name":"v","type":"uint8"},{"internalType":"bytes32","name":"r","type":"bytes32"},{"internalType":"bytes32","name":"s","type":"bytes32"}],"name":"removeLiquidityETHWithPermit","outputs":[{"internalType":"uint256","name":"amountToken","type":"uint256"},{"internalType":"uint256","name":"amountETH","type":"uint256"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"token","type":"address"},{"internalType":"uint256","name":"liquidity","type":"uint256"},{"internalType":"uint256","name":"amountTokenMin","type":"uint256"},{"internalType":"uint256","name":"amountETHMin","type":"uint256"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"deadline","type":"uint256"},{"internalType":"bool","name":"approveMax","type":"bool"},{"internalType":"uint8","name":"v","type":"uint8"},{"internalType":"bytes32","name":"r","type":"bytes32"},{"internalType":"bytes32","name":"s","type":"bytes32"}],"name":"removeLiquidityETHWithPermitSupportingFeeOnTransferTokens","outputs":[{"internalType":"uint256","name":"amountETH","type":"uint256"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"tokenA","type":"address"},{"internalType":"address","name":"tokenB","type":"address"},{"internalType":"uint256","name":"liquidity","type":"uint256"},{"internalType":"uint256","name":"amountAMin","type":"uint256"},{"internalType":"uint256","name":"amountBMin","type":"uint256"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"deadline","type":"uint256"},{"internalType":"bool","name":"approveMax","type":"bool"},{"internalType":"uint8","name":"v","type":"uint8"},{"internalType":"bytes32","name":"r","type":"bytes32"},{"internalType":"bytes32","name":"s","type":"bytes32"}],"name":"removeLiquidityWithPermit","outputs":[{"internalType":"uint256","name":"amountA","type":"uint256"},{"internalType":"uint256","name":"amountB","type":"uint256"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"uint256","name":"amountOut","type":"uint256"},{"internalType":"address[]","name":"path","type":"address[]"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"deadline","type":"uint256"}],"name":"swapETHForExactTokens","outputs":[{"internalType":"uint256[]","name":"amounts","type":"uint256[]"}],"stateMutability":"payable","type":"function"},{"inputs":[{"internalType":"uint256","name":"amountOutMin","type":"uint256"},{"internalType":"address[]","name":"path","type":"address[]"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"deadline","type":"uint256"}],"name":"swapExactETHForTokens","outputs":[{"internalType":"uint256[]","name":"amounts","type":"uint256[]"}],"stateMutability":"payable","type":"function"},{"inputs":[{"internalType":"uint256","name":"amountOutMin","type":"uint256"},{"internalType":"address[]","name":"path","type":"address[]"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"deadline","type":"uint256"}],"name":"swapExactETHForTokensSupportingFeeOnTransferTokens","outputs":[],"stateMutability":"payable","type":"function"},{"inputs":[{"internalType":"uint256","name":"amountIn","type":"uint256"},{"internalType":"uint256","name":"amountOutMin","type":"uint256"},{"internalType":"address[]","name":"path","type":"address[]"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"deadline","type":"uint256"}],"name":"swapExactTokensForETH","outputs":[{"internalType":"uint256[]","name":"amounts","type":"uint256[]"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"uint256","name":"amountIn","type":"uint256"},{"internalType":"uint256","name":"amountOutMin","type":"uint256"},{"internalType":"address[]","name":"path","type":"address[]"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"deadline","type":"uint256"}],"name":"swapExactTokensForETHSupportingFeeOnTransferTokens","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"uint256","name":"amountIn","type":"uint256"},{"internalType":"uint256","name":"amountOutMin","type":"uint256"},{"internalType":"address[]","name":"path","type":"address[]"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"deadline","type":"uint256"}],"name":"swapExactTokensForTokens","outputs":[{"internalType":"uint256[]","name":"amounts","type":"uint256[]"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"uint256","name":"amountIn","type":"uint256"},{"internalType":"uint256","name":"amountOutMin","type":"uint256"},{"internalType":"address[]","name":"path","type":"address[]"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"deadline","type":"uint256"}],"name":"swapExactTokensForTokensSupportingFeeOnTransferTokens","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"uint256","name":"amountOut","type":"uint256"},{"internalType":"uint256","name":"amountInMax","type":"uint256"},{"internalType":"address[]","name":"path","type":"address[]"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"deadline","type":"uint256"}],"name":"swapTokensForExactETH","outputs":[{"internalType":"uint256[]","name":"amounts","type":"uint256[]"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"uint256","name":"amountOut","type":"uint256"},{"internalType":"uint256","name":"amountInMax","type":"uint256"},{"internalType":"address[]","name":"path","type":"address[]"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"deadline","type":"uint256"}],"name":"swapTokensForExactTokens","outputs":[{"internalType":"uint256[]","name":"amounts","type":"uint256[]"}],"stateMutability":"nonpayable","type":"function"},{"stateMutability":"payable","type":"receive"}]`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decodeABI, err := DecodeABI(tt.args.abi)
			assert.Nil(t, err)
			//endAddress := NullAddress
			//params := []any{
			//	common.HexToAddress(ModuleStartAddress),
			//	big.NewInt(100),
			//}
			//addresses := make([]string, 0)
			//assert.Nil(t, err)
			//for {
			//	if endAddress == ModuleStartAddress {
			//		break
			//	}
			//
			//	if endAddress != NullAddress {
			//		params[0] = common.HexToAddress(endAddress)
			//	}
			//
			//	res, err := CallWithInput(tt.args.ctx, tt.args.c, tt.args.contract, "getModulesPaginated", decodeABI, params...)
			//	assert.Nil(t, err)
			//
			//	if len(res) != 2 {
			//		return
			//	}
			//
			//	if addrs, can := res[0].([]common.Address); can {
			//		for _, v := range addrs {
			//			a := strings.ToLower(v.String())
			//			fmt.Println(a)
			//			addresses = append(addresses, a)
			//		}
			//	}
			//
			//	if next, can := res[1].(common.Address); can {
			//		endAddress = strings.ToLower(next.String())
			//	}
			//}

			addresses := []common.Address{
				common.HexToAddress("0xfFf9976782d46CC05630D1f6eBAb18b2324d6B14"),
				common.HexToAddress("0xBebDA684b353201D3C38F013c9F09EA520Eb5490"),
			}
			params := []any{
				big.NewInt(1000000000000),
				addresses,
			}

			res, err := CallWithInput(tt.args.ctx, tt.args.c, tt.args.contract, "getAmountsIn", decodeABI, params...)
			assert.Nil(t, err)

			for _, v := range res {
				fmt.Println(v)
			}
		})
	}
}

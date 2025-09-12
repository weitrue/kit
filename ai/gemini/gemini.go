package gemini

import (
	"cloud.google.com/go/auth/credentials"
	"context"
	"encoding/json"
	"fmt"
	"google.golang.org/api/option"
	"io"

	genai1 "cloud.google.com/go/vertexai/genai"
	"google.golang.org/genai"
)

var (
	credentialsData            = `{}`
	projectID, model, location = "", "gemini-2.0-flash", "us-east4"
)

func tryGemini(w io.Writer, projectID string) error {
	ctx := context.Background()
	client, err := genai1.NewClient(ctx, projectID, location, option.WithCredentialsJSON([]byte(credentialsData)))
	if err != nil {
		return fmt.Errorf("error creating client: %w", err)
	}
	gemini := client.GenerativeModel(model)

	prompt := genai1.Text("You are a proficient and experienced on-chain transaction analyser. I will give you details of an Ethereum transaction based on Etherscan, please summarize the transaction, including any relevant smart contract interactions. You may analyze and understand all token transfers, qualify the transactions, guess the motivation based on transaction actions, and make a summary answering what the initiator(`From` address) and relating addresses did or caused during this transaction.\nHere are some knowledge: \n* Token Transfer could be invoked by contract.\n* `Labels` and `Inputdata` are crucial to understand a transaction. They give you a useful mapping that could help identify the transaction type and motivation.\n* Etherscan has labeled all fake tokens as \"ERC-20 TOKEN*\", transactions where it appears are probably related to a phishing scam.\n* Transaction where exploiter address appears as `From` is probably related to a exploitation.\n* \"revoke\" refers to the process of withdrawing previously granted permissions, such as allowing a smart contract to spend tokens on a user's behalf.\n* \"approve\" refers to the process of granting permissions, such as allowing a smart contract/EOA to spend tokens on a user's behalf.\n* The asset in an approval is usually the token issued by the `InteractedWith` contract.\n* Definition of different fields in the transaction detail:\n** `From`: The sending party of the transaction.\n** `InteractedWith`: The contract to be interacted with.\n** `Value`: The value being transacted in Ether.\n** `Inputdata`: Additional data included for this transaction. Commonly used as part of contract interaction or as a message sent to the recipient.\n** `TransactionAction`: Etherscan highlighted events of the transaction. Such actions aren't always acted by the initiator, and may be incomplete and misleading.\n** `ETHTransfer`: List of native token transferred in the transaction.\n** `ERC{X}Transfer`: List of ERC-{X} tokens transferred in the transaction.\n** `Nonce`: a sequential number assigned to each transaction made by an account.\n** `PositionInBlock`: the position or index of a specific transaction within the block it is included in.\n** `TxnType`: \"1\" means legacy transaction, \"2\" means EIP-1159 transaction.\n** `TransactionFee`: Amount paid to the miner and block producer for processing the transaction.\n** `Gas`: Maximum amount of gas allocated for the transaction & the amount eventually used.\n** `TransactionNote`: Some keywords about the transaction.\n** `Labels`: A list of labels that are associated with the transaction.\nThese are some output format requirments:\n1. You should recognize the service and tell the main components in a service transaction, `TimeStamp` usually means a lot for a service.\n2. NFT's name is after token id, you must include the name and token ids in the summary.\n3. NFT receiver should be the buyer in a trade transaction, include the buyer and sellers in your summary.\nNow, please summarize this transaction concisely and insightfully in 1 sentence.\nTransaction details: \n{\"TxnType\":\"2\",\"From\":\"0xe08e164ba85890ac94dbeea77353d46f55ddf261\",\"InteractedWith\":\"0x858646372cc42e1a627fce94aa7a7033e7cf075a\",\"Value\":\"0\",\"Status\":\"Success\",\"Transaction Action\":\"Stake 161.823311771942712767 EIGEN by 0xE08e164b...f55dDF261 on EigenLayer\",\"ERC20Transfer\":\"From 0xE08e164b...f55dDF261 To 0xaCB55C53...9D50ED8F7 For 161.823311771942712767Eigen(EIGEN)\\nFrom 0xaCB55C53...9D50ED8F7 To EigenLayer: EIGEN  Token For 161.823311771942712767Eigen(EIGEN)\\nFrom EigenLayer: EIGEN  Token To 0xaCB55C53...9D50ED8F7 For 161.823311771942712767Backing Eige...(bEIGEN)\\n\",\"TransactionFee\":\"0.001039975126212096\",\"Gas\":\"340,424 | 205,056(60.24%)\",\"InputData\":\"Function: depositIntoStrategy(address strategy, address token, uint256 amount) returns(uint256 shares)\\nstrategy=0xaCB55C530Acdb2849e6d4f36992Cd8c9D50ED8F7\\ntoken=0xec53bF9167f50cDEB3Ae105f56099aaaB9061F83\\namount=161823311771942712767\\n\",\"Nonce\":\"9\",\"PositionInBlock\":\"22\",\"TimeStamp\":\"May-15-2024 03:59:35 AM +UTC\",\"Labels\":{\"0x858646372cc42e1a627fce94aa7a7033e7cf075a\":\"0x858646372CC42E1A627fcE94aa7A7033e7CF075A\",\"0xacb55c530acdb2849e6d4f36992cd8c9d50ed8f7\":\"0xaCB55C530Acdb2849e6d4f36992Cd8c9D50ED8F7\",\"0xe08e164ba85890ac94dbeea77353d46f55ddf261\":\"0xE08e164ba85890aC94dbEEA77353d46f55dDF261\",\"0xec53bf9167f50cdeb3ae105f56099aaab9061f83\":\"0xec53bF9167f50cDEB3Ae105f56099aaaB9061F83\"}}")

	resp, err := gemini.GenerateContent(ctx, prompt)
	if err != nil {
		return fmt.Errorf("error generating content: %w", err)
	}
	rb, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		return fmt.Errorf("json.MarshalIndent: %w", err)
	}
	fmt.Fprintln(w, string(rb))
	return nil
}

func translate(ctx context.Context, content, lang string) (string, error) {
	prompt := fmt.Sprintf("You are a proficient language master, given %s(delimited by triple backticks), please translate it into %s. Remember not to alter the original meaning.", content, lang)
	cred, err := credentials.DetectDefault(&credentials.DetectOptions{
		Scopes:          []string{"https://www.googleapis.com/auth/cloud-platform"},
		CredentialsJSON: []byte(credentialsData),
	})
	if err != nil {
		return "", fmt.Errorf("failed to find default credentials: %w", err)
	}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		Backend:     genai.BackendVertexAI,
		Project:     projectID,
		Location:    location,
		Credentials: cred,
	})
	if err != nil {
		return "", err
	}

	var config = &genai.GenerateContentConfig{
		Temperature:     genai.Ptr[float32](0.6),
		MaxOutputTokens: 1000,
	}

	// Create a new Chat.
	chat, err := client.Chats.Create(ctx, model, config, nil)
	if err != nil {
		return "", err
	}
	result, err := chat.SendMessage(ctx, genai.Part{Text: prompt})
	if err != nil {
		return "", err
	}

	return result.Text(), nil
}

package BLC

import (
	"flag"
	"fmt"
	"log"
	"os"
)

type CLI struct {
}

func (cli *CLI) validateArgs(fs... *flag.FlagSet)  {
	if len(os.Args) < 2 {
		if len(fs) > 0 {
			for _,v:=range fs {
				v.Usage()
			}
		}
		os.Exit(-1)
	}
}

func (cli *CLI) printChain()  {
}
func (cli *CLI) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  getbalance -address ADDRESS - Get balance of ADDRESS")
	fmt.Println("  createblockchain -address ADDRESS - Create a blockchain and send genesis block reward to ADDRESS")
	fmt.Println("  printchain - Print all the blocks of the blockchain")
	fmt.Println("  send -from FROM -to TO -amount AMOUNT - Send AMOUNT of coins from FROM address to TO")
}

func (cli *CLI)newBlockchain(address string){
	bc:=CreateBlockchain(address)
	UTXOSet := UTXOSet{bc}
	UTXOSet.Reindex()
	fmt.Printf("\n创建了一个新的区块链,创世区块-》 %x",(*bc).Tip)

}

func (cli *CLI) getBalance(address string)  {
	if !ValidateAddress(address) {
		log.Panic("ERROR: Address is not valid")
	}
	bc := NewBlockchain(address)
	UTXOSet := UTXOSet{bc}
	balance := 0
	pubKeyHash := Base58Decode([]byte(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	UTXOs := UTXOSet.FindUTXO(pubKeyHash)

	for k, out := range UTXOs {
		fmt.Println("\n444",k,out.Value)
		balance += out.Value
	}
    fmt.Printf("\n%s : %d -- UTXOs %d ",address,balance,len(UTXOs))
}

func (cli *CLI) send(from,to string,amount int){
	if !ValidateAddress(from) {
		log.Panic("ERROR: Sender address is not valid")
	}
	if !ValidateAddress(to) {
		log.Panic("ERROR: Recipient address is not valid")
	}

	bc :=NewBlockchain(from)
	UTXOSet := UTXOSet{bc}
	tx:=NewUTXOTransaction(from,to,amount,&UTXOSet)
	cbTx := NewCoinbaseTX(from, "")
	txs := []*Transaction{cbTx, tx}
	newBlock := bc.MineBlock(txs)
	UTXOSet.Update(newBlock)
	fmt.Println("success!")
}

func (cli *CLI) Run()  {
	cli.validateArgs()
	getBalanceCmd := flag.NewFlagSet("getbalance",flag.ExitOnError)
	newBlockchainCmd := flag.NewFlagSet("newblockchain",flag.ExitOnError)
	createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send",flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	reindexUTXOCmd := flag.NewFlagSet("reindexutxo", flag.ExitOnError)
	startNodeCmd := flag.NewFlagSet("startnode", flag.ExitOnError)

	getBalanceAddress := getBalanceCmd.String("address","","获取这个地址的余额")
	createBlockchainAddress := newBlockchainCmd.String("address","","创建创世区块，并接受奖励的地址")
	sendFrom := sendCmd.String("from","","发送货币的地址")
	sendTo :=sendCmd.String("to","","接受货币的地址")
	sendAmount := sendCmd.Int("amount",0,"发送货币的数量")
	sendMine := sendCmd.Bool("mine", false, "Mine immediately on the same node")
	startNodeMiner := startNodeCmd.String("miner", "", "Enable mining mode and send reward to ADDRESS")

	switch os.Args[1] {
	case "getbalance":
		err := getBalanceCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "newblockchain":
		err := newBlockchainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "createwallet":
		err := createWalletCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "send":
		err := sendCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "reindexutxo":
		err := reindexUTXOCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "startnode":
		err := startNodeCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	default:
		cli.printUsage()
		os.Exit(1)
	}

	if getBalanceCmd.Parsed(){
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			os.Exit(1)
		}
		cli.getBalance(*getBalanceAddress)
	}

	if newBlockchainCmd.Parsed(){
		if *createBlockchainAddress == "" {
			newBlockchainCmd.Usage()
			os.Exit(1)
		}
		cli.newBlockchain(*createBlockchainAddress)
	}
	if createWalletCmd.Parsed() {
		cli.createWallet()
	}
	if printChainCmd.Parsed() {
		cli.printChain()
	}

	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
			sendCmd.Usage()
			os.Exit(1)
		}

		cli.send(*sendFrom, *sendTo, *sendAmount)
	}

}
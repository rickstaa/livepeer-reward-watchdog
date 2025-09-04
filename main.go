package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Contract: https://arbiscan.io/address/0x35Bcf3c30594191d53231E4FF333E8A770453e40
var bondingManager = common.HexToAddress("0x35Bcf3c30594191d53231E4FF333E8A770453e40")

// RoundsManager contract: https://arbiscan.io/address/0xdd6f56DcC28D3F5f27084381fE8Df634985cc39f
var roundsManager = common.HexToAddress("0xdd6f56DcC28D3F5f27084381fE8Df634985cc39f")

// connect tries to connect to one of the provided RPC URLs and returns the first
// successful connection.
func connect(rpcs []string) (*ethclient.Client, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for _, url := range rpcs {
		c, err := ethclient.DialContext(ctx, url)
		if err == nil {
			_, err2 := c.BlockNumber(ctx)
			if err2 == nil {
				return c, url, nil
			}
			c.Close()
		}
	}
	return nil, "", fmt.Errorf("all RPCs failed")
}

// sendTelegramAlert sends a message to a Telegram chat using a bot.
func sendTelegramAlert(botToken, chatID, message string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)
	payload := map[string]string{"chat_id": chatID, "text": message}
	body, _ := json.Marshal(payload)
	resp, err := http.Post(url, "application/json", strings.NewReader(string(body)))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func main() {
	// Flags for delay and notification interval
	delayFlag := flag.Duration("delay", 2*time.Hour, "Time to wait after new round before warning (e.g. 2h, 30m)")
	notifyIntervalFlag := flag.Duration("notify-interval", 0, "How often to repeat warning if reward not called (e.g. 1h, 0 for no repeat)")
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		log.Fatalf("Usage: %s <orchestrator-address> [rpc1 rpc2 ...]", os.Args[0])
	}
	orch := common.HexToAddress(args[0])

	rpcs := []string{"https://arb1.arbitrum.io/rpc"}
	if len(args) > 1 {
		rpcs = args[1:]
	}

	client, usedRPC, err := connect(rpcs)
	if err != nil {
		log.Fatalf("RPC connection failed: %v", err)
	}
	defer client.Close()
	log.Printf("Connected to %s", usedRPC)

	// Load config values from environment.
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	chatID := os.Getenv("TELEGRAM_CHAT_ID")
	if botToken == "" || chatID == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN and TELEGRAM_CHAT_ID must be set in the environment")
	}

	// Load ABIs.
	bondingABIBytes, err := os.ReadFile("ABI/BondingManager.json")
	if err != nil {
		log.Fatalf("failed to read BondingManager ABI file: %v", err)
	}
	bondingABI, err := abi.JSON(strings.NewReader(string(bondingABIBytes)))
	if err != nil {
		log.Fatalf("parse BondingManager ABI: %v", err)
	}
	rewardEvent := bondingABI.Events["Reward"]
	roundsABIBytes, err := os.ReadFile("ABI/RoundsManager.json")
	if err != nil {
		log.Fatalf("failed to read RoundsManager ABI file: %v", err)
	}
	roundsABI, err := abi.JSON(strings.NewReader(string(roundsABIBytes)))
	if err != nil {
		log.Fatalf("parse RoundsManager ABI: %v", err)
	}
	newRoundEvent := roundsABI.Events["NewRound"]

	// Subscribe to events.
	rewardCh := make(chan types.Log)
	rewardSub, err := client.SubscribeFilterLogs(context.Background(), ethereum.FilterQuery{
		Addresses: []common.Address{bondingManager},
		Topics: [][]common.Hash{
			{rewardEvent.ID},
			{common.BytesToHash(orch.Bytes())},
		},
	}, rewardCh)
	if err != nil {
		log.Fatalf("Subscribe error (Reward): %v", err)
	}
	roundCh := make(chan types.Log)
	roundSub, err := client.SubscribeFilterLogs(context.Background(), ethereum.FilterQuery{
		Addresses: []common.Address{roundsManager},
		Topics: [][]common.Hash{
			{newRoundEvent.ID},
		},
	}, roundCh)
	if err != nil {
		log.Fatalf("Subscribe error (NewRound): %v", err)
	}

	// Monitor rounds and rewards and send alerts.
	var currentRound uint64
	var roundStart time.Time
	rewardCalled := false
	log.Println("Monitoring started...")
	ticker := time.NewTicker(*notifyIntervalFlag)
	defer ticker.Stop()
	for {
		select {
		case err := <-rewardSub.Err():
			log.Printf("Reward subscription error: %v", err)
			sendTelegramAlert(botToken, chatID, fmt.Sprintf("⚠️ Reward subscription error: %v", err))
			return
		case err := <-roundSub.Err():
			log.Printf("NewRound subscription error: %v", err)
			sendTelegramAlert(botToken, chatID, fmt.Sprintf("⚠️ NewRound subscription error: %v", err))
			return
		case vLog := <-rewardCh:
			// Reward called for this round.
			rewardCalled = true
			alertMsg := fmt.Sprintf("✅ Reward called for %s at block %d, tx %s", orch.Hex(), vLog.BlockNumber, vLog.TxHash.Hex())
			log.Println(alertMsg)
			sendTelegramAlert(botToken, chatID, alertMsg)
		case vLog := <-roundCh:
			// New round started.
			var roundNum uint64
			if len(vLog.Topics) > 1 {
				roundNum = vLog.Topics[1].Big().Uint64()
			}
			currentRound = roundNum
			roundStart = time.Now()
			rewardCalled = false
			log.Printf("New round %d started", currentRound)
		case <-ticker.C:
			if !rewardCalled && !roundStart.IsZero() {
				elapsed := time.Since(roundStart)
				if elapsed >= *delayFlag {
					alertMsg := fmt.Sprintf("❌ No reward called for %s in round %d after %s", orch.Hex(), currentRound, delayFlag.String())
					log.Println(alertMsg)
					sendTelegramAlert(botToken, chatID, alertMsg)
				}
			}
		}
	}
}

package main

import (
	"context"
	"fmt"
	"os"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: loader <json_file_path>")
		os.Exit(1)
	}

	filePath := os.Args[1]
	fmt.Printf("Loading file: %s ...\n", filePath)

	start := time.Now()

	// Initialize Parser via Wire
	ctx := context.Background()
	l, cleanup, err := InitializeParser(ctx)
	if err != nil {
		fmt.Printf("Error initializing loader: %v\n", err)
		os.Exit(1)
	}
	defer cleanup()

	// Load Data
	data, err := l.LoadExtractedJSON(context.Background(), filePath)
	if err != nil {
		fmt.Printf("Error loading file: %v\n", err)
		os.Exit(1)
	}

	duration := time.Since(start)

	// Print Stats
	fmt.Println("--------------------------------------------------")
	fmt.Printf("Load Complete in %s\n", duration)
	fmt.Println("--------------------------------------------------")
	fmt.Printf("Quests:          %d\n", len(data.Quests))
	fmt.Printf("Dialogue Groups: %d\n", len(data.DialogueGroups))
	fmt.Printf("NPCs:            %d\n", len(data.NPCs))
	fmt.Printf("Items:           %d\n", len(data.Items))
	fmt.Printf("Magic:           %d\n", len(data.Magic))
	fmt.Printf("Locations:       %d\n", len(data.Locations))
	fmt.Printf("Cells:           %d\n", len(data.Cells))
	fmt.Printf("System:          %d\n", len(data.System))
	fmt.Printf("Messages:        %d\n", len(data.Messages))
	fmt.Printf("Load Screens:    %d\n", len(data.LoadScreens))
	fmt.Println("--------------------------------------------------")
}

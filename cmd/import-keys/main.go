// import-keys is the operator tool for refilling the license-key pool.
//
//	go run ./cmd/import-keys <product_id> <path-to-keys.txt>
//
// The file is one key per line; blank lines and leading/trailing whitespace are
// ignored. Duplicates (against either the pool or already-issued keys) are
// skipped without erroring. The SQLite path comes from the same SQLITE_PATH env
// var the server uses (default ./data/auranewweb.sqlite), so this hits the same
// database the running app reads from. Stop the server before importing only
// if your platform's SQLite-locking story makes you nervous; WAL mode tolerates
// concurrent readers and one writer, so a running app is generally fine.
package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"

	"aura-optimizer/internal/db"
	"aura-optimizer/internal/repo"
)

func main() {
	log.SetFlags(0)

	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: import-keys <product_id> <path-to-keys.txt>")
		fmt.Fprintln(os.Stderr, "example: import-keys lifetime ./keys.txt")
		os.Exit(2)
	}
	productID := os.Args[1]
	keysPath := os.Args[2]

	keys, err := readKeysFile(keysPath)
	if err != nil {
		log.Fatalf("read keys file: %v", err)
	}
	if len(keys) == 0 {
		log.Fatalf("no keys found in %s", keysPath)
	}

	dbPath := os.Getenv("SQLITE_PATH")
	if dbPath == "" {
		dbPath = "./data/auranewweb.sqlite"
	}
	sqlDB, err := db.Open(dbPath)
	if err != nil {
		log.Fatalf("open db at %s: %v", dbPath, err)
	}
	defer sqlDB.Close()

	licRepo := repo.NewLicenseRepository(sqlDB)
	ctx := context.Background()

	before, _ := licRepo.CountAvailable(ctx, productID)

	added, skipped, err := licRepo.AddAvailable(ctx, productID, keys)
	if err != nil {
		log.Fatalf("add to pool: %v", err)
	}

	after, _ := licRepo.CountAvailable(ctx, productID)

	fmt.Printf("Pool for product %q: %d → %d keys (added %d, skipped %d duplicates/blanks)\n",
		productID, before, after, added, skipped)
}

func readKeysFile(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var keys []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		keys = append(keys, sc.Text())
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return keys, nil
}

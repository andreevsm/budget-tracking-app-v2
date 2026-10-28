package main

import (
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg, err := pgxpool.ParseConfig(os.Args[1])
	if err != nil {
		fmt.Println("PARSE ERROR:", err)
		return
	}
	cc := cfg.ConnConfig
	fmt.Printf("host=%q port=%d db=%q user=%q password=%q\n",
		cc.Host, cc.Port, cc.Database, cc.User, cc.Password)
}

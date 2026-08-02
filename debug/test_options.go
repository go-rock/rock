package main

import (
	"fmt"
	"github.com/go-rock/rock/trie"
)

func main() {
	fmt.Println("=== 测试默认选项 ===")
	t1 := trie.New()
	
	node1 := t1.Define("/test/")
	node1.Handle("GET", func() {})
	
	fmt.Println("默认 Trie - 请求 /test:")
	res1 := t1.Match("/test")
	fmt.Printf("Node: %v, TSR: %s, FPR: %s\n", res1.Node, res1.TSR, res1.FPR)
	
	fmt.Println("\n=== 测试显式启用重定向选项 ===")
	t2 := trie.New(trie.Options{
		TrailingSlashRedirect: true,
		FixedPathRedirect:     true,
	})
	
	node2 := t2.Define("/test/")
	node2.Handle("GET", func() {})
	
	fmt.Println("显式选项 Trie - 请求 /test:")
	res2 := t2.Match("/test")
	fmt.Printf("Node: %v, TSR: %s, FPR: %s\n", res2.Node, res2.TSR, res2.FPR)
}
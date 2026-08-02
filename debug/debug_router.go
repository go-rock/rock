package main

import (
	"fmt"
	"github.com/go-rock/rock"
	"github.com/go-rock/rock/trie"
)

func main() {
	// 创建一个简单的 trie 来测试
	t := trie.New()
	
	// 定义路由
	node1 := t.Define("/test/")
	node1.Handle("GET", func() {})
	
	node2 := t.Define("/test2")
	node2.Handle("GET", func() {})
	
	// 测试匹配
	fmt.Println("=== 测试 /test ===")
	res1 := t.Match("/test")
	fmt.Printf("Node: %v\n", res1.Node)
	fmt.Printf("TSR: %s\n", res1.TSR)
	fmt.Printf("FPR: %s\n", res1.FPR)
	fmt.Printf("Params: %v\n", res1.Params)
	
	fmt.Println("\n=== 测试 /test/ ===")
	res2 := t.Match("/test/")
	fmt.Printf("Node: %v\n", res2.Node)
	if res2.Node != nil {
		fmt.Printf("Node Pattern: %s\n", res2.Node.GetPattern())
	}
	fmt.Printf("TSR: %s\n", res2.TSR)
	fmt.Printf("FPR: %s\n", res2.FPR)
	fmt.Printf("Params: %v\n", res2.Params)
	
	fmt.Println("\n=== 测试 /test2 ===")
	res3 := t.Match("/test2")
	fmt.Printf("Node: %v\n", res3.Node)
	if res3.Node != nil {
		fmt.Printf("Node Pattern: %s\n", res3.Node.GetPattern())
	}
	fmt.Printf("TSR: %s\n", res3.TSR)
	fmt.Printf("FPR: %s\n", res3.FPR)
	fmt.Printf("Params: %v\n", res3.Params)
	
	fmt.Println("\n=== 测试 /test2/ ===")
	res4 := t.Match("/test2/")
	fmt.Printf("Node: %v\n", res4.Node)
	fmt.Printf("TSR: %s\n", res4.TSR)
	fmt.Printf("FPR: %s\n", res4.FPR)
	fmt.Printf("Params: %v\n", res4.Params)
}
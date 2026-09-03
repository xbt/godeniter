// Package router 提供了纯 Go 标准库实现的高性能 Trie（前缀树）路由器。
// 支持 RESTful HTTP 动词匹配、命名动态路由参数 (:param)、全匹配通配符 (*filepath)、路由分组及分组级中间件机制。
package router

import (
	"strings"
)

// Params 是从动态路由中提取出的键值对映射表，例如 map[string]string{"id": "123"}。
type Params map[string]string

// Get 获取指定名称的路由参数值。
func (p Params) Get(key string) string {
	if p == nil {
		return ""
	}
	return p[key]
}

// node 代表 Trie 前缀树上的一个节点。
type node struct {
	pattern  string           // 完整的路由注册路径（仅在有效路由的叶子/终点节点上有值，如 "/users/:id"）
	part     string           // 当前节点对应的路径片段（如 "users" 或 ":id" 或 "*filepath"）
	children []*node          // 子节点列表
	isWild   bool             // 是否为动态节点（包含 ':' 或 '*' 时为 true）
	handlers []interface{}    // 挂载在该路由上的处理函数与中间件链
}

// matchChild 查找第一个匹配 part 的子节点（优先精准匹配，其次动态匹配）。
func (n *node) matchChild(part string) *node {
	for _, child := range n.children {
		if child.part == part || child.isWild {
			return child
		}
	}
	return nil
}

// matchChildren 查找所有可能匹配 part 的子节点。
func (n *node) matchChildren(part string) []*node {
	nodes := make([]*node, 0)
	for _, child := range n.children {
		if child.part == part || child.isWild {
			nodes = append(nodes, child)
		}
	}
	return nodes
}

// insert 递归向 Trie 树中插入一个路由规则。
// pattern: 完整路由字符串，如 "/user/:id"
// parts: 切分后的片段列表，如 ["user", ":id"]
// height: 当前递归遍历的层级深度
// handlers: 挂载到该端点的处理函数列表
func (n *node) insert(pattern string, parts []string, height int, handlers []interface{}) {
	if len(parts) == height {
		n.pattern = pattern
		n.handlers = handlers
		return
	}

	part := parts[height]
	child := n.matchChild(part)
	if child == nil {
		// 创建新节点，如果以 ':' 或 '*' 开头则标记为动态通配节点
		child = &node{
			part:   part,
			isWild: len(part) > 0 && (part[0] == ':' || part[0] == '*'),
		}
		n.children = append(n.children, child)
	}

	child.insert(pattern, parts, height+1, handlers)
}

// search 递归在 Trie 树中搜索匹配请求路径的节点。
// parts: 请求路径切分后的片段列表
// height: 当前递归层级
func (n *node) search(parts []string, height int) *node {
	// 如果到达路径末尾，或者当前节点是 '*' 全局通配符，返回当前节点
	if len(parts) == height || strings.HasPrefix(n.part, "*") {
		if n.pattern == "" {
			return nil
		}
		return n
	}

	part := parts[height]
	children := n.matchChildren(part)

	for _, child := range children {
		result := child.search(parts, height+1)
		if result != nil {
			return result
		}
	}

	return nil
}

// parsePattern 将路径按 '/' 分隔为片段切片，并处理特殊通配符。
// 例如: "/users/:id" -> ["users", ":id"]
// 例如: "/static/*filepath/more" -> ["static", "*filepath"] (* 后面内容作为整体捕获)
func parsePattern(pattern string) []string {
	vs := strings.Split(pattern, "/")
	parts := make([]string, 0)
	for _, item := range vs {
		if item != "" {
			parts = append(parts, item)
			if item[0] == '*' {
				// '*' 通配符只允许匹配一次，捕获后续所有内容
				break
			}
		}
	}
	return parts
}

package engine

import "strings"

// 前缀树
type treeNode struct {
	pattern  string      // 待匹配的路由，例如/p/:lang
	part     string      // 本节点存储的字符，路由中的一部分，例如 :lang
	children []*treeNode // 子节点
	isWild   bool        // 是否精确匹配，part含有:或者*时为true
}

// 从子节点中查找是否有匹配的
func (n *treeNode) mathChildNode(part string) *treeNode {
	for _, child := range n.children {
		if child.part == part || child.isWild {
			return child
		}
	}
	return nil
}

// 查找子节点集合
func (n *treeNode) matchChildrenNodes(part string) []*treeNode {
	nodes := make([]*treeNode, 0)
	for _, child := range n.children {
		if child.part == part || child.isWild {
			nodes = append(nodes, child)
		}
	}
	return nodes
}

// api/v1/user
// api/v2/user
// insert 递归插入节点，从顶层往下层，层层查找并插入
func (n *treeNode) insert(pattern string, parts []string, height int) {
	if len(parts) == height {
		n.pattern = pattern
		return
	}
	part := parts[height]
	child := n.mathChildNode(part)
	// 如果从子节点中没找到，说明没有，需要创建
	if child == nil {
		child = &treeNode{
			part:   part,
			isWild: part[0] == ':' || part[0] == '*',
		}
		n.children = append(n.children, child)
	}
	child.insert(pattern, parts, height+1)
}

// 递归查询节点,从顶层往下层查找
func (n *treeNode) search(parts []string, height int) *treeNode {
	if len(parts) == height || strings.HasPrefix(n.part, "*") {
		if n.pattern == "" {
			return nil
		}
		return n
	}
	part := parts[height]
	children := n.matchChildrenNodes(part)
	for _, child := range children {
		result := child.search(parts, height+1) // todo 改动？
		if result != nil {
			return result
		}
	}
	return nil
}

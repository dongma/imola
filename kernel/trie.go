package kernel

import (
	"errors"
	"strings"
)

// tree代表树结构，root表示树的根节点
type Tree struct {
	root *node
}

// 表示tree树中的node节点，包含子节点列表、segment、末尾节点信息
type node struct {
	isLast  bool              // 代表这个节点是否可以成为最终的路由规则，该节点是否能成为一个独立的url
	segment string            // uri中额字符串，代表这个节点表示的路由中某个段的字符串
	handler ControllerHandler // 此节点包含的控制器
	childs  []*node           // 包含的叶子节点
}

func newNode() *node {
	return &node{
		isLast:  false,
		segment: "",
		childs:  []*node{},
	}
}

func NewTree() *Tree {
	root := newNode()
	return &Tree{root}
}

// 判断一个segment是否是通用segment，即以:开头
func isWildSegment(segment string) bool {
	return strings.HasPrefix(segment, ":")
}

func (n *node) filterChildNodes(segment string) []*node {
	if len(n.childs) == 0 {
		return nil
	}

	// 如果segment是通配符，则所有下一层节点都满足要求
	if isWildSegment(segment) {
		return n.childs
	}

	nodes := make([]*node, 0, len(n.childs))
	// 过滤所有的下一层子节点
	for _, cnode := range n.childs {
		if isWildSegment(cnode.segment) {
			// 若下一层子节点有通配符，则满足要求
			nodes = append(nodes, cnode)
		} else if cnode.segment == segment {
			// 若下一层子节点没有通配符，但是文本完全匹配，则满足要求
			nodes = append(nodes, cnode)
		}
	}
	return nodes
}

// 判读路由是否已经在节点的所有子节点树中存在了
func (n *node) matchNode(url string) *node {
	// 使用分隔符将url切割为两部分
	segments := strings.SplitN(url, "/", 2)
	// 第一部分用于匹配下一层子节点
	segment := segments[0]
	if !isWildSegment(segment) {
		segment = strings.ToUpper(segment)
	}

	// 匹配符合的下一层子节点，当前子节点没有一个符合，那么说明这个url一定是之前不存在。直接返回了nil
	cnodes := n.filterChildNodes(segment)
	if cnodes == nil || len(cnodes) == 0 {
		return nil
	}

	// 如果只有一个segment，则是最后一个标记，则判断这些cnode是否有isLast标签
	if len(segments) == 1 {
		for _, tn := range cnodes {
			if tn.isLast {
				return tn
			}
		}
		return nil
	}

	// 如果有2个segment，递归每个子节点进行查找
	for _, tn := range cnodes {
		tnMatch := tn.matchNode(segments[1])
		if tnMatch != nil {
			return tnMatch
		}
	}
	return nil
}

// 增加路由节点, 路由节点有先后顺序，示例路由url如下:
/*
/book/list
/book/:id (冲突)
/book/:id/name
/book/:student/age
/:user/name
/:user/name/:age (冲突)
*/
func (tree *Tree) AddRouter(url string, handler ControllerHandler) error {
	n := tree.root
	if n.matchNode(url) != nil {
		return errors.New("route exists: " + url)
	}

	segments := strings.Split(url, "/")
	for index, segment := range segments {
		if !isWildSegment(segment) {
			segment = strings.ToUpper(segment)
		}
		isLast := index == len(segments)-1
		var objNode *node // 标记是否有合适的子节点
		childNodes := n.filterChildNodes(segment)

		// 若有匹配的子节点，则选择这个子节点
		if len(childNodes) > 0 {
			for _, cnode := range childNodes {
				if cnode.segment == segment {
					objNode = cnode
					break
				}
			}
		}

		// 当objNode为空时，则创建一个Node节点
		if objNode == nil {
			cnode := newNode()
			cnode.segment = segment
			if isLast {
				cnode.isLast = true
				cnode.handler = handler
			}
			n.childs = append(n.childs, cnode)
			objNode = cnode
		}
		n = objNode
	}
	return nil
}

// 匹配url，直接复用matchNode()函数，url是不带通配符的地址
func (tree *Tree) FindHandler(url string) ControllerHandler {
	matchNode := tree.root.matchNode(url)
	if matchNode == nil {
		return nil
	}
	return matchNode.handler
}

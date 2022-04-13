/*
 * @Descripation: 队列和栈
 * @Date: 2021-11-04 17:25:27
 */

package srt

// Queue 队列
type Queue struct {
	nodes []interface{}
}

// NewQueue 创建空队列
func NewQueue() *Queue {
	return &Queue{
		nodes: []interface{}{},
	}
}

// Push 入队
func (q *Queue) Push(node interface{}) {
	q.nodes = append(q.nodes, node)
}

// Pop 出队并删除队首
func (q *Queue) Pop() (node interface{}) {
	node = q.nodes[0]
	q.nodes = q.nodes[1:]
	return node
}

// Empty 判断队列是否为空
func (q *Queue) Empty() bool {
	return len(q.nodes) == 0
}

// Stack 栈
type Stack struct {
	nodes []interface{}
}

// NewStack 创建空栈
func NewStack() *Stack {
	return &Stack{
		nodes: []interface{}{},
	}
}

// Push 入栈
func (s *Stack) Push(node interface{}) {
	s.nodes = append(s.nodes, node)
}

// Pop 出栈并删除栈顶
func (s *Stack) Pop() (node interface{}) {
	l := len(s.nodes)
	node = s.nodes[l-1]
	s.nodes = s.nodes[:l-1]
	return node
}

// Empty 判断栈是否为空
func (s *Stack) Empty() bool {
	return len(s.nodes) == 0
}

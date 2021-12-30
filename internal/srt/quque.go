/*
 * @Descripation: 队列和栈
 * @Date: 2021-11-04 17:25:27
 */

package srt

/**
 * @description: 队列
 */
type Queue struct {
	nodes []interface{}
}

/**
 * @description: 创建空队列
 * @return {*Queue} 空队列
 */
func NewQueue() *Queue {
	return &Queue{
		nodes: []interface{}{},
	}
}

/**
 * @description: 入队
 * @param {interface{}} node 入队元素
 */
func (q *Queue) Push(node interface{}) {
	q.nodes = append(q.nodes, node)
}

/**
 * @description: 出队并删除队首
 * @return {interface{}} 队首元素
 */
func (q *Queue) Pop() (node interface{}) {
	node = q.nodes[0]
	q.nodes = q.nodes[1:]
	return node
}

/**
 * @description: 判断队列是否为空
 * @return {bool} 队列为空返回true
 */
func (q *Queue) Empty() bool {
	return len(q.nodes) == 0
}

/**
 * @description: 栈
 */
type Stack struct {
	nodes []interface{}
}

/**
 * @description: 创建空栈
 * @return {*Stack} 空栈
 */
func NewStack() *Stack {
	return &Stack{
		nodes: []interface{}{},
	}
}

/**
 * @description: 入栈
 * @param {interface{}} node 入栈元素
 */
func (s *Stack) Push(node interface{}) {
	s.nodes = append(s.nodes, node)
}

/**
 * @description: 出栈并删除栈顶
 * @return {interface{}} 栈顶元素
 */
func (s *Stack) Pop() (node interface{}) {
	l := len(s.nodes)
	node = s.nodes[l-1]
	s.nodes = s.nodes[:l-1]
	return node
}

/**
 * @description: 判断栈是否为空
 * @return {bool} 栈为空返回true
 */
func (s *Stack) Empty() bool {
	return len(s.nodes) == 0
}

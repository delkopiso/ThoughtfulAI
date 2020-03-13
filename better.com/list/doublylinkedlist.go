package list

type Node struct {
	Value int
	Left  *Node
	Right *Node
}

type DoublyLinkedList struct {
	count int
	head  *Node
	tail  *Node
}

func NewDoublyLinkedList() *DoublyLinkedList {
	return &DoublyLinkedList{}
}

func (l *DoublyLinkedList) Push(item int) {
	defer func() {
		l.count += 1
	}()

	if l.count == 0 {
		l.head = &Node{Value: item}
		l.tail = l.head
		return
	}

	right := l.head
	newRoot := &Node{
		Value: item,
		Left:  nil,
		Right: right,
	}
	right.Left = newRoot
	l.head = newRoot
}

func (l *DoublyLinkedList) Pop() int {
	if l.count == 0 {
		return -1
	}
	defer func() {
		l.count -= 1
	}()
	i := l.tail.Value
	newTail := l.tail.Left
	if newTail != nil {
		newTail.Right = nil
	}
	l.tail = newTail
	return i
}

func (l *DoublyLinkedList) Count() int {
	return l.count
}

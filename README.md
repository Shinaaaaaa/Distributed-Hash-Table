# Distributed-Hash-Table
## Chord协议
 ### Chrodnode.go 
- 实现 interface 的接口
 ### utility.go
- 实现 SHA1 , ping , dial 等辅助函数
 ### node.go
```
type Node struct {
	Self              IP
	Data              map[string]string
	SuccessorList     [successorListLen + 1]IP
	FingerTable       [M + 1]IP
	Predecessor       *IP
	InNet             bool
	PreData           map[string]string
	DataLock          sync.Mutex
	PreDataLock       sync.Mutex
	SuccessorListLock sync.Mutex
	StabilizeLock     sync.Mutex
	FingerTableLock   sync.Mutex
}
```
- self 记录该节点信息
- Data 记录储存在该节点数据
- SuccessorList 长度为20的后继表
- Finger Table 长度为160
- PreMap 备份前一个节点的数据，处理 Force Quit
###### 主要实现函数：
- GetTargetSuccessor 找到key对应的节点
- stabilize Notify checkPre fixFingerTable 进行维护
- 数据储存并备份到后一个节点

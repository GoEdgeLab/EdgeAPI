package services

import (
	"context"
	"fmt"
	teaconst "github.com/TeaOSLab/EdgeAPI/internal/const"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/goman"
	"github.com/TeaOSLab/EdgeAPI/internal/remotelogs"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/logs"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

var primaryNodeId int64 = 0

// CommandRequest 命令请求相关
type CommandRequest struct {
	Id          int64
	Code        string
	CommandJSON []byte
}

type CommandRequestWaiting struct {
	Timestamp int64
	Chan      chan *pb.NodeStreamMessage
}

func (this *CommandRequestWaiting) Close() {
	defer func() {
		_ = recover()
	}()

	close(this.Chan)
}

var nodeResponseChanMap = map[int64]*CommandRequestWaiting{} // request id => response
var commandRequestId = int64(0)

var nodeLocker = &sync.Mutex{}
var nodeRequestChanMap = map[int64]chan *CommandRequest{} // node id => chan

func NextCommandRequestId() int64 {
	return atomic.AddInt64(&commandRequestId, 1)
}

func init() {
	// 清理WaitingChannelMap
	var ticker = time.NewTicker(30 * time.Second)
	goman.New(func() {
		for range ticker.C {
			nodeLocker.Lock()
			for requestId, request := range nodeResponseChanMap {
				if time.Now().Unix()-request.Timestamp > 3600 {
					nodeResponseChanMap[requestId].Close()
					delete(nodeResponseChanMap, requestId)
				}
			}
			nodeLocker.Unlock()
		}
	})
}

// NodeStream 节点stream
func (this *NodeService) NodeStream(server pb.NodeService_NodeStreamServer) error {
	// TODO 使用此stream快速通知边缘节点更新
	// 校验节点
	_, _, nodeId, err := rpcutils.ValidateRequest(server.Context(), rpcutils.UserTypeNode)
	if err != nil {
		return err
	}

	// 选择一个作为主节点
	if primaryNodeId == 0 {
		primaryNodeId = nodeId
	}

	defer func() {
		// 修改当前API节点的主边缘节点
		/// TODO 每个集群应该有一个primaryNodeId
		if primaryNodeId == nodeId {
			primaryNodeId = 0

			nodeLocker.Lock()
			if len(nodeRequestChanMap) > 0 {
				for anotherNodeId := range nodeRequestChanMap {
					primaryNodeId = anotherNodeId
					break
				}
			}
			nodeLocker.Unlock()
		}

		// 修改在线状态
		err = models.SharedNodeDAO.UpdateNodeActive(nil, nodeId, false)
		if err != nil {
			remotelogs.Error("NODE_SERVICE", "change node active failed: "+err.Error())
		}
	}()

	// 设置API节点
	err = models.SharedNodeDAO.UpdateNodeConnectedAPINodes(nil, nodeId, []int64{teaconst.NodeId})
	if err != nil {
		return err
	}

	var tx = this.NullTx()

	// 是否发送恢复通知
	oldIsActive, err := models.SharedNodeDAO.FindNodeActive(tx, nodeId)
	if err != nil {
		return err
	}
	if !oldIsActive {
		inactiveNotifiedAt, err := models.SharedNodeDAO.FindNodeInactiveNotifiedAt(tx, nodeId)
		if err != nil {
			return err
		}

		// 设置为活跃
		err = models.SharedNodeDAO.UpdateNodeActive(tx, nodeId, true)
		if err != nil {
			return err
		}

		if inactiveNotifiedAt > 0 {
			// 发送恢复消息
			clusterId, err := models.SharedNodeDAO.FindNodeClusterId(tx, nodeId)
			if err != nil {
				return err
			}
			nodeName, err := models.SharedNodeDAO.FindNodeName(tx, nodeId)
			if err != nil {
				return err
			}
			var subject = "节点\"" + nodeName + "\"已经恢复在线"
			var msg = "节点\"" + nodeName + "\"已经恢复在线"
			err = models.SharedMessageDAO.CreateNodeMessage(tx, nodeconfigs.NodeRoleNode, clusterId, nodeId, models.MessageTypeNodeActive, models.MessageLevelSuccess, subject, msg, nil, false)
			if err != nil {
				return err
			}
		}
	}

	nodeLocker.Lock()
	requestChan, ok := nodeRequestChanMap[nodeId]
	if !ok {
		requestChan = make(chan *CommandRequest, 1024)
		nodeRequestChanMap[nodeId] = requestChan
	}
	nodeLocker.Unlock()

	defer func() {
		nodeLocker.Lock()
		delete(nodeRequestChanMap, nodeId)
		nodeLocker.Unlock()
	}()

	// 发送请求
	goman.New(func() {
		for {
			select {
			case <-server.Context().Done():
				return
			case commandRequest := <-requestChan:
				// logs.Println("[RPC]sending command '" + commandRequest.Code + "' to node '" + strconv.FormatInt(nodeId, 10) + "'")
				retries := 3 // 错误重试次数
				for i := 0; i < retries; i++ {
					err := server.Send(&pb.NodeStreamMessage{
						RequestId: commandRequest.Id,
						Code:      commandRequest.Code,
						DataJSON:  commandRequest.CommandJSON,
					})
					if err != nil {
						if i == retries-1 {
							logs.Println("[RPC]send command '" + commandRequest.Code + "' failed: " + err.Error())
						} else {
							time.Sleep(1 * time.Second)
						}
					} else {
						break
					}
				}
			}
		}
	})

	// 接受请求
	for {
		req, err := server.Recv()
		if err != nil {
			// 修改节点状态
			err1 := models.SharedNodeDAO.UpdateNodeIsActive(tx, nodeId, false)
			if err1 != nil {
				logs.Println(err1.Error())
			}

			return err
		}

		func(req *pb.NodeStreamMessage) {
			// 因为 responseChan.Chan 有被关闭的风险，所以我们使用recover防止panic
			defer func() {
				_ = recover()
			}()

			nodeLocker.Lock()
			responseChan, ok := nodeResponseChanMap[req.RequestId]
			if ok {
				select {
				case responseChan.Chan <- req:
				default:

				}
			}
			nodeLocker.Unlock()
		}(req)
	}
}

// SendCommandToNode 向节点发送命令
func (this *NodeService) SendCommandToNode(ctx context.Context, req *pb.NodeStreamMessage) (*pb.NodeStreamMessage, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	nodeId := req.NodeId
	if nodeId <= 0 {
		return nil, errors.New("node id should not be less than 0")
	}

	return SendCommandToNode(req.NodeId, req.RequestId, req.Code, req.DataJSON, req.TimeoutSeconds, true)
}

// SendCommandToNode 向节点发送命令
func SendCommandToNode(nodeId int64, requestId int64, messageCode string, dataJSON []byte, timeoutSeconds int32, forceConnecting bool) (result *pb.NodeStreamMessage, err error) {
	nodeLocker.Lock()
	requestChan, ok := nodeRequestChanMap[nodeId]
	nodeLocker.Unlock()

	if !ok {
		if forceConnecting {
			return &pb.NodeStreamMessage{
				RequestId: requestId,
				IsOk:      false,
				Message:   "node '" + strconv.FormatInt(nodeId, 10) + "' not connected yet",
			}, nil
		} else {
			return &pb.NodeStreamMessage{
				RequestId: requestId,
				IsOk:      true,
			}, nil
		}
	}

	requestId = NextCommandRequestId()

	select {
	case requestChan <- &CommandRequest{
		Id:          requestId,
		Code:        messageCode,
		CommandJSON: dataJSON,
	}:
		// 加入到等待队列中
		respChan := make(chan *pb.NodeStreamMessage, 1)
		waiting := &CommandRequestWaiting{
			Timestamp: time.Now().Unix(),
			Chan:      respChan,
		}

		nodeLocker.Lock()
		nodeResponseChanMap[requestId] = waiting
		nodeLocker.Unlock()

		// 等待响应
		if timeoutSeconds <= 0 {
			timeoutSeconds = 10
		}
		timeout := time.NewTimer(time.Duration(timeoutSeconds) * time.Second)
		select {
		case resp := <-respChan:
			// 从队列中删除
			nodeLocker.Lock()
			delete(nodeResponseChanMap, requestId)
			waiting.Close()
			nodeLocker.Unlock()

			if resp == nil {
				return &pb.NodeStreamMessage{
					RequestId: requestId,
					Code:      messageCode,
					Message:   "response timeout",
					IsOk:      false,
				}, nil
			}

			return resp, nil
		case <-timeout.C:
			// 从队列中删除
			nodeLocker.Lock()
			delete(nodeResponseChanMap, requestId)
			waiting.Close()
			nodeLocker.Unlock()

			return &pb.NodeStreamMessage{
				RequestId: requestId,
				Code:      messageCode,
				Message:   "response timeout over " + fmt.Sprintf("%d", timeoutSeconds) + " seconds",
				IsOk:      false,
			}, nil
		}
	default:
		return &pb.NodeStreamMessage{
			RequestId: requestId,
			Code:      messageCode,
			Message:   "command queue is full over " + strconv.Itoa(len(requestChan)),
			IsOk:      false,
		}, nil
	}
}

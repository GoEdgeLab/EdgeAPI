package services

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/TeaOSLab/EdgeAPI/internal/configs"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/messageconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/logs"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

// 命令请求相关
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
		recover()
	}()

	close(this.Chan)
}

var responseChanMap = map[int64]*CommandRequestWaiting{} // request id => response
var commandRequestId = int64(0)

var nodeLocker = &sync.Mutex{}
var requestChanMap = map[int64]chan *CommandRequest{} // node id => chan

func NextCommandRequestId() int64 {
	return atomic.AddInt64(&commandRequestId, 1)
}

func init() {
	// 清理WaitingChannelMap
	ticker := time.NewTicker(30 * time.Second)
	go func() {
		for range ticker.C {
			nodeLocker.Lock()
			for requestId, request := range responseChanMap {
				if time.Now().Unix()-request.Timestamp > 3600 {
					responseChanMap[requestId].Close()
					delete(responseChanMap, requestId)
				}
			}
			nodeLocker.Unlock()
		}
	}()
}

// 节点stream
func (this *NodeService) NodeStream(server pb.NodeService_NodeStreamServer) error {
	// TODO 使用此stream快速通知边缘节点更新
	// 校验节点
	_, nodeId, err := rpcutils.ValidateRequest(server.Context(), rpcutils.UserTypeNode)
	if err != nil {
		return err
	}

	// 返回连接成功
	{
		apiConfig, err := configs.SharedAPIConfig()
		if err != nil {
			return err
		}
		connectedMessage := &messageconfigs.ConnectedAPINodeMessage{APINodeId: apiConfig.NumberId()}
		connectedMessageJSON, err := json.Marshal(connectedMessage)
		if err != nil {
			return errors.Wrap(err)
		}
		err = server.Send(&pb.NodeStreamMessage{
			Code:     messageconfigs.MessageCodeConnectedAPINode,
			DataJSON: connectedMessageJSON,
		})
		if err != nil {
			return err
		}
	}

	logs.Println("[RPC]accepted node '" + strconv.FormatInt(nodeId, 10) + "' connection")

	nodeLocker.Lock()
	requestChan, ok := requestChanMap[nodeId]
	if !ok {
		requestChan = make(chan *CommandRequest, 1024)
		requestChanMap[nodeId] = requestChan
	}
	nodeLocker.Unlock()

	// 发送请求
	go func() {
		for {
			select {
			case <-server.Context().Done():
				return
			case commandRequest := <-requestChan:
				logs.Println("[RPC]sending command '" + commandRequest.Code + "' to node '" + strconv.FormatInt(nodeId, 10) + "'")
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
	}()

	// 接受请求
	for {
		req, err := server.Recv()
		if err != nil {
			// 修改节点状态
			err1 := models.SharedNodeDAO.UpdateNodeIsActive(nodeId, false)
			if err1 != nil {
				logs.Println(err1.Error())
			}

			return err
		}

		func(req *pb.NodeStreamMessage) {
			// 因为 responseChan.Chan 有被关闭的风险，所以我们使用recover防止panic
			defer func() {
				recover()
			}()

			nodeLocker.Lock()
			responseChan, ok := responseChanMap[req.RequestId]
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

// 向节点发送命令
func (this *NodeService) SendCommandToNode(ctx context.Context, req *pb.NodeStreamMessage) (*pb.NodeStreamMessage, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	nodeId := req.NodeId
	if nodeId <= 0 {
		return nil, errors.New("node id should not be less than 0")
	}

	nodeLocker.Lock()
	requestChan, ok := requestChanMap[nodeId]
	nodeLocker.Unlock()

	if !ok {
		return &pb.NodeStreamMessage{
			RequestId: req.RequestId,
			IsOk:      false,
			Message:   "node '" + strconv.FormatInt(nodeId, 10) + "' not connected yet",
		}, nil
	}

	req.RequestId = NextCommandRequestId()

	select {
	case requestChan <- &CommandRequest{
		Id:          req.RequestId,
		Code:        req.Code,
		CommandJSON: req.DataJSON,
	}:
		// 加入到等待队列中
		respChan := make(chan *pb.NodeStreamMessage, 1)
		waiting := &CommandRequestWaiting{
			Timestamp: time.Now().Unix(),
			Chan:      respChan,
		}

		nodeLocker.Lock()
		responseChanMap[req.RequestId] = waiting
		nodeLocker.Unlock()

		// 等待响应
		timeoutSeconds := req.TimeoutSeconds
		if timeoutSeconds <= 0 {
			timeoutSeconds = 10
		}
		timeout := time.NewTimer(time.Duration(timeoutSeconds) * time.Second)
		select {
		case resp := <-respChan:
			// 从队列中删除
			nodeLocker.Lock()
			delete(responseChanMap, req.RequestId)
			waiting.Close()
			nodeLocker.Unlock()

			if resp == nil {
				return &pb.NodeStreamMessage{
					RequestId: req.RequestId,
					Code:      req.Code,
					Message:   "response timeout",
					IsOk:      false,
				}, nil
			}

			return resp, nil
		case <-timeout.C:
			// 从队列中删除
			nodeLocker.Lock()
			delete(responseChanMap, req.RequestId)
			waiting.Close()
			nodeLocker.Unlock()

			return &pb.NodeStreamMessage{
				RequestId: req.RequestId,
				Code:      req.Code,
				Message:   "response timeout over " + fmt.Sprintf("%d", timeoutSeconds) + " seconds",
				IsOk:      false,
			}, nil
		}
	default:
		return &pb.NodeStreamMessage{
			RequestId: req.RequestId,
			Code:      req.Code,
			Message:   "command queue is full over " + strconv.Itoa(len(requestChan)),
			IsOk:      false,
		}, nil
	}
}

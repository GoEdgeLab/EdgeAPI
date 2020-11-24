package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// 文件相关服务
type FileService struct {
	BaseService
}

// 创建文件
func (this *FileService) CreateFile(ctx context.Context, req *pb.CreateFileRequest) (*pb.CreateFileResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	fileId, err := models.SharedFileDAO.CreateFile("ipLibrary", "", req.Filename, req.Size)
	if err != nil {
		return nil, err
	}
	return &pb.CreateFileResponse{FileId: fileId}, nil
}

// 将文件置为已完成
func (this *FileService) UpdateFileFinished(ctx context.Context, req *pb.UpdateFileFinishedRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedFileDAO.UpdateFileIsFinished(req.FileId)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

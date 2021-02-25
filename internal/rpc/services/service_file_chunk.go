package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// 文件片段相关服务
type FileChunkService struct {
	BaseService
}

// 创建文件片段
func (this *FileChunkService) CreateFileChunk(ctx context.Context, req *pb.CreateFileChunkRequest) (*pb.CreateFileChunkResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	chunkId, err := models.SharedFileChunkDAO.CreateFileChunk(tx, req.FileId, req.Data)
	if err != nil {
		return nil, err
	}
	return &pb.CreateFileChunkResponse{FileChunkId: chunkId}, nil
}

// 获取的一个文件的所有片段IDs
func (this *FileChunkService) FindAllFileChunkIds(ctx context.Context, req *pb.FindAllFileChunkIdsRequest) (*pb.FindAllFileChunkIdsResponse, error) {
	// 校验请求
	_, _, err := this.ValidateAdminAndUser(ctx, 0, -1)
	if err != nil {
		return nil, err
	}

	// TODO 校验用户

	tx := this.NullTx()

	chunkIds, err := models.SharedFileChunkDAO.FindAllFileChunkIds(tx, req.FileId)
	if err != nil {
		return nil, err
	}
	return &pb.FindAllFileChunkIdsResponse{FileChunkIds: chunkIds}, nil
}

// 下载文件片段
func (this *FileChunkService) DownloadFileChunk(ctx context.Context, req *pb.DownloadFileChunkRequest) (*pb.DownloadFileChunkResponse, error) {
	// 校验请求
	_, _, err := this.ValidateAdminAndUser(ctx, 0, -1)
	if err != nil {
		return nil, err
	}

	// TODO 校验用户

	tx := this.NullTx()

	chunk, err := models.SharedFileChunkDAO.FindFileChunk(tx, req.FileChunkId)
	if err != nil {
		return nil, err
	}
	if chunk == nil {
		return &pb.DownloadFileChunkResponse{FileChunk: nil}, nil
	}
	return &pb.DownloadFileChunkResponse{FileChunk: &pb.FileChunk{Data: []byte(chunk.Data)}}, nil
}

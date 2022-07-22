package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// FileChunkService 文件片段相关服务
type FileChunkService struct {
	BaseService
}

// CreateFileChunk 创建文件片段
func (this *FileChunkService) CreateFileChunk(ctx context.Context, req *pb.CreateFileChunkRequest) (*pb.CreateFileChunkResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	chunkId, err := models.SharedFileChunkDAO.CreateFileChunk(tx, req.FileId, req.Data)
	if err != nil {
		return nil, err
	}
	return &pb.CreateFileChunkResponse{FileChunkId: chunkId}, nil
}

// FindAllFileChunkIds 获取的一个文件的所有片段IDs
func (this *FileChunkService) FindAllFileChunkIds(ctx context.Context, req *pb.FindAllFileChunkIdsRequest) (*pb.FindAllFileChunkIdsResponse, error) {
	// 校验请求
	_, _, err := this.ValidateNodeId(ctx, rpcutils.UserTypeNode, rpcutils.UserTypeAdmin, rpcutils.UserTypeUser)
	if err != nil {
		return nil, err
	}

	// TODO 校验用户

	var tx = this.NullTx()

	chunkIds, err := models.SharedFileChunkDAO.FindAllFileChunkIds(tx, req.FileId)
	if err != nil {
		return nil, err
	}
	return &pb.FindAllFileChunkIdsResponse{FileChunkIds: chunkIds}, nil
}

// DownloadFileChunk 下载文件片段
func (this *FileChunkService) DownloadFileChunk(ctx context.Context, req *pb.DownloadFileChunkRequest) (*pb.DownloadFileChunkResponse, error) {
	// 校验请求
	_, _, err := this.ValidateNodeId(ctx, rpcutils.UserTypeNode, rpcutils.UserTypeAdmin, rpcutils.UserTypeUser)
	if err != nil {
		return nil, err
	}

	// TODO 校验用户

	var tx = this.NullTx()

	chunk, err := models.SharedFileChunkDAO.FindFileChunk(tx, req.FileChunkId)
	if err != nil {
		return nil, err
	}
	if chunk == nil {
		return &pb.DownloadFileChunkResponse{FileChunk: nil}, nil
	}
	return &pb.DownloadFileChunkResponse{FileChunk: &pb.FileChunk{Data: chunk.Data}}, nil
}

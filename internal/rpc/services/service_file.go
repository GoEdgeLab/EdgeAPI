package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// FileService 文件相关服务
type FileService struct {
	BaseService
}

// FindEnabledFile 查找文件
func (this *FileService) FindEnabledFile(ctx context.Context, req *pb.FindEnabledFileRequest) (*pb.FindEnabledFileResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, false)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	file, err := models.SharedFileDAO.FindEnabledFile(tx, req.FileId)
	if err != nil {
		return nil, err
	}
	if file == nil {
		return &pb.FindEnabledFileResponse{File: nil}, nil
	}

	if !file.IsPublic {
		// 校验权限
		if userId > 0 && int64(file.UserId) != userId {
			return nil, this.PermissionError()
		}
	}

	return &pb.FindEnabledFileResponse{
		File: &pb.File{
			Id:        int64(file.Id),
			Filename:  file.Filename,
			Size:      int64(file.Size),
			CreatedAt: int64(file.CreatedAt),
			IsPublic:  file.IsPublic,
			MimeType:  file.MimeType,
			Type:      file.Type,
		},
	}, nil
}

// CreateFile 创建文件
func (this *FileService) CreateFile(ctx context.Context, req *pb.CreateFileRequest) (*pb.CreateFileResponse, error) {
	adminId, userId, err := this.ValidateAdminAndUser(ctx, false)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	fileId, err := models.SharedFileDAO.CreateFile(tx, adminId, userId, req.Type, "", req.Filename, req.Size, req.MimeType, req.IsPublic)
	if err != nil {
		return nil, err
	}
	return &pb.CreateFileResponse{FileId: fileId}, nil
}

// UpdateFileFinished 将文件置为已完成
func (this *FileService) UpdateFileFinished(ctx context.Context, req *pb.UpdateFileFinishedRequest) (*pb.RPCSuccess, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, false)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	if userId > 0 {
		err = models.SharedFileDAO.CheckUserFile(tx, userId, req.FileId)
		if err != nil {
			return nil, err
		}
	}

	err = models.SharedFileDAO.UpdateFileIsFinished(tx, req.FileId)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

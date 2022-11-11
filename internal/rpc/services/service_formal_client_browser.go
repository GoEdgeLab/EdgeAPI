// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .

package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/types"
)

// FormalClientBrowserService 浏览器信息库服务
type FormalClientBrowserService struct {
	BaseService
}

// CreateFormalClientBrowser 创建浏览器信息
func (this *FormalClientBrowserService) CreateFormalClientBrowser(ctx context.Context, req *pb.CreateFormalClientBrowserRequest) (*pb.CreateFormalClientBrowserResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	// 检查dataId是否存在
	var tx = this.NullTx()
	browser, err := models.SharedFormalClientBrowserDAO.FindBrowserWithDataId(tx, req.DataId)
	if err != nil {
		return nil, err
	}
	if browser != nil {
		return nil, errors.New("dataId '" + req.DataId + "' already exists")
	}

	browserId, err := models.SharedFormalClientBrowserDAO.CreateBrowser(tx, req.Name, req.Codes, req.DataId)
	if err != nil {
		return nil, err
	}
	return &pb.CreateFormalClientBrowserResponse{
		FormalClientBrowserId: browserId,
	}, nil
}

// CountFormalClientBrowsers 计算浏览器信息数量
func (this *FormalClientBrowserService) CountFormalClientBrowsers(ctx context.Context, req *pb.CountFormalClientBrowsersRequest) (*pb.RPCCountResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	count, err := models.SharedFormalClientBrowserDAO.CountBrowsers(tx, req.Keyword)
	if err != nil {
		return nil, err
	}

	return this.SuccessCount(count)
}

// ListFormalClientBrowsers 列出单页浏览器信息
func (this *FormalClientBrowserService) ListFormalClientBrowsers(ctx context.Context, req *pb.ListFormalClientBrowsersRequest) (*pb.ListFormalClientBrowsersResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	browsers, err := models.SharedFormalClientBrowserDAO.ListBrowsers(tx, req.Keyword, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}
	var pbBrowsers = []*pb.FormalClientBrowser{}
	for _, browser := range browsers {
		pbBrowsers = append(pbBrowsers, &pb.FormalClientBrowser{
			Id:     int64(browser.Id),
			Name:   browser.Name,
			Codes:  browser.DecodeCodes(),
			DataId: browser.DataId,
			State:  types.Int32(browser.State),
		})
	}
	return &pb.ListFormalClientBrowsersResponse{
		FormalClientBrowsers: pbBrowsers,
	}, nil
}

// UpdateFormalClientBrowser 修改浏览器信息
func (this *FormalClientBrowserService) UpdateFormalClientBrowser(ctx context.Context, req *pb.UpdateFormalClientBrowserRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	if len(req.DataId) == 0 {
		return nil, errors.New("invalid dataId")
	}

	var tx = this.NullTx()

	// 检查dataId是否已经被使用
	oldBrowser, err := models.SharedFormalClientBrowserDAO.FindBrowserWithDataId(tx, req.DataId)
	if err != nil {
		return nil, err
	}
	if oldBrowser != nil && int64(oldBrowser.Id) != req.FormalClientBrowserId {
		return nil, errors.New("the dataId '" + req.DataId + "' already has been used")
	}

	err = models.SharedFormalClientBrowserDAO.UpdateBrowser(tx, req.FormalClientBrowserId, req.Name, req.Codes, req.DataId)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// FindFormalClientBrowserWithDataId 通过dataId查询浏览器信息
func (this *FormalClientBrowserService) FindFormalClientBrowserWithDataId(ctx context.Context, req *pb.FindFormalClientBrowserWithDataIdRequest) (*pb.FindFormalClientBrowserWithDataIdResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	browser, err := models.SharedFormalClientBrowserDAO.FindBrowserWithDataId(tx, req.DataId)
	if err != nil {
		return nil, err
	}
	if browser == nil {
		return &pb.FindFormalClientBrowserWithDataIdResponse{
			FormalClientBrowser: nil,
		}, nil
	}

	return &pb.FindFormalClientBrowserWithDataIdResponse{
		FormalClientBrowser: &pb.FormalClientBrowser{
			Id:     int64(browser.Id),
			Name:   browser.Name,
			Codes:  browser.DecodeCodes(),
			DataId: browser.DataId,
			State:  types.Int32(browser.State),
		}}, nil
}

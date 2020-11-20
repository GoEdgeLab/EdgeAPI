package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/iplibrary"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// IP库服务
type IPLibraryService struct {
}

// 创建IP库
func (this *IPLibraryService) CreateIPLibrary(ctx context.Context, req *pb.CreateIPLibraryRequest) (*pb.CreateIPLibraryResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	ipLibraryId, err := models.SharedIPLibraryDAO.CreateIPLibrary(req.Type, req.FileId)
	if err != nil {
		return nil, err
	}

	return &pb.CreateIPLibraryResponse{
		IpLibraryId: ipLibraryId,
	}, nil
}

// 查找单个IP库
func (this *IPLibraryService) FindEnabledIPLibrary(ctx context.Context, req *pb.FindEnabledIPLibraryRequest) (*pb.FindEnabledIPLibraryResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	ipLibrary, err := models.SharedIPLibraryDAO.FindEnabledIPLibrary(req.IpLibraryId)
	if err != nil {
		return nil, err
	}
	if ipLibrary == nil {
		return &pb.FindEnabledIPLibraryResponse{IpLibrary: nil}, nil
	}

	// 文件相关
	var pbFile *pb.File = nil
	file, err := models.SharedFileDAO.FindEnabledFile(int64(ipLibrary.FileId))
	if err != nil {
		return nil, err
	}
	if file != nil {
		pbFile = &pb.File{
			Id:       int64(file.Id),
			Filename: file.Filename,
			Size:     int64(file.Size),
		}
	}

	return &pb.FindEnabledIPLibraryResponse{
		IpLibrary: &pb.IPLibrary{
			Id:        int64(ipLibrary.Id),
			Type:      ipLibrary.Type,
			File:      pbFile,
			CreatedAt: int64(ipLibrary.CreatedAt),
		},
	}, nil
}

// 查找最新的IP库
func (this *IPLibraryService) FindLatestIPLibraryWithType(ctx context.Context, req *pb.FindLatestIPLibraryWithTypeRequest) (*pb.FindLatestIPLibraryWithTypeResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeNode)
	if err != nil {
		return nil, err
	}

	ipLibrary, err := models.SharedIPLibraryDAO.FindLatestIPLibraryWithType(req.Type)
	if err != nil {
		return nil, err
	}
	if ipLibrary == nil {
		return &pb.FindLatestIPLibraryWithTypeResponse{IpLibrary: nil}, nil
	}

	// 文件相关
	var pbFile *pb.File = nil
	file, err := models.SharedFileDAO.FindEnabledFile(int64(ipLibrary.FileId))
	if err != nil {
		return nil, err
	}
	if file != nil {
		pbFile = &pb.File{
			Id:       int64(file.Id),
			Filename: file.Filename,
			Size:     int64(file.Size),
		}
	}

	return &pb.FindLatestIPLibraryWithTypeResponse{
		IpLibrary: &pb.IPLibrary{
			Id:        int64(ipLibrary.Id),
			Type:      ipLibrary.Type,
			File:      pbFile,
			CreatedAt: int64(ipLibrary.CreatedAt),
		},
	}, nil
}

// 列出某个类型的所有IP库
func (this *IPLibraryService) FindAllEnabledIPLibrariesWithType(ctx context.Context, req *pb.FindAllEnabledIPLibrariesWithTypeRequest) (*pb.FindAllEnabledIPLibrariesWithTypeResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	ipLibraries, err := models.SharedIPLibraryDAO.FindAllEnabledIPLibrariesWithType(req.Type)
	if err != nil {
		return nil, err
	}
	result := []*pb.IPLibrary{}
	for _, library := range ipLibraries {
		// 文件相关
		var pbFile *pb.File = nil
		file, err := models.SharedFileDAO.FindEnabledFile(int64(library.FileId))
		if err != nil {
			return nil, err
		}
		if file != nil {
			pbFile = &pb.File{
				Id:       int64(file.Id),
				Filename: file.Filename,
				Size:     int64(file.Size),
			}
		}

		result = append(result, &pb.IPLibrary{
			Id:        int64(library.Id),
			Type:      library.Type,
			File:      pbFile,
			CreatedAt: int64(library.CreatedAt),
		})
	}
	return &pb.FindAllEnabledIPLibrariesWithTypeResponse{IpLibraries: result}, nil
}

// 删除IP库
func (this *IPLibraryService) DeleteIPLibrary(ctx context.Context, req *pb.DeleteIPLibraryRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedIPLibraryDAO.DisableIPLibrary(req.IpLibraryId)
	if err != nil {
		return nil, err
	}
	return rpcutils.Success()
}

// 查询某个IP信息
func (this *IPLibraryService) LookupIPRegion(ctx context.Context, req *pb.LookupIPRegionRequest) (*pb.LookupIPRegionResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	result, err := iplibrary.SharedLibrary.Lookup(req.Ip)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return &pb.LookupIPRegionResponse{Region: nil}, nil
	}

	countryId, err := models.SharedRegionCountryDAO.FindCountryIdWithCountryName(result.Country)
	if err != nil {
		return nil, err
	}

	provinceId, err := models.SharedRegionProvinceDAO.FindProvinceIdWithProvinceName(result.Province)
	if err != nil {
		return nil, err
	}

	return &pb.LookupIPRegionResponse{Region: &pb.IPRegion{
		Country:    result.Country,
		Region:     result.Region,
		Province:   result.Province,
		City:       result.City,
		Isp:        result.ISP,
		CountryId:  countryId,
		ProvinceId: provinceId,
	}}, nil
}

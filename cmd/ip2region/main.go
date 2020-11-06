package main

import (
	"bytes"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/iwind/TeaGo/Tea"
	_ "github.com/iwind/TeaGo/bootstrap"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/lists"
	"github.com/iwind/TeaGo/logs"
	"io/ioutil"
	"os"
	"regexp"
)

func main() {
	// 导入数据
	if lists.ContainsString(os.Args, "import") {
		dbs.NotifyReady()

		data, err := ioutil.ReadFile(Tea.Root + "/resources/ipdata/ip2region/global_region.csv")
		if err != nil {
			logs.Println("[ERROR]" + err.Error())
			return
		}
		if len(data) == 0 {
			logs.Println("[ERROR]file content should not be empty")
			return
		}
		lines := bytes.Split(data, []byte{'\n'})
		for _, line := range lines {
			line = bytes.TrimSpace(line)
			if len(line) == 0 {
				continue
			}

			s := string(line)
			reg := regexp.MustCompile(`(?U)(\d+),(\d+),(.+),(\d+),`)
			if !reg.MatchString(s) {
				continue
			}
			result := reg.FindStringSubmatch(s)
			dataId := result[1]
			parentDataId := result[2]
			name := result[3]
			level := result[4]

			switch level {
			case "1": // 国家|地区
				countryId, err := models.SharedRegionCountryDAO.FindCountryIdWithDataId(dataId)
				if err != nil {
					logs.Println("[ERROR]" + err.Error())
					return
				}
				if countryId == 0 {
					logs.Println("creating country or region ", name)
					_, err = models.SharedRegionCountryDAO.CreateCountry(name, dataId)
					if err != nil {
						logs.Println("[ERROR]" + err.Error())
						return
					}
				}
			case "2": // 省份|地区
				provinceId, err := models.SharedRegionProvinceDAO.FindProvinceIdWithDataId(dataId)
				if err != nil {
					logs.Println("[ERROR]" + err.Error())
					return
				}
				if provinceId == 0 {
					logs.Println("creating province", name)

					countryId, err := models.SharedRegionCountryDAO.FindCountryIdWithDataId(parentDataId)
					if err != nil {
						logs.Println("[ERROR]" + err.Error())
						return
					}
					if countryId == 0 {
						logs.Println("[ERROR]can not find country from data id '" + parentDataId + "'")
						return
					}

					_, err = models.SharedRegionProvinceDAO.CreateProvince(countryId, name, dataId)
					if err != nil {
						logs.Println("[ERROR]" + err.Error())
						return
					}
				}
			case "3": // 城市
				cityId, err := models.SharedRegionCityDAO.FindCityWithDataId(dataId)
				if err != nil {
					logs.Println("[ERROR]" + err.Error())
					return
				}
				if cityId == 0 {
					logs.Println("creating city", name)

					provinceId, err := models.SharedRegionProvinceDAO.FindProvinceIdWithDataId(parentDataId)
					if err != nil {
						logs.Println("[ERROR]" + err.Error())
						return
					}
					_, err = models.SharedRegionCityDAO.CreateCity(provinceId, name, dataId)
					if err != nil {
						logs.Println("[ERROR]" + err.Error())
						return
					}
				}
			}
		}

		logs.Println("done")
	}
}

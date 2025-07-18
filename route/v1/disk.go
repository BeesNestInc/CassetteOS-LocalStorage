package v1

import (
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/BeesNestInc/CassetteOS-Common/model"
	"github.com/BeesNestInc/CassetteOS-Common/utils/common_err"
	"github.com/BeesNestInc/CassetteOS-Common/utils/logger"
	"github.com/BeesNestInc/CassetteOS-LocalStorage/common"
	model1 "github.com/BeesNestInc/CassetteOS-LocalStorage/model"
	"github.com/BeesNestInc/CassetteOS-LocalStorage/service"
	"github.com/labstack/echo/v4"
	"github.com/shirou/gopsutil/v3/disk"
	"go.uber.org/zap"
)

const messagePathStorageStatus = common.ServiceName + ":storage_status"

var diskMap = make(map[string]string)

type StorageMessage struct {
	Type   string `json:"type"`   // sata,usb
	Action string `json:"action"` // remove add
	Path   string `json:"path"`
	Volume string `json:"volume"`
	Size   uint64 `json:"size"`
}

// @Summary disk list
// @Produce  application/json
// @Accept application/json
// @Tags disk
// @Security ApiKeyAuth
// @Success 200 {string} string "ok"
// @Router /disk/list [get]
func GetDiskList(ctx echo.Context) error {
	blkList := service.MyService.Disk().LSBLK(false)

	dbList, err := service.MyService.Disk().GetSerialAllFromDB()
	if err != nil {
		logger.Error("error when getting all volumes from database", zap.Error(err))
		return ctx.JSON(http.StatusInternalServerError, model.Result{Success: common_err.SERVICE_ERROR, Message: err.Error()})
	}

	part := make(map[string]int64, len(dbList))
	for _, v := range dbList {
		part[v.MountPoint] = v.CreatedAt
	}

	disks := []model1.Drive{}
	avail := []model1.Drive{}

	var systemDisk *model1.LSBLKModel

	for _, currentDisk := range blkList {
		childre := []model1.DiskChildren{}
		supported := true
		if len(currentDisk.Children) > 0 {
			for _, v := range currentDisk.Children {
				if !service.IsFormatSupported(v) {
					supported = false
				}
				t := model1.DiskChildren{
					Name:      v.Name,
					Size:      v.Size,
					Format:    v.FsType,
					Supported: service.IsFormatSupported(v),
				}
				childre = append(childre, t)
			}
		} else {
			if !service.IsFormatSupported(currentDisk) {
				supported = false
			}
		}

		disk := model1.Drive{
			Serial:         currentDisk.Serial,
			Name:           currentDisk.Name,
			Size:           currentDisk.Size,
			Path:           currentDisk.Path,
			Model:          currentDisk.Model,
			ChildrenNumber: len(currentDisk.Children),
			Children:       childre,
			Supported:      supported,
		}

		if currentDisk.Rota {
			disk.DiskType = "HDD"
		} else {
			disk.DiskType = "SSD"
		}
		if currentDisk.Tran == "usb" {
			disk.DiskType = "USB"
		}

		temp := service.MyService.Disk().SmartCTL(currentDisk.Path)
		disk.Temperature = temp.Temperature.Current

		if systemDisk == nil {
			// go 5 level deep to look for system block device by mount point being "/"
			systemDisk := service.WalkDisk(currentDisk, 5, func(blk model1.LSBLKModel) bool { return blk.MountPoint == "/" })

			if systemDisk != nil {
				disk.Model = "System"
				if strings.Contains(systemDisk.SubSystems, "mmc") {
					disk.DiskType = "MMC"
				} else if strings.Contains(systemDisk.SubSystems, "usb") {
					disk.DiskType = "USB"
				}
				disk.Health = "true"

				disks = append(disks, disk)
				continue
			}
		}

		if !service.IsDiskSupported(currentDisk) {
			continue
		}

		if reflect.DeepEqual(temp, model1.SmartctlA{}) {
			temp.SmartStatus.Passed = true
		}

		isAvail := true
		if len(currentDisk.MountPoint) != 0 {
			isAvail = false
		} else {
			for _, v := range currentDisk.Children {
				if v.MountPoint != "" {
					isAvail = false
				}
			}
		}

		if isAvail {
			disk.NeedFormat = false
			avail = append(avail, disk)
		}

		disk.Health = strconv.FormatBool(temp.SmartStatus.Passed)

		disks = append(disks, disk)
	}

	data := map[string]interface{}{
		"disks": disks,
		"avail": avail,
	}

	return ctx.JSON(common_err.SUCCESS, model.Result{Success: common_err.SUCCESS, Message: common_err.GetMsg(common_err.SUCCESS), Data: data})
}

// @Summary disk list
// @Produce  application/json
// @Accept application/json
// @Tags disk
// @Security ApiKeyAuth
// @Success 200 {string} string "ok"
// @Router /disk/list [get]

func DeleteDisksUmount(ctx echo.Context) error {
	js := make(map[string]string)
	if err := ctx.Bind(&js); err != nil {
		return ctx.JSON(http.StatusBadRequest, model.Result{Success: common_err.INVALID_PARAMS, Message: common_err.GetMsg(common_err.INVALID_PARAMS), Data: err.Error()})
	}

	// requires password from user to confirm the action
	// if claims, err := jwt.ParseToken(c.GetHeader("Authorization"), false); err != nil || encryption.GetMD5ByStr(js["password"]) != claims.Password {
	// 	return ctx.JSON(http.StatusUnauthorized, model.Result{Success: common_err.PWD_INVALID, Message: common_err.GetMsg(common_err.PWD_INVALID)})
	// 	return
	// }

	path := js["path"]

	if len(path) == 0 {
		return ctx.JSON(common_err.CLIENT_ERROR, model.Result{Success: common_err.INVALID_PARAMS, Message: common_err.GetMsg(common_err.INVALID_PARAMS)})
	}

	if _, ok := diskMap[path]; ok {
		return ctx.JSON(common_err.SERVICE_ERROR, model.Result{Success: common_err.DISK_BUSYING, Message: common_err.GetMsg(common_err.DISK_BUSYING)})
	}

	diskInfo := service.MyService.Disk().GetDiskInfo(path)
	if len(diskInfo.Children) == 0 && service.IsDiskSupported(diskInfo) {
		t := diskInfo
		t.Children = nil
		diskInfo.Children = append(diskInfo.Children, t)
	}
	for _, v := range diskInfo.Children {
		if err := service.MyService.Disk().UmountPointAndRemoveDir(v); err != nil {
			return ctx.JSON(http.StatusInternalServerError, model.Result{Success: common_err.REMOVE_MOUNT_POINT_ERROR, Message: err.Error()})
		}

		// delete data
		if err := service.MyService.Disk().DeleteMountPointFromDB(v.Path, v.MountPoint); err != nil {
			logger.Error("error when deleting mount point from database", zap.Error(err), zap.String("path", v.Path), zap.String("mount point", v.MountPoint))
		}

		if err := service.MyService.Shares().DeleteShare(v.MountPoint); err != nil {
			logger.Error("error when deleting share by mount point", zap.Error(err), zap.String("mount point", v.MountPoint))
		}
	}

	service.MyService.Disk().RemoveLSBLKCache()

	// send notify to client
	go func() {
		message := map[string]interface{}{
			"data": StorageMessage{
				Action: "REMOVED",
				Path:   path,
				Volume: "",
				Size:   0,
				Type:   "",
			},
		}

		if err := service.MyService.Notify().SendNotify(messagePathStorageStatus, message); err != nil {
			logger.Error("error when sending notification", zap.Error(err), zap.String("message path", messagePathStorageStatus), zap.Any("message", message))
		}
	}()

	return ctx.JSON(common_err.SUCCESS, model.Result{Success: common_err.SUCCESS, Message: common_err.GetMsg(common_err.SUCCESS), Data: path})
}

func GetDiskSize(ctx echo.Context) error {
	path := ctx.QueryParam("path")
	if len(path) == 0 {
		return ctx.JSON(common_err.CLIENT_ERROR, model.Result{Success: common_err.INVALID_PARAMS, Message: common_err.GetMsg(common_err.INVALID_PARAMS)})
	}
	p, err := disk.Usage(path)
	if err != nil {
		return ctx.JSON(common_err.SERVICE_ERROR, model.Result{Success: common_err.SERVICE_ERROR, Message: err.Error()})
	}
	data := map[string]interface{}{
		"path": path,
		"free": p.Free,
		"used": p.Used,
	}
	return ctx.JSON(common_err.SUCCESS, model.Result{Success: common_err.SUCCESS, Message: common_err.GetMsg(common_err.SUCCESS), Data: data})
}

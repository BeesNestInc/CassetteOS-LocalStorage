package v1

import (
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/BeesNestInc/CassetteOS-Common/model"
	"github.com/BeesNestInc/CassetteOS-Common/utils/common_err"
	"github.com/BeesNestInc/CassetteOS-Common/utils/file"
	"github.com/BeesNestInc/CassetteOS-Common/utils/logger"
	model1 "github.com/BeesNestInc/CassetteOS-LocalStorage/model"
	"github.com/BeesNestInc/CassetteOS-LocalStorage/pkg/config"
	"github.com/BeesNestInc/CassetteOS-LocalStorage/service"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

const messagePathSysUSB = "sys_usb"

// @Summary Turn off usb auto-mount
// @Produce  application/json
// @Accept application/json
// @Tags sys
// @Security ApiKeyAuth
// @Success 200 {string} string "ok"
// @Router /sys/usb/off [put]
func PutSystemUSBAutoMount(ctx echo.Context) error {
	js := make(map[string]string)
	if err := ctx.Bind(&js); err != nil {
		return ctx.JSON(http.StatusBadRequest, model.Result{Success: common_err.INVALID_PARAMS, Message: err.Error()})
	}

	status := js["state"]
	if status == "on" {
		service.MyService.USB().UpdateUSBAutoMount("True")
		service.MyService.USB().ExecUSBAutoMountShell("True")
	} else {
		service.MyService.USB().UpdateUSBAutoMount("False")
		service.MyService.USB().ExecUSBAutoMountShell("False")
	}

	go func() {
		message := map[string]interface{}{
			"data": service.MyService.Disk().GetUSBDriveStatusList(),
		}

		if err := service.MyService.Notify().SendNotify(messagePathSysUSB, message); err != nil {
			logger.Error("failed to send notify", zap.Any("message", message), zap.Error(err))
		}
	}()

	return ctx.JSON(common_err.SUCCESS,
		model.Result{
			Success: common_err.SUCCESS,
			Message: common_err.GetMsg(common_err.SUCCESS),
		})
}

// @Summary Turn off usb auto-mount
// @Produce  application/json
// @Accept application/json
// @Tags sys
// @Security ApiKeyAuth
// @Success 200 {string} string "ok"
// @Router /sys/usb [get]
func GetSystemUSBAutoMount(ctx echo.Context) error {
	state := "True"
	if strings.ToLower(config.ServerInfo.USBAutoMount) != "true" {
		state = "False"
	}

	go func() {
		message := map[string]interface{}{
			"data": service.MyService.Disk().GetUSBDriveStatusList(),
		}

		if err := service.MyService.Notify().SendNotify(messagePathSysUSB, message); err != nil {
			logger.Error("failed to send notify", zap.Any("message", message), zap.Error(err))
		}
	}()

	return ctx.JSON(common_err.SUCCESS,
		model.Result{
			Success: common_err.SUCCESS,
			Message: common_err.GetMsg(common_err.SUCCESS),
			Data:    state,
		})
}

func GetDisksUSBList(ctx echo.Context) error {
	list := service.MyService.Disk().LSBLK(false)
	data := []model1.USBDriveStatus{}
	for _, v := range list {
		if v.Tran == "usb" {
			temp := model1.USBDriveStatus{}
			temp.Model = v.Model
			temp.Name = v.Label
			if temp.Name == "" {
				temp.Name = v.Name
			}
			temp.Size = v.Size
			children := []model1.USBChildren{}
			for _, child := range v.Children {
				if len(child.MountPoint) > 0 {
					tempChildren := model1.USBChildren{}
					tempChildren.MountPoint = child.MountPoint
					tempChildren.Size, _ = strconv.ParseUint(child.FSSize.String(), 10, 64)
					tempChildren.Avail, _ = strconv.ParseUint(child.FSAvail.String(), 10, 64)
					tempChildren.Name = child.Label
					if len(tempChildren.Name) == 0 {
						tempChildren.Name = filepath.Base(child.MountPoint)
					}
					avail, _ := strconv.ParseUint(child.FSAvail.String(), 10, 64)
					children = append(children, tempChildren)
					temp.Avail += avail
				}
			}

			temp.Children = children
			data = append(data, temp)
		}
	}
	return ctx.JSON(common_err.SUCCESS, model.Result{Success: common_err.SUCCESS, Message: common_err.GetMsg(common_err.SUCCESS), Data: data})
}

func DeleteDiskUSB(ctx echo.Context) error {
	js := make(map[string]string)
	if err := ctx.Bind(&js); err != nil {
		return ctx.JSON(http.StatusBadRequest, model.Result{Success: common_err.INVALID_PARAMS, Message: common_err.GetMsg(common_err.INVALID_PARAMS), Data: err.Error()})
	}
	mountPoint := js["mount_point"]
	if file.CheckNotExist(mountPoint) {
		return ctx.JSON(http.StatusBadRequest, model.Result{Success: common_err.DIR_NOT_EXISTS, Message: common_err.GetMsg(common_err.DIR_NOT_EXISTS)})
	}

	if err := service.MyService.Disk().UmountUSB(mountPoint); err != nil {
		return ctx.JSON(http.StatusInternalServerError, model.Result{Success: common_err.SERVICE_ERROR, Message: err.Error()})
	}

	return ctx.JSON(common_err.SUCCESS, model.Result{Success: common_err.SUCCESS, Message: common_err.GetMsg(common_err.SUCCESS), Data: mountPoint})
}

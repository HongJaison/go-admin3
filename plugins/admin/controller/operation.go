package controller

import (
	"github.com/HongJaison/go-admin3/context"
	"github.com/HongJaison/go-admin3/modules/config"
	"github.com/HongJaison/go-admin3/plugins/admin/modules/constant"
	"github.com/HongJaison/go-admin3/plugins/admin/modules/response"
)

func (h *Handler) Operation(ctx *context.Context) {
	id := ctx.Query("__goadmin_op_id")
	if !h.OperationHandler(config.Url("/operation/"+id), ctx) {
		errMsg := "not found"
		if ctx.Headers(constant.PjaxHeader) == "" && ctx.Method() != "GET" {
			response.BadRequest(ctx, errMsg)
		} else {
			response.Alert(ctx, errMsg, errMsg, errMsg, h.conn, h.navButtons)
		}
		return
	}
}

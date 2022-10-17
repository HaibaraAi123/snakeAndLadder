package http

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoLog "github.com/labstack/gommon/log"
	"golang.org/x/net/websocket"
	"net/http"
	"snakeAndLadder/edge/ws"
	"snakeAndLadder/utl"
	"strings"
	"time"
)

func InitRouter() *echo.Echo {
	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			logger := utl.GetLogger("", "/log/backend.log", echoLog.DEBUG)
			c.SetLogger(logger)
			return next(c)
		}
	})
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: `${time_rfc3339}  reqId:${id}  remote_ip:${remote_ip}  ` +
			`host:${host}  method:${method}  uri:${uri}  user_agent:${user_agent}  ` +
			`status:${status}  error:${error}  latency:${latency_human}  ` +
			`bytes_in:${bytes_in}  bytes_out:${bytes_out}` + "\n",
		CustomTimeFormat: "",
	}))
	e.Use(middleware.Recover())
	//e.File()
	/*
		base
	*/
	pathGroup := e.Group("/backend")
	baseGroup := pathGroup.Group("")
	baseGroup.POST("/api", func(ctx echo.Context) error {
		fn, logger := "api", ctx.Logger()
		msg := &ws.Msg{}
		err := json.NewDecoder(ctx.Request().Body).Decode(msg)
		if err != nil {
			logger.Warnf("%v, decode post body fail, err:%v", fn, err)
			return err
		}
		if msg.Type != ws.TypeMsgRequest || len(msg.Data) == 0 || len(msg.Method) == 0 {
			logger.Warnf("%v, invalid msg:%v", fn, msg)
			return errors.New("invalid msg")
		}
		method := string(msg.Method)
		l := strings.Split(method, "/")
		if len(l) != 3 {
			logger.Printf("%v, msg invalid:%v", fn, msg)
			return errors.New("invalid msg.Method")
		}
		service := l[1]
		dialer := utl.Dialer{}
		conn, err := dialer.Dial(service)
		if err != nil {
			logger.Printf("%v, dial fail, err:%v, service:%v", fn, err, service)
			return err
		}
		dialCtx, cancel := context.WithTimeout(context.TODO(), time.Second*5)
		defer cancel()
		var resp []byte
		err = conn.Invoke(dialCtx, method, msg.Data, &resp)
		if err != nil {
			logger.Printf("%v, invoke %v fail, err:%v", fn, method, err)
			return err
		}
		return ctx.JSON(http.StatusOK, &ws.Msg{
			Type:       ws.TypeMsgResponse,
			Seq:        msg.Seq,
			IsNeedResp: false,
			State:      0,
			Method:     msg.Method,
			Data:       resp,
		})
	})

	/*
		init wss
	*/
	wsGroup := pathGroup.Group("/ws")
	wsGroup.Any("", echo.WrapHandler(websocket.Handler(ws.WebsocketHandler)))

	return e
}
